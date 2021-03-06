//
// Copyright (c) PavedRoad. All rights reserved.
// Licensed under the Apache2. See LICENSE file in the project root for full license information.
//

// User project / copyright / usage information
// Microservice for managing a backend persistent store for an object

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"os"
	"time"
)

// Contants to build up a k8s style URL
const (
	// UsersAPIVersion Version API URL
	UsersAPIVersion string = "/api/v1"
	// UsersNamespaceID Prefix for namespaces
	UsersNamespaceID string = "namespace"
	// UsersDefaultNamespace Default namespace
	UsersDefaultNamespace string = "pavedroad.io"
	// UsersResourceType CRD Type per k8s
	UsersResourceType string = "users"
	// The email or account login used by 3rd parth provider
	UsersKey string = "/{key}"
)

// Options for looking up a user
const (
	UUID = iota
	NAME
)

// holds pointers to database and http server
type UsersApp struct {
	Router *mux.Router
	DB     *sql.DB
}

// both db and http configuration can be changed using environment varialbes
type databaseConfig struct {
	username string
	password string
	database string
	sslMode  string
	dbDriver string
	ip       string
	port     string
}

// HTTP server configuration
type httpConfig struct {
	ip              string
	port            string
	shutdownTimeout time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	listenString    string
	logPath         string
}

// Global for use in the module

// Set default database configuration
var dbconf = databaseConfig{username: "root", password: "", database: "pavedroad", sslMode: "disable", dbDriver: "postgres", ip: "127.0.0.1", port: "26257"}

// Set default http configuration
var httpconf = httpConfig{ip: "127.0.0.1", port: "8082", shutdownTimeout: 15, readTimeout: 60, writeTimeout: 60, listenString: "127.0.0.1:8082", logPath: "logs/users.log"}

// shutdownTimeout will be initialized based on the default or HTTP_SHUTDOWN_TIMEOUT
var shutdowTimeout time.Duration

// GitTag is used for namespace functionality
var GitTag string

// Vesion release
var Version string

// Build release
var Build string

// printVersion
func printVersion() {
	fmt.Printf("{\"Version\": \"%v\", \"Build\": \"%v\", \"GitTag\": \"%v\"}\n",
		Version, Build, GitTag)
	os.Exit(0)
}

// main entry point for server
func main() {

	versionFlag := flag.Bool("v", false, "Print version information")
	flag.Parse()

	if *versionFlag {
		printVersion()
	}

	// Setup loggin
	openLogFile(httpconf.logPath)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Logfile opened %s", httpconf.logPath)

	a := UsersApp{}
	a.Initialize()
	a.Run(httpconf.listenString)
}
