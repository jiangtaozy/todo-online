package main

import (
  "net/http"
  "fmt"
  "log"
  "time"
  "encoding/json"
  "reflect"
  "github.com/justinas/alice"
  "github.com/gorilla/context"
  "github.com/julienschmidt/httprouter"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
)

func main() {
  session, err := mgo.Dial("localhost")
  if err != nil {
    panic(err)
  }
  defer session.Close()
  session.SetMode(mgo.Monotonic, true)
  appC := appContext{session.DB("test")}
  commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler)
  router := NewRouter()
  router.Get("/", commonHandlers.ThenFunc(indexHandler))
  router.Get("/about", commonHandlers.ThenFunc(aboutHandler))
  router.Get("/admin", commonHandlers.Append(authHandler).ThenFunc(adminHandler))
  router.Get("/teas/:id", commonHandlers.ThenFunc(appC.teaHandler))
  router.Post("/teas", commonHandlers.Append(bodyParserHandler(TeaResource{})).ThenFunc(appC.createTeaHandler))
  http.ListenAndServe(":8080", router)
}

type appContext struct {
  db *mgo.Database
}

type Tea struct {
  Id bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
  Name string `json:"name"`
  Category string `json:"category"`
}

type TeaResource struct {
  Data Tea `json:"data"`
}

type TeaRepo struct {
  coll *mgo.Collection
}

type router struct {
  *httprouter.Router
}

func (r *router) Get(path string, handler http.Handler) {
  r.GET(path, wrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
  r.POST(path, wrapHandler(handler))
}

func NewRouter() *router{
  return &router{httprouter.New()}
}

func (r *TeaRepo) Find(id string) (TeaResource, error) {
  result := TeaResource{}
  err := r.coll.FindId(bson.ObjectIdHex(id)).One(&result.Data)
  if err != nil {
    return result, err
  }
  return result, nil
}

func (r *TeaRepo) Create(tea *Tea) error {
  id := bson.NewObjectId()
  _, err := r.coll.UpsertId(id, tea)
  if err != nil {
    return err
  }
  tea.Id = id
  return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Welcome!")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "You are on the about page.")
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
  user := context.Get(r, "user")
  json.NewEncoder(w).Encode(user)
}

func (c *appContext) teaHandler(w http.ResponseWriter, r *http.Request) {
  params := context.Get(r, "params").(httprouter.Params)
  repo := TeaRepo{c.db.C("teas")}
  tea, err := repo.Find(params.ByName("id"))
  if err != nil {
    panic(err)
  }
  w.Header().Set("Content-Type", "application/vnd.api+json")
  json.NewEncoder(w).Encode(tea)
}

func (c *appContext) createTeaHandler(w http.ResponseWriter, r *http.Request) {
  body := context.Get(r, "body").(*TeaResource)
  repo := TeaRepo{c.db.C("teas")}
  err := repo.Create(&body.Data)
  if err != nil {
    panic(err)
  }
  w.Header().Set("Content-Type", "application/vnd.api+json")
  w.WriteHeader(201)
  json.NewEncoder(w).Encode(body)
}

func bodyParserHandler(v interface{}) func(http.Handler) http.Handler {
  t := reflect.TypeOf(v)
  m := func(next http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
      val := reflect.New(t).Interface()
      err := json.NewDecoder(r.Body).Decode(val)
      if err != nil {
        WriteError(w, ErrBadRequest)
        return
      }
      context.Set(r, "body", val)
      next.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fn)
  }
  return m
}

func loggingHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    t1 := time.Now()
    next.ServeHTTP(w, r)
    t2 := time.Now()
    log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
  }
  return http.HandlerFunc(fn)
}

func recoverHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      if err := recover(); err != nil {
        log.Printf("panic: %+v", err)
        WriteError(w, ErrInternalServer)
      }
    }()
    next.ServeHTTP(w, r)
  }
  return http.HandlerFunc(fn)
}

func authHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    user := "jemo"
    context.Set(r, "user", user)
    next.ServeHTTP(w, r)
  }
  return http.HandlerFunc(fn)
}

func wrapHandler(h http.Handler) httprouter.Handle {
  return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    context.Set(r, "params", ps)
    h.ServeHTTP(w, r)
  }
}

// Errors
type Errors struct {
  Errors []*Error `json:"errors"`
}

type Error struct {
  Id string `json:"id"`
  Status int `json:"status"`
  Title string `json:"title"`
  Detail string `json:"detail"`
}

func WriteError(w http.ResponseWriter, err *Error) {
  w.Header().Set("Content-Type", "application/vnd.api+json")
  w.WriteHeader(err.Status)
  json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}

var (
  ErrBadRequest = &Error{"bad_request", 400, "Bad request", "Request body is not well-formed. It must be JSON."}
  ErrNotAcceptable = &Error{"not_acceptalbe", 406, "Not Acceptable", "Accept header must be set to 'application/vnd.api+json'."}
  ErrUnsupportedMediaType = &Error{"unsupported_media_type", 415, "Unsupported Media Type", "Content-Type header must be set to: 'application/vnd.api+json'."}
  ErrInternalServer = &Error{"internal_server_error", 500, "Internal Server Error", "Something went wrong."}
)
