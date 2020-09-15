package main

import (
  "fmt"
  "log"
  "database/sql"
  "github.com/gorilla/mux"
  _ "github.com/lib/pq"
  "net/http"
  "strconv"
  "encoding/json"
)

type App struct {
  Router *mux.Router
  DB     *sql.DB
}

func (a *App) Run(addr string) {
  log.Fatal(http.ListenAndServe(":8010", a.Router))
}

func (a *App) Initialize(user, password, dbname string) {
  connectionString :=
  fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

  var err error
  a.DB, err = sql.Open("postgres", connectionString)
  if err != nil {
    log.Fatal(err)
  }

  a.Router = mux.NewRouter()

  a.initializeRoutes()
}

func (a *App) initializeRoutes() {
  a.Router.HandleFunc("/people", a.getPeople).Methods("GET")
  a.Router.HandleFunc("/person", a.createPerson).Methods("POST")
  a.Router.HandleFunc("/person/{id:[0-9]+}", a.getPerson).Methods("GET")
  a.Router.HandleFunc("/person/{id:[0-9]+}", a.updatePerson).Methods("PUT")
  a.Router.HandleFunc("/person/{id:[0-9]+}", a.deletePerson).Methods("DELETE")
}

func (a *App) getPeople(w http.ResponseWriter, r *http.Request) {
  count, _ := strconv.Atoi(r.FormValue("count"))
  start, _ := strconv.Atoi(r.FormValue("start"))

  if count > 10 || count < 1 {
    count = 10
  }
  if start < 0 {
    start = 0
  }

  people, err := getPeople(a.DB, start, count)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, people)
}

func (a *App) createPerson(w http.ResponseWriter, r *http.Request) {
  var p person
  decoder := json.NewDecoder(r.Body)
  if err := decoder.Decode(&p); err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid request payload")
    return
  }
  defer r.Body.Close()

  if err := p.createPerson(a.DB); err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) getPerson(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid person ID")
    return
  }

  p := person{ID: id}
  if err := p.getPerson(a.DB); err != nil {
    switch err {
    case sql.ErrNoRows:
      respondWithError(w, http.StatusNotFound, "Person not found")
    default:
      respondWithError(w, http.StatusInternalServerError, err.Error())
    }
    return
  }

  respondWithJSON(w, http.StatusOK, p)
}

func (a *App) updatePerson(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid person ID")
    return
  }

  var p person
  decoder := json.NewDecoder(r.Body)
  if err := decoder.Decode(&p); err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
    return
  }
  defer r.Body.Close()
  p.ID = id

  if err := p.updatePerson(a.DB); err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deletePerson(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, err := strconv.Atoi(vars["id"])
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid Person ID")
    return
  }

  p := person{ID: id}
  if err := p.deletePerson(a.DB); err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
  respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
  response, _ := json.Marshal(payload)

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(code)
  w.Write(response)
}
