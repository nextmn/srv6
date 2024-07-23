// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	app_api "github.com/nextmn/srv6/internal/app/api"
	"github.com/nextmn/srv6/internal/database"
)

// DBTask initializes the database
type DBTask struct {
	WithName
	WithState
	db       *database.Database
	user     string
	port     string
	password string
	host     string
	dbname   string
	registry app_api.Registry
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
func (db *DBTask) RunInit(ctx context.Context) error {
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
		user_file, ok := os.LookupEnv("POSTGRES_USER_FILE")
		if !ok {
			db.user = "postgres"
		} else {
			b, err := os.ReadFile(user_file)
			if err != nil {
				return fmt.Errorf("Could not read file %s to get postgres user", user_file)
			}
			db.user = string(b)
		}
	}
	dbname, ok := os.LookupEnv("POSTGRES_DB")
	db.dbname = dbname
	if !ok {
		db.dbname = db.user
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

	// Create a conn for this database
	psqlconn := fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' dbname='%s' sslmode='disable'", db.host, db.port, db.user, db.password, db.dbname)
	postgres, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return fmt.Errorf("Error while openning postgres database: %s", err)
	}

	maxAttempts := 16
	ok = false
	for errcnt := 0; (errcnt < maxAttempts) && !ok; errcnt++ {
		if errcnt > 2 {
			log.Printf("Could not connect to postgres database. Retrying (attempt %d)\n", errcnt)
		}
		wait, cancel := context.WithTimeout(ctx, 100*(1<<errcnt*time.Millisecond)) // Exponential backoff
		if err := postgres.Ping(); err == nil {
			ok = true
			cancel()
		}
		select {
		case <-ctx.Done():
			break // abort
		case <-wait.Done():
		}
	}
	if !ok {
		return fmt.Errorf("Could not connect to postgres database after %d attempts: %s", maxAttempts, err)
	}

	db.db = database.NewDatabase(postgres)
	if err := db.db.Init(ctx); err != nil {
		return err
	}

	if db.registry != nil {
		db.registry.RegisterDB(db.db)
	}

	return nil
}

// Exit
func (db *DBTask) RunExit() error {
	db.state = false
	if db.registry != nil {
		db.registry.DeleteDB()
	}
	if db.db == nil {
		return fmt.Errorf("No database")
	}
	db.db.Exit()
	db.db.Close()
	return nil
}
