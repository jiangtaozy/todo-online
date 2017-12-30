package main

import (
  "time"
  "log"
  "reflect"
  "net/http"
  "encoding/json"

  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "github.com/gorilla/context"
  "github.com/julienschmidt/httprouter"
  "github.com/justinas/alice"
  "github.com/rs/cors"
)

// Repo
type Todo struct {
  Id bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
  // You can use tags on struct field declarations
  // to customize the encoded JSON key names.
  Text string `json:"text"`
  Completed bool `json:"completed"`
  CreateAt time.Time `json:"create_at"`
}

type TodoCollection struct {
  Data []Todo `json:"data"`
}

type TodoResource struct {
  Data Todo `json:"data"`
}

type TodoRepository struct {
  Collection *mgo.Collection
}

func (repo *TodoRepository) All() (TodoCollection, error) {
  result := TodoCollection{[]Todo{}}
  err := repo.Collection.Find(nil).All(&result.Data)
  return result, err
}

func (repo *TodoRepository) Find(id string) (TodoResource, error) {
  result := TodoResource{}
  err := repo.Collection.FindId(bson.ObjectIdHex(id)).One(&result.Data)
  return result, err
}

func (repo *TodoRepository) Create(todo *Todo) error {
  id := bson.NewObjectId()
  _, err := repo.Collection.UpsertId(id, todo)
  if err != nil {
    return err
  }
  todo.Id = id
  return nil
}

func (repo *TodoRepository) Update(todo *Todo) error {
  err := repo.Collection.UpdateId(todo.Id, todo)
  return err
}

func (repo *TodoRepository) Delete(id string) error {
  err := repo.Collection.RemoveId(bson.ObjectIdHex(id))
  return err
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
    ErrBadRequest           = &Error{"bad_request", 400, "Bad request", "Request body is not well-formed. It must be JSON."}
    ErrNotAcceptable        = &Error{"not_acceptable", 406, "Not Acceptable", "Accept header must be set to 'application/vnd.api+json'."}
    ErrUnsupportedMediaType = &Error{"unsupported_media_type", 415, "Unsupported Media Type", "Content-Type header must be set to: 'application/vnd.api+json'."}
    ErrInternalServer       = &Error{"internal_server_error", 500, "Internal Server Error", "Something went wrong."}
)

// Middlewares

func recoverHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      if err := recover(); err != nil {
        log.Printf("panic: %v", err)
        WriteError(w, ErrInternalServer)
      }
    }()
    next.ServeHTTP(w, r)
  }
  return http.HandlerFunc(fn)
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

func acceptHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    log.Printf("Accept: %s\n", r.Header.Get("Accept"))
    accept := r.Header.Get("Accept")
    if accept != "application/vnd.api+json" && accept != "*/*" {
      WriteError(w, ErrNotAcceptable)
      return
    }
    next.ServeHTTP(w, r)
  }
  return http.HandlerFunc(fn)
}

func contentTypeHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    log.Printf("Content-Type: %s\n", r.Header.Get("Content-Type"))
    contentType := r.Header.Get("Content-Type")
    if contentType != "application/vnd.api+json" && contentType != "application/json" {
      WriteError(w, ErrUnsupportedMediaType)
      return
    }
    next.ServeHTTP(w, r)
  }
  return http.HandlerFunc(fn)
}

func bodyHandler(v interface{}) func(http.Handler) http.Handler {
  t := reflect.TypeOf(v)
  m := func(next http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
      val := reflect.New(t).Interface()
      err := json.NewDecoder(r.Body).Decode(val)
      if err != nil {
        WriteError(w, ErrBadRequest)
        return
      }
      if next != nil {
        context.Set(r, "body", val)
        next.ServeHTTP(w, r)
      }
    }
    return http.HandlerFunc(fn)
  }
  return m
}

// Main handlers

type appContext struct {
  db *mgo.Database
}

