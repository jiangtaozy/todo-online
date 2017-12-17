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
)

func main() {
  commonHandlers := alice.New(context.ClearHandler, loggingHandler)
  router := httprouter.New()
  router.GET("/", wrapHandler(commonHandlers.ThenFunc(indexHandler)))
  router.GET("/about", wrapHandler(commonHandlers.ThenFunc(aboutHandler)))
  router.GET("/admin", wrapHandler(commonHandlers.Append(authHandler).ThenFunc(adminHandler)))
  router.GET("/teas/:id", wrapHandler(commonHandlers.ThenFunc(teaHandler)))
  http.ListenAndServe(":8080", router)
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

func teaHandler(w http.ResponseWriter, r *http.Request) {
  params := context.Get(r, "params").(httprouter.Params)
  id := params.ByName("id")
  json.NewEncoder(w).Encode(id)
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
