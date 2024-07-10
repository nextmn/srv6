// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lib/pq"
	"github.com/nextmn/json-api/jsonapi"
	"net/netip"
)

//go:embed database.sql
var database_sql string

type Database struct {
	*sql.DB
	stmt map[string]*sql.Stmt
}

func (db *Database) prepare(name string, query string) error {
	s, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("Could not prepare statement %s: %s", s, err)
	}
	db.stmt[name] = s
	return nil
}

func NewDatabase(db *sql.DB) (*Database, error) {
	_, err := db.Exec(database_sql)
	if err != nil {
		return nil, fmt.Errorf("Could not initialize database: %s", err)
	}
	l := map[string]string{
		"get_action":           `SELECT action_uuid FROM uplink_gtp4 WHERE (uplink_teid = $1 AND srgw_ip = $2 AND gnb_ip = $3)`,
		"insert_action":        `CALL insert_action($1, $2, $3, $4)`,
		"update_action":        `CALL update_action($1, $2, $3, $4)`,
		"insert_uplink_rule":   `CALL insert_uplink_rule($1, $2, $3, $4, $5, $6)`,
		"insert_downlink_rule": `CALL insert_downlink_rule($1, $2, $3, $4, $5)`,
		"enable_rule":          `CALL enable_rule($1)`,
		"disable_rule":         `CALL disable_rule($1)`,
		"delete_rule":          `CALL delete_rule($1)`,
	}
	stmt := make(map[string]*sql.Stmt)

	for k, v := range l {
		s, err := db.Prepare(v)
		if err != nil {
			return nil, fmt.Errorf("Could not prepare statement %s: %s", k, err)
		}
		stmt[k] = s
	}
	return &Database{
		DB:   db,
		stmt: stmt,
	}, nil
}

func (db *Database) Exit() {
	for k, v := range db.stmt {
		v.Close()
		delete(db.stmt, k)
	}
}

func (db *Database) InsertRule(uuid uuid.UUID, r jsonapi.Rule) error {
	srh := []string{}
	for _, ip := range r.Action.SRH {
		srh = append(srh, ip.String())
	}
	switch r.Type {
	case "uplink":
		if stmt, ok := db.stmt["insert_uplink_rule"]; ok {
			_, err := stmt.Exec(uuid.String(), r.Enabled, r.Match.UEIpPrefix.String(), r.Match.GNBIpPrefix.String(), r.Action.NextHop.String(), pq.Array(srh))
			return err
		} else {
			return fmt.Errorf("Procedure not registered")
		}
	case "downlink":
		if stmt, ok := db.stmt["insert_downlink_rule"]; ok {
			_, err := stmt.Exec(uuid.String(), r.Enabled, r.Match.UEIpPrefix.String(), r.Action.NextHop.String(), pq.Array(srh))
			return err
		} else {
			return fmt.Errorf("Procedure not registered")
		}
	default:
		return fmt.Errorf("Wrong type for the rule")
	}
}

func (db *Database) EnableRule(uuid uuid.UUID) error {
	if stmt, ok := db.stmt["enable_rule"]; ok {
		_, err := stmt.Exec(uuid.String())
		return err
	} else {
		return fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) DisableRule(uuid uuid.UUID) error {
	if stmt, ok := db.stmt["disable_rule"]; ok {
		_, err := stmt.Exec(uuid.String())
		return err
	} else {
		return fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) InsertAction(uplinkTeid uint32, srgwIp netip.Addr, gnbIp netip.Addr, actionUuid uuid.UUID) error {
	if stmt, ok := db.stmt["insert_action"]; ok {
		_, err := stmt.Exec(uplinkTeid, srgwIp.String(), gnbIp.String(), actionUuid.String())
		return err
	} else {
		return fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) UpdateAction(uplinkTeid uint32, srgwIp netip.Addr, gnbIp netip.Addr, actionUuid uuid.UUID) error {
	if stmt, ok := db.stmt["update_action"]; ok {
		_, err := stmt.Exec(uplinkTeid, srgwIp.String(), gnbIp.String(), actionUuid.String())
		return err
	} else {
		return fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) GetUplinkAction(uplinkTeid uint32, srgwIp netip.Addr, gnbIp netip.Addr) (uuid.UUID, error) {
	actionUuid := uuid.UUID{}
	if stmt, ok := db.stmt["get_action"]; ok {
		err := stmt.QueryRow(uplinkTeid, srgwIp.String(), gnbIp.String()).Scan(&actionUuid)
		return actionUuid, err
	} else {
		return actionUuid, fmt.Errorf("Procedure not registered")
	}
}
