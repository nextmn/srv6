// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package database

import (
	"database/sql"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/nextmn/json-api/jsonapi"
	"net/netip"
)

type Database struct {
	*sql.DB
	get_action           *sql.Stmt
	insert_action        *sql.Stmt
	update_action        *sql.Stmt
	insert_uplink_rule   *sql.Stmt
	insert_downlink_rule *sql.Stmt
}

func NewDatabase(db *sql.DB) (*Database, error) {
	// UplinkGTP4
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS uplink_gtp4 (
		uplink_teid INTEGER,
		srgw_ip INET,
		gnb_ip INET,
		action_uuid UUID NOT NULL,
		PRIMARY KEY(uplink_teid, srgw_ip, gnb_ip)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not create table uplink_gtp4 in database: %s", err)
	}

	// Rules - Actions
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS action (
		id SERIAL PRIMARY KEY,
		next_hop INET NOT NULL,
		srh INET ARRAY NOT NULL
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not create table action in database: %s", err)
	}

	// Rules - Match
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS match (
		id SERIAL PRIMARY KEY,
		ue_ip_prefix CIDR NOT NULL,
		gnb_ip_prefix CIDR
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not create table match in database: %s", err)
	}

	// Rules
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS rule (
		uuid UUID PRIMARY KEY,
		type_uplink BOOL NOT NULL,
		enabled BOOL NOT NULL,
		match_id INTEGER NOT NULL REFERENCES match(id),
		action_id INTEGER NOT NULL REFERENCES action(id)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not create table rule in database: %s", err)
	}

	_, err = db.Exec(`CREATE OR REPLACE PROCEDURE insert_uplink_rule(IN uuid UUID, IN enabled BOOL, IN ue_ip_prefix CIDR, IN gnb_ip_prefix CIDR, IN next_hop INET, IN srh INET ARRAY)
		LANGUAGE plpgsql AS $$
		BEGIN
			INSERT INTO match(ue_ip_prefix, gnb_ip_prefix) VALUES (ue_ip_prefix, gnb_ip_prefix) RETURNING id AS match_id;
			INSERT INTO action(next_hop, srh) VALUES (next_hop, srh) RETURNING id AS action_id;
			INSERT INTO rule(uuid, type_uplink, enabled, match_id, action_id) VALUES(uuid, TRUE, enabled, match_id, action_id);
		END;$$;
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not create procedure insert_uplink_rule in database: %s", err)
	}

	_, err = db.Exec(`CREATE OR REPLACE PROCEDURE insert_downlink_rule(IN uuid UUID, IN enabled BOOL, IN ue_ip_prefix CIDR, IN next_hop INET, IN srh INET ARRAY)
		LANGUAGE plpgsql AS $$
		BEGIN
			INSERT INTO match(ue_ip_prefix) VALUES (ue_ip_prefix) RETURNING id AS match_id;
			INSERT INTO action(next_hop, srh) VALUES (next_hop, srh) RETURNING id AS action_id;
			INSERT INTO rule(uuid, type_uplink, enabled, match_id, action_id) VALUES(uuid, FALSE, enabled, match_id, action_id);
		END;$$;
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not create procedure insert_downlink_rule in database: %s", err)
	}

	get_action, err := db.Prepare(`SELECT action_uuid FROM uplink_gtp4 WHERE (uplink_teid = $1 AND srgw_ip = $2 AND gnb_ip = $3)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for get_action: %s", err)
	}
	insert_action, err := db.Prepare(`INSERT INTO uplink_gtp4 (uplink_teid, srgw_ip, gnb_ip, action_uuid) VALUES($1, $2, $3, $4)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for insert: %s", err)
	}

	update_action, err := db.Prepare(`UPDATE uplink_gtp4 SET action_uuid = $4 WHERE (uplink_teid =$1 AND srgw_ip = $2 AND gnb_ip = $3)`)
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

	return &Database{
		DB:                   db,
		get_action:           get_action,
		insert_action:        insert_action,
		update_action:        update_action,
		insert_uplink_rule:   insert_uplink_rule,
		insert_downlink_rule: insert_downlink_rule,
	}, nil
}

func (db *Database) InsertRule(uuid uuid.UUID, r jsonapi.Rule) error {
	srh := []string{}
	for _, ip := range r.Action.SRH {
		srh = append(srh, ip.String())
	}
	switch r.Type {
	case "uplink":
		_, err := db.insert_uplink_rule.Exec(uuid.String(), r.Enabled, r.Match.UEIpPrefix.String(), r.Match.GNBIpPrefix.String(), r.Action.NextHop.String(), srh)
		return err
	case "downlink":
		_, err := db.insert_downlink_rule.Exec(uuid.String(), r.Enabled, r.Match.UEIpPrefix.String(), r.Action.NextHop.String(), srh)
		return err
	default:
		return fmt.Errorf("Wrong type for the rule")
	}
}
func (db *Database) InsertAction(uplinkTeid uint32, SrgwIp netip.Addr, GnbIp netip.Addr, actionUuid uuid.UUID) error {
	_, err := db.insert_action.Exec(uplinkTeid, SrgwIp.String(), GnbIp.String(), actionUuid.String())
	return err
}

func (db *Database) UpdateAction(uplinkTeid uint32, SrgwIp netip.Addr, GnbIp netip.Addr, actionUuid uuid.UUID) error {
	_, err := db.update_action.Exec(uplinkTeid, SrgwIp.String(), GnbIp.String(), actionUuid.String())
	return err
}

func (db *Database) GetAction(UplinkTeid uint32, SrgwIp netip.Addr, GnbIp netip.Addr) (uuid.UUID, error) {
	actionUuid := uuid.UUID{}
	err := db.get_action.QueryRow(UplinkTeid, SrgwIp.String(), GnbIp.String()).Scan(&actionUuid)
	return actionUuid, err
}
