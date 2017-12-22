package main

import (
  "net/http"
  "fmt"
  "log"
  "time"
  "encoding/json"
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
  commonHandlers := alice.New(context.ClearHandler, loggingHandler)
  router := NewRouter()
  router.Get("/", commonHandlers.ThenFunc(indexHandler))
  router.Get("/about", commonHandlers.ThenFunc(aboutHandler))
  router.Get("/admin", commonHandlers.Append(authHandler).ThenFunc(adminHandler))
  router.Get("/teas/:id", commonHandlers.ThenFunc(appC.teaHandler))
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
  repo := TeaRepo{c.db.C("test")}
  tea, err := repo.Find(params.ByName("id"))
  if err != nil {
    panic(err)
  }
  w.Header().Set("Content-Type", "application/vnd.api+json")
  json.NewEncoder(w).Encode(tea)
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
