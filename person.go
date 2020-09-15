package main

import (
  "database/sql"
)

type person struct {
  ID     int     `json:"id"`
  Name   string  `json:"name"`
  Age    int     `json:"age"`
  Gender string  `json:"gender"`
}

func (p *person) getPerson(db *sql.DB) error {
  return db.QueryRow("SELECT name, age, gender FROM people WHERE id=$1",
  p.ID).Scan(&p.Name, &p.Age, &p.Gender)
}

func (p *person) createPerson(db *sql.DB) error {
  err := db.QueryRow(
    "INSERT INTO people(name, age, gender) VALUES($1, $2, $3) RETURNING id",
    p.Name, p.Age, p.Gender).Scan(&p.ID)

    if err != nil {
      return err
    }

    return nil
  }

func (p *person) updatePerson(db *sql.DB) error {
  _, err :=
  db.Exec("UPDATE people SET name=$1, age=$2, gender = $3 WHERE id=$4",
  p.Name, p.Age, p.Gender, p.ID)

  return err
}

func (p *person) deletePerson(db *sql.DB) error {
  _, err := db.Exec("DELETE FROM people WHERE id=$1", p.ID)

  return err
}

func getPeople(db *sql.DB, start, count int) ([]person, error) {
  rows, err := db.Query(
    "SELECT id, name, age, gender FROM people LIMIT $1 OFFSET $2",
    count, start)

  if err != nil {
    return nil, err
  }

  defer rows.Close()

  people := []person{}

  for rows.Next() {
    var p person
    if err := rows.Scan(&p.ID, &p.Name, &p.Age, &p.Gender); err != nil {
      return nil, err
    }
    people = append(people, p)
  }

  return people, nil
}
