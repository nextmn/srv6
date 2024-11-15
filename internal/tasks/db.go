// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"
	"time"

	_ "github.com/lib/pq"
	app_api "github.com/nextmn/srv6/internal/app/api"
	"github.com/nextmn/srv6/internal/database"
	"github.com/sirupsen/logrus"
)

// DBTask initializes the database
type DBTask struct {
	WithName
	WithState
	db             *database.Database
	user           string
	port           string
	password       string
	host           string
	dbname         string
	unixSocketPath string
	registry       app_api.Registry
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

	var psqlconn string
	unixSocket, ok := os.LookupEnv("POSTGRES_UNIX_SOCKET_PATH")
	if !ok {
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
		psqlconn = fmt.Sprintf("host='%s' port='%s' user='%s' password='%s' dbname='%s' sslmode='disable'", db.host, db.port, db.user, db.password, db.dbname)
	} else {
		db.unixSocketPath = path.Clean(unixSocket)
		if db.unixSocketPath != "/" {
			db.unixSocketPath += "/"
		}
		psqlconn = fmt.Sprintf("postgres:///%s?host=%s&user=%s", db.dbname, db.unixSocketPath, db.user)
	}

	postgres, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return fmt.Errorf("Error while openning postgres database: %s", err)
	}

	maxAttempts := 16
	ok = false
	for errcnt := 0; (errcnt < maxAttempts) && !ok; errcnt++ {
		wait, cancel := context.WithTimeout(ctx, 100*(1<<errcnt*time.Millisecond)) // Exponential backoff
		defer cancel()
		if err := postgres.Ping(); err == nil {
			ok = true
			logrus.WithFields(logrus.Fields{"attempt": errcnt}).Info("Connected to postgres database.")
			cancel()
		} else if errcnt < maxAttempts-1 {
			logrus.WithFields(logrus.Fields{"attempt": errcnt}).Warn("Could not connect to postgres database. Another attempt is scheduled.")
		} else {
			logrus.WithFields(logrus.Fields{"attempt": errcnt}).Warn("Could not connect to postgres database.")
		}
		// blocks until success, timeout, or ctx.Done()
		select {
		case <-wait.Done():
		}
		// check if wait.Done() is a result of ctx.Done()
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
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
