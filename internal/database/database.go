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
	get_action           *sql.Stmt
	insert_action        *sql.Stmt
	update_action        *sql.Stmt
	insert_uplink_rule   *sql.Stmt
	insert_downlink_rule *sql.Stmt
	enable_rule          *sql.Stmt
	disable_rule         *sql.Stmt
	delete_rule          *sql.Stmt
}

func NewDatabase(db *sql.DB) (*Database, error) {
	_, err := db.Exec(database_sql)
	if err != nil {
		return nil, fmt.Errorf("Could not initialize database: %s", err)
	}

	get_action, err := db.Prepare(`SELECT action_uuid FROM uplink_gtp4 WHERE (uplink_teid = $1 AND srgw_ip = $2 AND gnb_ip = $3)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for get_action: %s", err)
	}
	insert_action, err := db.Prepare(`INSERT INTO uplink_gtp4 (uplink_teid, srgw_ip, gnb_ip, action_uuid) VALUES($1, $2, $3, $4)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for insert: %s", err)
	}

	update_action, err := db.Prepare(`UPDATE uplink_gtp4 SET action_uuid = $4 WHERE (uplink_teid = $1 AND srgw_ip = $2 AND gnb_ip = $3)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for update: %s", err)
	}

	insert_uplink_rule, err := db.Prepare(`CALL insert_uplink_rule($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for insert_uplink_rule: %s", err)
	}
	insert_downlink_rule, err := db.Prepare(`CALL insert_downlink_rule($1, $2, $3, $4, $5)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for insert_downlink_rule: %s", err)
	}

	enable_rule, err := db.Prepare(`UPDATE rule SET enabled = true WHERE uuid = $1`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for enable_rule: %s", err)
	}
	disable_rule, err := db.Prepare(`UPDATE rule SET enabled = false WHERE uuid = $1`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for disable_rule: %s", err)
	}

	delete_rule, err := db.Prepare(`DELETE FROM rule WHERE uuid = $1`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for disable_rule: %s", err)
	}
	return &Database{
		DB:                   db,
		get_action:           get_action,
		insert_action:        insert_action,
		update_action:        update_action,
		insert_uplink_rule:   insert_uplink_rule,
		insert_downlink_rule: insert_downlink_rule,
		enable_rule:          enable_rule,
		disable_rule:         disable_rule,
		delete_rule:          delete_rule,
	}, nil
}

func (db *Database) InsertRule(uuid uuid.UUID, r jsonapi.Rule) error {
	srh := []string{}
	for _, ip := range r.Action.SRH {
		srh = append(srh, ip.String())
	}
	switch r.Type {
	case "uplink":
		_, err := db.insert_uplink_rule.Exec(uuid.String(), r.Enabled, r.Match.UEIpPrefix.String(), r.Match.GNBIpPrefix.String(), r.Action.NextHop.String(), pq.Array(srh))
		return err
	case "downlink":
		_, err := db.insert_downlink_rule.Exec(uuid.String(), r.Enabled, r.Match.UEIpPrefix.String(), r.Action.NextHop.String(), pq.Array(srh))
		return err
	default:
		return fmt.Errorf("Wrong type for the rule")
	}
}

func (db *Database) EnableRule(uuid uuid.UUID) error {
	_, err := db.enable_rule.Exec(uuid.String())
	return err
}

func (db *Database) DisableRule(uuid uuid.UUID) error {
	_, err := db.disable_rule.Exec(uuid.String())
	return err
}

func (db *Database) InsertAction(uplinkTeid uint32, srgwIp netip.Addr, gnbIp netip.Addr, actionUuid uuid.UUID) error {
	_, err := db.insert_action.Exec(uplinkTeid, srgwIp.String(), gnbIp.String(), actionUuid.String())
	return err
}

func (db *Database) UpdateAction(uplinkTeid uint32, srgwIp netip.Addr, gnbIp netip.Addr, actionUuid uuid.UUID) error {
	_, err := db.update_action.Exec(uplinkTeid, srgwIp.String(), gnbIp.String(), actionUuid.String())
	return err
}

func (db *Database) GetUplinkAction(uplinkTeid uint32, srgwIp netip.Addr, gnbIp netip.Addr) (uuid.UUID, error) {
	actionUuid := uuid.UUID{}
	err := db.get_action.QueryRow(uplinkTeid, srgwIp.String(), gnbIp.String()).Scan(&actionUuid)
	return actionUuid, err
}