func (c *appContext) todosHandler(w http.ResponseWriter, r *http.Request) {
  repository := TodoRepository{c.db.C("todo")}
  todos, err := repository.All()
  if err != nil {
    panic(err)
  }
  w.Header().Set("Content-Type", "application/vnd.api+json")
  json.NewEncoder(w).Encode(todos)
}

func (c *appContext) todoHandler(w http.ResponseWriter, r *http.Request) {
  params := context.Get(r, "params").(httprouter.Params)
  repository := TodoRepository{c.db.C("todo")}
  todo, err := repository.Find(params.ByName("id"))
  if err != nil {
    panic(err)
  }
  w.Header().Set("Content-Type", "application/vnd.api+json")
  json.NewEncoder(w).Encode(todo)
}

func (c *appContext) createTodoHandler(w http.ResponseWriter, r *http.Request) {
  body := context.Get(r, "body").(*TodoResource)
  body.Data.CreateAt = time.Now()
  repository := TodoRepository{c.db.C("todo")}
  err := repository.Create(&body.Data)
  if err != nil {
    panic(err)
  }
  w.Header().Set("Content-Type", "application/vnd.api+json")
  w.WriteHeader(201)
  json.NewEncoder(w).Encode(body)
}

func (c *appContext) updateTodoHandler(w http.ResponseWriter, r *http.Request) {
  params := context.Get(r, "params").(httprouter.Params)
  body := context.Get(r, "body").(*TodoResource)
  body.Data.Id = bson.ObjectIdHex(params.ByName("id"))
  repository := TodoRepository{c.db.C("todo")}
  err := repository.Update(&body.Data)
  if err != nil {
    panic(err)
  }
  w.WriteHeader(204)
  w.Write([]byte("\n"))
}

func (c *appContext) deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
  params := context.Get(r, "params").(httprouter.Params)
  repository := TodoRepository{c.db.C("todo")}
  err := repository.Delete(params.ByName("id"))
  if err != nil {
    panic(err)
  }
  w.WriteHeader(204)
  w.Write([]byte("\n"))
}

// Router

type router struct {
  *httprouter.Router
}

func wrapHandler(h http.Handler) httprouter.Handle {
  return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    context.Set(r, "params", ps)
    h.ServeHTTP(w, r)
  }
}

func (r *router) Get(path string, handler http.Handler) {
  r.GET(path, wrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
  r.POST(path, wrapHandler(handler))
}

func (r *router) Put(path string, handler http.Handler) {
  r.PUT(path, wrapHandler(handler))
}

func (r *router) Delete(path string, handler http.Handler) {
  r.DELETE(path, wrapHandler(handler))
}

func NewRouter() *router {
  return &router{httprouter.New()}
}

func main() {
  session, err := mgo.Dial("localhost")
  if err != nil {
    panic(err)
  }
  defer session.Close()
  session.SetMode(mgo.Monotonic, true)
  appC := appContext{session.DB("todoonline")}
  // cors
  //c := cors.AllowAll()
  c := cors.New(cors.Options{
    AllowedOrigins: []string{"*"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    Debug: true,
  })
  commonHandlers := alice.New(c.Handler, context.ClearHandler, loggingHandler, recoverHandler)
  router := NewRouter()
  router.Get("/todos/:id", commonHandlers.ThenFunc(appC.todoHandler))
  router.Put("/todos/:id", commonHandlers.Append(bodyHandler(TodoResource{})).ThenFunc(appC.updateTodoHandler))
  router.Delete("/todos/:id", commonHandlers.ThenFunc(appC.deleteTodoHandler))
  router.Get("/todos", commonHandlers.ThenFunc(appC.todosHandler))
  router.Post("/todos", commonHandlers.Append(bodyHandler(TodoResource{})).ThenFunc(appC.createTodoHandler))
  router.NotFound = http.FileServer(http.Dir("client/build"))
  log.Println("listen at 2048")
  http.ListenAndServe(":2048", router)
}
