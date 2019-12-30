//
// Copyright (c) PavedRoad. All rights reserved.
// Licensed under the Apache2. See LICENSE file in the project root for full license information.
//
// Apache2

// User project / copyright / usage information
// Microservice for managing a backend persistent store for an object

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

// A GenericError is the default error message that is generated.
// For certain status codes there are more appropriate error structures.
//
// swagger:response genericError
type GenericError struct {
	// The error message
	// in: body
	Body struct {
		// Code: integer code for error message
		Code int32 `json:"code"`
		// Message: Error message called with Method()
		Message error `json:"message"`
	} `json:"body"`
}

// Return list of userss
//
// TODO: add method of including subattributes
//
// swagger:response usersList
type listResponse struct {
	// in: body
	UUID string `json:"uuid"`
}

// Generated structures with Swagger docs
// swagger:response test
type test struct {
	// Key
	Key string `json:"key"`
}

// swagger:response metadata
type metadata struct {
	Test test `json:test`
	// Id
	Id string `json:"id"`
}

// swagger:response users
type users struct {
	// UsersUUID into JSONB

	UsersUUID string   `json:usersuuid`
	Metadata  metadata `json:metadata`
	// Id
	Id string `json:"id"`
	// Updated
	Updated time.Time `json:"updated"`
	// Created
	Created time.Time `json:"created"`
}

// UsersResponse model
//
// This is used for returning a response with a single users as body
//
// swagger:response usersResponse
type UsersResponse struct {
	// in: body
	response string `json:"order"`
}

// updateUsers in database
func (t *users) updateUsers(db *sql.DB, key string) error {
	update := `
	UPDATE Acme.users
    SET users = '%s'
  WHERE usersUUID = '%s';`

	jb, err := json.Marshal(t)
	if err != nil {
		log.Println("marshall failed")
		panic(err)
	}

	statement := fmt.Sprintf(update, jb, key)
	_, er1 := db.Query(statement)

	if er1 != nil {
		log.Println("Update failed")
		return er1
	}

	return nil
}

// createUsers in database
func (t *users) createUsers(db *sql.DB) (string, error) {
	jb, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}

	//  statement := fmt.Sprintf("INSERT INTO Acme.users(users) VALUES('%s') RETURNING UsersUUID", jb)
	//  rows, er1 := db.Query(statement)
	statement := `INSERT INTO Acme.users(users) VALUES($1) RETURNING UsersUUID;`
	rows, er1 := db.Query(statement, jb)

	if er1 != nil {
		log.Printf("Insert failed for: %s", t.UsersUUID)
		log.Printf("SQL Error: %s", er1)
		return "", er1
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&t.UsersUUID)
		if err != nil {
			return "", err
		}
	}

	return t.UsersUUID, nil

}

// listUsers: return a list of users
//
func (t *users) listUsers(db *sql.DB, start, count int) ([]listResponse, error) {
	/*
	   qry := `select uuid,
	         users ->> 'active' as active,
	         users -> 'Metadata' ->> 'name' as name
	         from Acme.users LIMIT %d OFFSET %d;`
	*/
	qry := `select UsersUUID
          from Acme.users LIMIT %d OFFSET %d;`
	statement := fmt.Sprintf(qry, count, start)
	rows, err := db.Query(statement)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ul := []listResponse{}

	for rows.Next() {
		var t listResponse
		err := rows.Scan(&t.UUID)

		if err != nil {
			log.Printf("SQL rows.Scan failed: %s", err)
			return ul, err
		}

		ul = append(ul, t)
	}

	return ul, nil
}

// getUsers: return a users based on the key
//
func (t *users) getUsers(db *sql.DB, key string, method int) error {
	var statement string

	switch method {
	case UUID:
		_, err := uuid.Parse(key)
		if err != nil {
			m := fmt.Sprintf("400: invalid UUID: %s", key)
			return errors.New(m)
		}
		statement = `
  SELECT UsersUUID, users
  FROM Acme.users
  WHERE UsersUUID = $1;`
	}

	row := db.QueryRow(statement, key)

	// Fill in mapper
	var jb []byte
	var uid string
	switch err := row.Scan(&uid, &jb); err {

	case sql.ErrNoRows:
		m := fmt.Sprintf("404:name %s does not exist", key)
		return errors.New(m)
	case nil:
		err = json.Unmarshal(jb, t)
		if err != nil {
			m := fmt.Sprintf("400:unmarshal failed %s", key)
			return errors.New(m)
		}
		t.UsersUUID = uid
		break
	default:
		//Some error to catch
		panic(err)
	}

	return nil
}

// deleteUsers: return a users based on UID
//
func (t *users) deleteUsers(db *sql.DB, key string) error {
	statement := `DELETE FROM Acme.users WHERE UsersUUID = $1;`
	result, err := db.Exec(statement, key)
	c, e := result.RowsAffected()

	if e == nil && c == 0 {
		em := fmt.Sprintf("UUID %s does not exist", key)
		log.Println(em)
		log.Println(e)
		return errors.New(em)
	}

	return err
}
