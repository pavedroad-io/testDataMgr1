

//
// Copyright (c) PavedRoad. All rights reserved.
// Licensed under the Apache2. See LICENSE file in the project root for full license information.
//

// User project / copyright / usage information
// Microservice for managing a backend persistent store for an object

package main

import (
  "context"
  "database/sql"
  "encoding/json"
  "fmt"
  "github.com/gorilla/mux"
  _ "github.com/lib/pq"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strconv"
  "time"
)

// Initialize setups database connection object and the http server
//
func (a *UsersApp) Initialize() {

  // Override defaults
  a.initializeEnvironment()

  // Build connection strings
  connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s host=%s port=%s",
    dbconf.username,
    dbconf.password,
    dbconf.database,
    dbconf.sslMode,
    dbconf.ip,
    dbconf.port)

  httpconf.listenString = fmt.Sprintf("%s:%s", httpconf.ip, httpconf.port)

  var err error
  a.DB, err = sql.Open(dbconf.dbDriver, connectionString)
  if err != nil {
    log.Fatal(err)
  }

  a.Router = mux.NewRouter()
  a.initializeRoutes()
}

// Start the server
func (a *UsersApp) Run(addr string) {

  log.Println("Listing at: " + addr)
  srv := &http.Server{
    Handler:      a.Router,
    Addr:         addr,
    WriteTimeout: httpconf.writeTimeout * time.Second,
    ReadTimeout:  httpconf.readTimeout * time.Second,
  }

  go func() {
    if err := srv.ListenAndServe(); err != nil {
      log.Println(err)
    }
  }()

  // Listen for SIGHUP
  c := make(chan os.Signal, 1)
  <-c

  // Create a deadline to wait for.
  ctx, cancel := context.WithTimeout(context.Background(), httpconf.shutdownTimeout)
  defer cancel()

  // Doesn't block if no connections, but will otherwise wait
  // until the timeout deadline.
  srv.Shutdown(ctx)
  log.Println("shutting down")
  os.Exit(0)
}

// Get for ennvironment variable overrides
func (a *UsersApp) initializeEnvironment() {
  var envVar = ""

  //look for environment variables overrides
  envVar = os.Getenv("APP_DB_USERNAME")
  if envVar != "" {
    dbconf.username = envVar
  }

  envVar = os.Getenv("APP_DB_PASSWORD")
  if envVar != "" {
    dbconf.password = envVar
  }

  envVar = os.Getenv("APP_DB_NAME")
  if envVar != "" {
    dbconf.database = envVar
  }
  envVar = os.Getenv("APP_DB_SSL_MODE")
  if envVar != "" {
    dbconf.sslMode = envVar
  }

  envVar = os.Getenv("APP_DB_SQL_DRIVER")
  if envVar != "" {
    dbconf.dbDriver = envVar
  }

  envVar = os.Getenv("APP_DB_IP")
  if envVar != "" {
    dbconf.ip = envVar
  }

   envVar = os.Getenv("APP_DB_PORT")
  if envVar != "" {
    dbconf.port = envVar
  }

  envVar = os.Getenv("HTTP_IP_ADDR")
  if envVar != "" {
    httpconf.ip = envVar
  }

  envVar = os.Getenv("HTTP_IP_PORT")
  if envVar != "" {
    httpconf.port = envVar
  }

  envVar = os.Getenv("HTTP_READ_TIMEOUT")
  if envVar != "" {
    to, err := strconv.Atoi(envVar)
    if err == nil {
      log.Printf("failed to convert HTTP_READ_TIMEOUT: %s to int", envVar)
    } else {
      httpconf.readTimeout = time.Duration(to) * time.Second
    }
    log.Printf("Read timeout: %d", httpconf.readTimeout)
  }

  envVar = os.Getenv("HTTP_WRITE_TIMEOUT")
  if envVar != "" {
    to, err := strconv.Atoi(envVar)
    if err == nil {
      log.Printf("failed to convert HTTP_READ_TIMEOUT: %s to int", envVar)
    } else {
      httpconf.writeTimeout = time.Duration(to) * time.Second
    }
    log.Printf("Write timeout: %d", httpconf.writeTimeout)
  }

  envVar = os.Getenv("HTTP_SHUTDOWN_TIMEOUT")
  if envVar != "" {
    if envVar != "" {
      to, err := strconv.Atoi(envVar)
      if err != nil {
        httpconf.shutdownTimeout = time.Second * time.Duration(to)
      } else {
        httpconf.shutdownTimeout = time.Second * httpconf.shutdownTimeout
      }
      log.Println("Shutdown timeout", httpconf.shutdownTimeout)
    }
  }

  envVar = os.Getenv("HTTP_LOG")
  if envVar != "" {
    httpconf.logPath = envVar
  }

}


