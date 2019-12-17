
// users_test.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	_ "strconv"
	"strings"
	"testing"
	"time"
)

const (
	Updated         string = "updated"
	Created         string = "created"
	Active          string = "active"
  UsersURL string = "/api/v1/namespace/pavedroad.io/users/%s"
)

var newUsersJSON=`{
	"usersuuid": "503e571c-115d-4320-8268-b251ef0302d7",
	"id": "2dnAeu0CaYVyDZI",
	"updated": "2019-12-16T17:13:06-05:00",
	"created": "2019-12-16T17:13:06-05:00",
	"metadata": {
		"id": "WOa6Tk5JuPCd9ZU",
		"test": {
			"key": "WOUtttwtzrG5JqY"
		}
	}
}`


var a UsersApp

func TestMain(m *testing.M) {
	a = UsersApp{}
	a.Initialize()

	clearDB()
	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		fmt.Println("Table check failed:", err)
		log.Fatal(err)
	}

	if _, err := a.DB.Exec(indexCreate); err != nil {
		fmt.Println("Table check failed:", err)
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM Acme.Users")
}

func clearDB() {
	a.DB.Exec("DROP DATABASE IF EXISTS Acme")
	a.DB.Exec("CREATE DATABASE Acme")
}

const tableCreationQuery = `
CREATE TABLE IF NOT EXISTS Acme.users (
    UsersUUID UUID DEFAULT uuid_v4()::UUID PRIMARY KEY,
    users JSONB
);`


const indexCreate = `
CREATE INDEX IF NOT EXISTS usersIdx ON Acme.users USING GIN (users);`


func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/api/v1/namespace/pavedroad.io/usersLIST", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
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

// TestGetWithBadUserUUID
// Get a users with an invalid UUID, should return 400
// and that it is an invalid UUID
//
func TestGetWithBadUserUUID(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET",
		"/api/v1/namespace/pavedroad.io/users/43ae99c9", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "400: invalid UUID: 43ae99c9" {
		t.Errorf("Expected the 'error' key of the response to be set to '400: invalid UUID: 43ae99c9'. Got '%s'", m["error"])
	}
}

// TestGetWrongUUID
// Is a valid UUID, but with leading zeros
// This will not be found and should return a 304
//
func TestGetWrongUUID(t *testing.T) {
	clearTable()
  nt := NewUsers()
  addUsers(nt)
  badUid := "00000000-d01d-4c09-a4e7-59026d143b89"

	statement := fmt.Sprintf(UsersURL, badUid)

	req, _ := http.NewRequest("GET", statement, nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}
// TestCreate
// Use sample data from newUsersJSON) to create
// a new record.
// TODO:
//  need to assert tests for subattributes being present
//
func TestCreateUsers(t *testing.T) {
	clearTable()

	payload := []byte(newUsersJSON)

	req, _ := http.NewRequest("POST", "/api/v1/namespace/pavedroad.io/users", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	//var md map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	//Test we can decode the data
	cs, ok := m["created"].(string)
	if ok {
		c, err := time.Parse(time.RFC3339, cs)
		if err != nil {
			t.Errorf("Parse failed on parse creataed time Got '%v'", c)
		}
	} else {
		t.Errorf("Expected creataed of string type Got '%v'", reflect.TypeOf(m["Created"]))
	}

	us, ok := m["updated"].(string)
	if ok {
		u, err := time.Parse(time.RFC3339, us)
		if err != nil {
			t.Errorf("Parse failed on parse updated time Got '%v'", u)
		}
	} else {
		t.Errorf("Expected updated of string type Got '%v'", reflect.TypeOf(m["Updated"]))
	}
}

func TestMarshallUsers(t *testing.T) {
	nt := NewUsers()
	_, err := json.Marshal(nt)
	if err != nil {
		t.Errorf("Marshal of Users failed: Got '%v'", err)
	}
}

// addUsers
// Inserts a new user into the database and returns the UUID
// for the record that was created
//
func addUsers(t *users) (string) {

  statement := fmt.Sprintf("INSERT INTO Acme.users(users) VALUES('%s') RETURNING usersUUID", newUsersJSON)
  rows, er1 := a.DB.Query(statement)

	if er1 != nil {
		log.Printf("Insert failed error %s", er1)
		return ""
	}

	defer rows.Close()

  for rows.Next() {
    err := rows.Scan(&t.UsersUUID)
    if err != nil {
      return ""
    }
  }

  return t.UsersUUID
}

// NewUsers
// Create a new instance of Users
// Iterate over the struct setting random values
// 
func NewUsers() (t *users) {
	var n users
  json.Unmarshal([]byte(newUsersJSON), &n) 
	return &n
}

//test getting a users
func TestGetUsers(t *testing.T) {
	clearTable()
	nt := NewUsers()
	uid := addUsers(nt)
	statement := fmt.Sprintf(UsersURL, uid)

	req, err := http.NewRequest("GET", statement, nil)
  if err != nil {
		panic(err)
  }

	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}
// TestUpdateUsers
func TestUpdateusers(t *testing.T) {
	clearTable()
	nt := NewUsers()
	uid := addUsers(nt)

	statement := fmt.Sprintf(UsersURL, uid)
	req, _ := http.NewRequest("GET", statement, nil)
	response := executeRequest(req)

	json.Unmarshal(response.Body.Bytes(), &nt)

	ut := nt

	//Update the new struct
	//ut.Active = "eslaf"

	jb, err := json.Marshal(ut)
	if err != nil {
		panic(err)
	}

	req, _ = http.NewRequest("PUT", statement, strings.NewReader(string(jb)))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

//	if m["active"] != "eslaf" {
//		t.Errorf("Expected active to be eslaf. Got %v", m["active"])
//	}
}

func TestDeleteusers(t *testing.T) {
	clearTable()
	nt := NewUsers()
	uid := addUsers(nt)

	statement := fmt.Sprintf(UsersURL, uid)
	req, _ := http.NewRequest("DELETE", statement, nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", statement, nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

/*
func TestDumpUsers(t *testing.T) {
	nt := NewUsers()

  err := dumpUser(*nt)

	if err != nil {
		t.Errorf("Expected erro to be nill. Got %v", err)
	}
}
*/