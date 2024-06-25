// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	app_api "github.com/nextmn/srv6/internal/app/api"
	"os"
)

// DBTask initializes the database
type DBTask struct {
	WithName
	WithState
	db       *sql.DB
	registry app_api.Registry
	user     string
	port     string
	password string
	host     string
	dbname   string
}

// Create a new DBTask
func NewDBTask(name string, registry app_api.Registry) *DBTask {
	return &DBTask{
		WithName:  NewName(name),
		WithState: NewState(),
		registry:  registry,
	}
}

// Init
func (db *DBTask) RunInit() error {
	db.state = true
	// Getting config from environment
	host, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok {
		return fmt.Errorf("No host provided for postgres")
	}
	db.host = host
	port, ok := os.LookupEnv("POSTGRES_PORT")
	db.port = port
	if !ok {
		db.port = "5432"
	}
	user, ok := os.LookupEnv("POSTGRES_USER")
	db.user = user
	if !ok {
		db.user = "postgres"
	}
	password, ok := os.LookupEnv("POSTGRES_PASSWORD")
	db.password = password
	if !ok {
		password_file, ok := os.LookupEnv("POSTGRES_PASSWORD_FILE")
		if !ok {
			return fmt.Errorf("No password provided for postgres")
		}
		b, err := os.ReadFile(password_file)
		if err != nil {
			return fmt.Errorf("Could not read file %s to get postgres password", password_file)
		}
		db.password = string(b)
	}
	cr, ok := db.registry.ControllerRegistry()
	if !ok {
		return fmt.Errorf("No controller registry")
	}
	db.dbname = cr.Resource
	if db.dbname == "" {
		return fmt.Errorf("Empty resource name for the router: cannot create db")
	}

	// Create database on postgres
	conninfo := fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' sslmode='disable'", db.host, db.port, db.user, db.password, db.dbname)
	initdb, err := sql.Open("postgres", conninfo)
	if err != nil {
		return fmt.Errorf("Could not open postgres database")
	}
	defer initdb.Close()
	if err := initdb.Ping(); err != nil {
		return fmt.Errorf("Could not connect to postgres database: %s", err)
	}
	if _, err := initdb.Exec(fmt.Sprintf("CREATE DATABASE %s", db.dbname)); err != nil {
		return fmt.Errorf("Could not create database: %s", err)
	}

	// Create a conn for this database
	psqlconn := fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' dbname='%s' sslmode='disable'", db.host, db.port, db.user, db.password, db.dbname)
	database, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return fmt.Errorf("Error while openning postgres database: %s", err)
	}
	db.db = database
	if err := db.db.Ping(); err != nil {
		return fmt.Errorf("Could not connect to postgres database: %s", err)
	}

	return nil
}

// Exit
func (db *DBTask) RunExit() error {
	db.state = false
	if db.db == nil {
		return fmt.Errorf("No database")
	}
	db.db.Close()
	// delete database on postgres
	conninfo := fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' sslmode='disable'", db.host, db.port, db.user, db.password, db.dbname)
	rmdb, err := sql.Open("postgres", conninfo)
	if err != nil {
		return fmt.Errorf("Could not open postgres database: %s", err)
	}
	defer rmdb.Close()
	if err := rmdb.Ping(); err != nil {
		return fmt.Errorf("Could not connect to postgres database: %s", err)
	}
	if _, err := rmdb.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", db.dbname)); err != nil {
		return fmt.Errorf("Could not create database: %s", err)
	}
	return nil
}
