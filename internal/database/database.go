// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package database

import (
	"database/sql"
	"fmt"
	"github.com/gofrs/uuid"
	"net/netip"
)

type Database struct {
	*sql.DB
	get_action    *sql.Stmt
	insert_action *sql.Stmt
	update_action *sql.Stmt
}

func NewDatabase(db *sql.DB) (*Database, error) {
	// UplinkGTP4
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS uplink_gtp4 (
		uplink_teid INTEGER,
		srgw_ip INET,
		action_uuid UUID NOT NULL,
		PRIMARY KEY(uplink_teid, srgw_ip)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not create table uplink_gtp4 in database: %s", err)
	}

	get_action, err := db.Prepare(`SELECT action_uuid FROM uplink_gtp4 WHERE (uplink_teid = $1 AND srgw_ip = $2)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for get_action: %s", err)
	}
	insert_action, err := db.Prepare(`INSERT INTO uplink_gtp4 (uplink_teid, srgw_ip, action_uuid) VALUES($1, $2, $3)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for insert: %s", err)
	}

	update_action, err := db.Prepare(`UPDATE uplink_gtp4 SET action_uuid = $3 WHERE (uplink_teid =$1 AND srgw_ip = $2)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for update: %s", err)
	}

	return &Database{
		DB:            db,
		get_action:    get_action,
		insert_action: insert_action,
		update_action: update_action,
	}, nil
}

func (db *Database) InsertAction(uplinkTeid uint32, SrgwIp netip.Addr, actionUuid uuid.UUID) error {
	_, err := db.insert_action.Exec(uplinkTeid, SrgwIp.String(), actionUuid.String())
	return err
}

func (db *Database) UpdateAction(uplinkTeid uint32, SrgwIp netip.Addr, actionUuid uuid.UUID) error {
	_, err := db.update_action.Exec(uplinkTeid, SrgwIp.String(), actionUuid.String())
	return err
}

func (db *Database) GetAction(UplinkTeid uint32, SrgwIp netip.Addr) (uuid.UUID, error) {
	actionUuid := uuid.UUID{}
	err := db.get_action.QueryRow(UplinkTeid, SrgwIp.String()).Scan(&actionUuid)
	return actionUuid, err
}
