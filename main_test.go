package main

import (
  "os"
  "testing"
  "log"
  "net/http"
  "net/http/httptest"
  "encoding/json"
  "bytes"
  "strconv"
)

var a App

func TestMain(m *testing.M) {
  a.Initialize(
    os.Getenv("APP_DB_USERNAME"),
    os.Getenv("APP_DB_PASSWORD"),
    os.Getenv("APP_DB_NAME"))

    ensureTableExists()
    code := m.Run()
    clearTable()
    os.Exit(code)
  }

func TestEmptyTable(t *testing.T) {
  clearTable()

  req, _ := http.NewRequest("GET", "/people", nil)
  response := executeRequest(req)

  checkResponseCode(t, http.StatusOK, response.Code)

  if body := response.Body.String(); body != "[]" {
    t.Errorf("Expected an empty array. Got %s", body)
  }
}

func TestGetNonExistentPerson(t *testing.T) {
  clearTable()

  req, _ := http.NewRequest("GET", "/person/1", nil)
  response := executeRequest(req)

  checkResponseCode(t, http.StatusNotFound, response.Code)

  var m map[string]string
  json.Unmarshal(response.Body.Bytes(), &m)
  if m["error"] != "Person not found" {
    t.Errorf("Expected the 'error' key of the response to be set to 'Person not found'. Got '%s'", m["error"])
  }
}

func TestCreatePerson(t *testing.T) {
  clearTable()

  var jsonStr = []byte(`{"name": "test", "age": 123, "gender": "male"}`)
  req, _ := http.NewRequest("POST", "/person", bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")

  response := executeRequest(req)
  checkResponseCode(t, http.StatusCreated, response.Code)

  var m map[string]interface{}
  json.Unmarshal(response.Body.Bytes(), &m)

  if m["name"] != "test" {
    t.Errorf("Expected person name to be 'test'. Got '%v'", m["name"])
  }

  if m["gender"] != "male" {
    t.Errorf("Expected person gender to be 'male'. Got '%v'", m["gender"])
  }

  if m["age"] != 123.0 {
    t.Errorf("Expected person age to be '123'. Got '%v'", m["age"])
  }

  if m["id"] != 1.0 {
    t.Errorf("Expected person ID to be '1'. Got '%v'", m["id"])
  }
}

func TestGetPeople(t *testing.T) {
  clearTable()
  addPeople(1)

  req, _ := http.NewRequest("GET", "/person/1", nil)
  response := executeRequest(req)

  checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdatePerson(t *testing.T) {
  clearTable()
  addPeople(1)

  req, _ := http.NewRequest("GET", "/person/1", nil)
  response := executeRequest(req)
  var originalPerson map[string]interface{}
  json.Unmarshal(response.Body.Bytes(), &originalPerson)

  var jsonStr = []byte(`{"name":"test", "age": 123}`)
  req, _ = http.NewRequest("PUT", "/person/1", bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")

  response = executeRequest(req)

  checkResponseCode(t, http.StatusOK, response.Code)

  var m map[string]interface{}
  json.Unmarshal(response.Body.Bytes(), &m)

  if m["id"] != originalPerson["id"] {
    t.Errorf("Expected the id to remain the same (%v). Got %v", originalPerson["id"], m["id"])
  }

  if m["name"] == originalPerson["name"] {
    t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalPerson["name"], m["name"], m["name"])
  }

  if m["age"] == originalPerson["age"] {
    t.Errorf("Expected the age to change from '%v' to '%v'. Got '%v'", originalPerson["age"], m["age"], m["age"])
  }
}

func TestDeletePerson(t *testing.T) {
  clearTable()
  addPeople(1)

  req, _ := http.NewRequest("GET", "/person/1", nil)
  response := executeRequest(req)
  checkResponseCode(t, http.StatusOK, response.Code)

  req, _ = http.NewRequest("DELETE", "/person/1", nil)
  response = executeRequest(req)

  checkResponseCode(t, http.StatusOK, response.Code)

  req, _ = http.NewRequest("GET", "/person/1", nil)
  response = executeRequest(req)
  checkResponseCode(t, http.StatusNotFound, response.Code)
}

func ensureTableExists() {
  if _, err := a.DB.Exec(tableCreationQuery); err != nil {
    log.Fatal(err)
  }
}

func clearTable() {
  a.DB.Exec("DELETE FROM people")
  a.DB.Exec("ALTER SEQUENCE people_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS people
(
    id SERIAL,
    name TEXT NOT NULL,
    age SMALLINT NOT NULL DEFAULT 0,
    gender TEXT,
    CONSTRAINT people_pkey PRIMARY KEY (id)
)`

func addPeople(count int) {
  if count < 1 {
    count = 1
  }

  for i := 0; i < count; i++ {
    a.DB.Exec("INSERT INTO people(name, age, gender) VALUES($1, $2, $3)", "Person "+strconv.Itoa(i), (i + 10), "male")
  }
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
  rr := httptest.NewRecorder()
  a.Router.ServeHTTP(rr, req)

  return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
  if expected != actual {
    t.Errorf("Expected response code %d. Got %d\n", expected, actual)
  }
}