func (a *UsersApp) initializeRoutes() {
  uri := UsersAPIVersion + "/" + UsersNamespaceID + "/{namespace}/" +
    UsersResourceType + "LIST"
  a.Router.HandleFunc(uri, a.listUsers).Methods("GET")

  uri = UsersAPIVersion + "/" + UsersNamespaceID + "/{namespace}/" +
    UsersResourceType + "/{key}"
  a.Router.HandleFunc(uri, a.getUsers).Methods("GET")

  uri = UsersAPIVersion + "/" + UsersNamespaceID + "/{namespace}/" + UsersResourceType
  a.Router.HandleFunc(uri, a.createUsers).Methods("POST")

  uri = UsersAPIVersion + "/" + UsersNamespaceID + "/{namespace}/" +
    UsersResourceType + UsersKey
  a.Router.HandleFunc(uri, a.updateUsers).Methods("PUT")

  uri = UsersAPIVersion + "/" + UsersNamespaceID + "/{namespace}/" +
    UsersResourceType + UsersKey
  a.Router.HandleFunc(uri, a.deleteUsers).Methods("DELETE")
}


// listUsers swagger:route GET /api/v1/namespace/pavedroad.io/usersLIST users listusers
//
// Returns a list of users
//
// Responses:
//    default: genericError
//        200: usersList

func (a *UsersApp) listUsers(w http.ResponseWriter, r *http.Request) {
  users := users{}

  count, _ := strconv.Atoi(r.FormValue("count"))
  start, _ := strconv.Atoi(r.FormValue("start"))

  if count > 10 || count < 1 {
    count = 10
  }
  if start < 0 {
    start = 0
  }

  mappings, err := users.listUsers(a.DB, start, count)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, err.Error())
    return
  }

  respondWithJSON(w, http.StatusOK, mappings)
}


// getUsers swagger:route GET /api/v1/namespace/pavedroad.io/users/{uuid} users getusers
//
// Returns a users given a key, where key is a UUID
//
// Responses:
//    default: genericError
//        200: usersResponse

func (a *UsersApp) getUsers(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  users := users{}

  //TODO: allows them to specify the column used to retrieve user
  err := users.getUsers(a.DB, vars["key"], UUID)

  if err != nil {
    errmsg := err.Error()
    errno :=  errmsg[0:3]
    if errno == "400" {
      respondWithError(w, http.StatusBadRequest, err.Error())
    } else {
      respondWithError(w, http.StatusNotFound, err.Error())
    }
    return
  }

  respondWithJSON(w, http.StatusOK, users)
}


// createUsers swagger:route POST /api/v1/namespace/pavedroad.io/users users createusers
//
// Create a new users
//
// Responses:
//    default: genericError
//        201: usersResponse
//        400: genericError
func (a *UsersApp) createUsers(w http.ResponseWriter, r *http.Request) {
  // New map structure
  users := users{}

  htmlData, err := ioutil.ReadAll(r.Body)
  if err != nil {
    log.Println(err)
    os.Exit(1)
  }

  err = json.Unmarshal(htmlData, &users)
  if err != nil {
    log.Println(err)
    os.Exit(1)
  }

  ct := time.Now().UTC()
  users.Created = ct
  users.Updated = ct

  // Save into backend storage
  // returns the UUID if needed
  if _, err := users.createUsers(a.DB); err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid request payload")
    return
  }

  respondWithJSON(w, http.StatusCreated, users)
}


// updateUsers swagger:route PUT /api/v1/namespace/pavedroad.io/users/{key} users updateusers
//
// Update a users specified by key, where key is a uuid
//
// Responses:
//    default: genericError
//        201: usersResponse
//        400: genericError
func (a *UsersApp) updateUsers(w http.ResponseWriter, r *http.Request) {
  users := users{}

  // Read URI variables
  // vars := mux.Vars(r)

  htmlData, err := ioutil.ReadAll(r.Body)
  if err != nil {
    log.Println(err)
    return
  }

  err = json.Unmarshal(htmlData, &users)
  if err != nil {
    log.Println(err)
    return
  }

  ct := time.Now().UTC()
  users.Updated = ct

  if err := users.updateUsers(a.DB, users.UsersUUID); err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid request payload")
    return
  }

  respondWithJSON(w, http.StatusOK, users)
}


// deleteUsers swagger:route DELETE /api/v1/namespace/pavedroad.io/users/{key} users deleteusers
//
// Update a users specified by key, which is a uuid
//
// Responses:
//    default: genericError
//        200: usersResponse
//        400: genericError
func (a *UsersApp) deleteUsers(w http.ResponseWriter, r *http.Request) {
  users := users{}
  vars := mux.Vars(r)

  err := users.deleteUsers(a.DB, vars["key"])
  if err != nil {
    respondWithError(w, http.StatusNotFound, err.Error())
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

func logRequest(handler http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
    handler.ServeHTTP(w, r)
  })
}

func openLogFile(logfile string) {
  if logfile != "" {
    lf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)

    if err != nil {
      log.Fatal("OpenLogfile: os.OpenFile:", err)
    }
    log.SetOutput(lf)
  }
}

/*
func dumpUsers(m Users) {
  fmt.Println("Dump users")
  
}
*/
