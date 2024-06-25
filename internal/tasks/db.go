// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

// DBTask initializes the database
type DBTask struct {
	WithName
	WithState
	dbname string
	db     *sql.DB
}

// Create a new DBTask
func NewDBTask(name string) *DBTask {
	dbname := "toto"
	return &DBTask{
		WithName:  NewName(name),
		WithState: NewState(),
		dbname:    dbname,
	}
}

// Init
func (db *DBTask) RunInit() error {
	db.state = true
	host, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok {
		return fmt.Errorf("No host provided for postgres")
	}
	port, ok := os.LookupEnv("POSTGRES_PORT")
	if !ok {
		port = "5432"
	}
	user, ok := os.LookupEnv("POSTGRES_USER")
	if !ok {
		user = "postgres"
	}
	password, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		password_file, ok := os.LookupEnv("POSTGRES_PASSWORD_FILE")
		if !ok {
			return fmt.Errorf("No password provided for postgres")
		}
		b, err := os.ReadFile(password_file)
		if err != nil {
			return fmt.Errorf("Could not read file %s to get postgres password", password_file)
		}
		password = string(b)
	}
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, db.dbname)
	database, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return fmt.Errorf("Error while openning postgres database")
	}
	db.db = database
	if err := db.db.Ping(); err != nil {
		return fmt.Errorf("Could not connect to postgres database")
	}

	return nil
}

// Exit
func (db *DBTask) RunExit() error {
	db.state = false
	db.db.Close()
	return nil
}
