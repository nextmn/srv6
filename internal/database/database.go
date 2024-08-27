// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/netip"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/lib/pq"
	"github.com/nextmn/json-api/jsonapi"
)

//go:generate go run gen.go database.sql

//go:embed database.sql
var database_sql string

type Database struct {
	*sql.DB
	stmt map[string]*sql.Stmt
}

func (db *Database) prepare(ctx context.Context, name string, query string) error {
	s, err := db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("Could not prepare statement %s: %s", name, err)
	}
	db.stmt[name] = s
	return nil
}

func NewDatabase(db *sql.DB) *Database {
	return &Database{
		DB:   db,
		stmt: make(map[string]*sql.Stmt, 0),
	}
}

func (db *Database) Init(ctx context.Context) error {
	_, err := db.Exec(database_sql)
	if err != nil {
		return fmt.Errorf("Could not initialize database: %s", err)
	}
	l := map[string]string{}
	// use generated code
	for k, v := range procedures {
		args := []string{}
		for i := range v.num_in {
			args = append(args, fmt.Sprintf("$%d", i+1))
		}
		for _ = range v.num_out {
			args = append(args, "NULL")
		}
		strargs := strings.Join(args, ", ")
		if v.is_procedure {
			l[k] = fmt.Sprintf("CALL %s(%s)", k, strargs)
		} else {
			l[k] = fmt.Sprintf("SELECT * FROM %s(%s)", k, strargs)
		}

	}

	for k, v := range l {
		if err := db.prepare(ctx, k, v); err != nil {
			return fmt.Errorf("Could not prepare statement %s: %s", k, err)
		}
	}
	return nil
}

func (db *Database) Exit() {
	for k, v := range db.stmt {
		v.Close()
		delete(db.stmt, k)
	}
}

func (db *Database) InsertRule(ctx context.Context, r jsonapi.Rule) (*uuid.UUID, error) {
	srh := []string{}
	for _, ip := range r.Action.SRH {
		srh = append(srh, ip.String())
	}
	switch r.Type {
	case "uplink":
		if stmt, ok := db.stmt["insert_uplink_rule"]; ok {
			var id uuid.UUID
			err := stmt.QueryRowContext(ctx, r.Enabled, r.Match.UEIpPrefix.String(), r.Match.GNBIpPrefix.String(), r.Action.NextHop.String(), pq.Array(srh)).Scan(&id)
			return &id, err
		} else {
			return nil, fmt.Errorf("Procedure not registered")
		}
	case "downlink":
		if stmt, ok := db.stmt["insert_downlink_rule"]; ok {
			var id uuid.UUID
			err := stmt.QueryRowContext(ctx, r.Enabled, r.Match.UEIpPrefix.String(), r.Action.NextHop.String(), pq.Array(srh)).Scan(&id)
			return &id, err
		} else {
			return nil, fmt.Errorf("Procedure not registered")
		}
	default:
		return nil, fmt.Errorf("Wrong type for the rule")
	}
}

func (db *Database) GetRule(ctx context.Context, uuid uuid.UUID) (jsonapi.Rule, error) {
	var type_uplink bool
	var enabled bool
	var action_next_hop string
	var action_srh []string
	var match_ue_ip_prefix string
	var match_gnb_ip_prefix string
	if stmt, ok := db.stmt["get_rule"]; ok {
		err := stmt.QueryRowContext(ctx, uuid.String()).Scan(&type_uplink, &enabled, &action_next_hop, pq.Array(&action_srh), &match_ue_ip_prefix, &match_gnb_ip_prefix)
		if err != nil {
			return jsonapi.Rule{}, err
		}
		var t string
		if type_uplink {
			t = "uplink"
		} else {
			t = "downlink"
		}
		rule := jsonapi.Rule{
			Enabled: enabled,
			Type:    t,
		}
		rule.Match = jsonapi.Match{}
		if match_ue_ip_prefix != "" {
			p, err := netip.ParsePrefix(match_ue_ip_prefix)
			if err == nil {
				rule.Match.UEIpPrefix = p
			}
		}
		if match_gnb_ip_prefix != "" {
			p, err := netip.ParsePrefix(match_gnb_ip_prefix)
			if err == nil {
				rule.Match.GNBIpPrefix = p
			}
		}

		srh, err := jsonapi.NewSRH(action_srh)
		if err != nil {
			return jsonapi.Rule{}, err
		}
		nh, err := jsonapi.NewNextHop(action_next_hop)
		if err != nil {
			return jsonapi.Rule{}, err
		}

		rule.Action = jsonapi.Action{
			NextHop: *nh,
			SRH:     *srh,
		}

		return rule, err
	} else {
		return jsonapi.Rule{}, fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) GetRules(ctx context.Context) (jsonapi.RuleMap, error) {
	var uuid uuid.UUID
	var type_uplink bool
	var enabled bool
	var action_next_hop string
	var action_srh []string
	var match_ue_ip_prefix string
	var match_gnb_ip_prefix *string
	m := jsonapi.RuleMap{}
	if stmt, ok := db.stmt["get_all_rules"]; ok {
		rows, err := stmt.QueryContext(ctx)
		if err != nil {
			return m, nil
		}
		for rows.Next() {
			select {
			case <-ctx.Done():
				// avoid looping if no longer necessary
				return jsonapi.RuleMap{}, ctx.Err()
			default:
				err := rows.Scan(&uuid, &type_uplink, &enabled, &action_next_hop, pq.Array(&action_srh), &match_ue_ip_prefix, &match_gnb_ip_prefix)
				if err != nil {
					return m, err
				}
				var t string
				if type_uplink {
					t = "uplink"
				} else {
					t = "downlink"
				}
				rule := jsonapi.Rule{
					Enabled: enabled,
					Type:    t,
				}
				rule.Match = jsonapi.Match{}
				if match_ue_ip_prefix != "" {
					p, err := netip.ParsePrefix(match_ue_ip_prefix)
					if err == nil {
						rule.Match.UEIpPrefix = p
					}
				}
				if match_gnb_ip_prefix != nil {
					p, err := netip.ParsePrefix(*match_gnb_ip_prefix)
					if err == nil {
						rule.Match.GNBIpPrefix = p
					}
				}

				srh, err := jsonapi.NewSRH(action_srh)
				if err != nil {
					return jsonapi.RuleMap{}, err
				}
				nh, err := jsonapi.NewNextHop(action_next_hop)
				if err != nil {
					return jsonapi.RuleMap{}, err
				}

				rule.Action = jsonapi.Action{
					NextHop: *nh,
					SRH:     *srh,
				}
				m[uuid] = rule
			}
		}
		return m, nil

	} else {
		return jsonapi.RuleMap{}, fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) EnableRule(ctx context.Context, uuid uuid.UUID) error {
	if stmt, ok := db.stmt["enable_rule"]; ok {
		_, err := stmt.ExecContext(ctx, uuid.String())
		return err
	} else {
		return fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) DisableRule(ctx context.Context, uuid uuid.UUID) error {
	if stmt, ok := db.stmt["disable_rule"]; ok {
		_, err := stmt.ExecContext(ctx, uuid.String())
		return err
	} else {
		return fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) DeleteRule(ctx context.Context, uuid uuid.UUID) error {
	if stmt, ok := db.stmt["delete_rule"]; ok {
		_, err := stmt.ExecContext(ctx, uuid.String())
		return err
	} else {
		return fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) GetUplinkAction(ctx context.Context, uplinkTeid uint32, srgwIp netip.Addr, gnbIp netip.Addr) (jsonapi.Action, error) {
	var action_next_hop jsonapi.NextHop
	var action_srh []string
	if stmt, ok := db.stmt["get_uplink_action"]; ok {
		err := stmt.QueryRowContext(ctx, uplinkTeid, srgwIp.String(), gnbIp.String()).Scan(&action_next_hop, pq.Array(&action_srh))
		if err != nil {
			return jsonapi.Action{}, err
		}
		srh, err := jsonapi.NewSRH(action_srh)
		if err != nil {
			return jsonapi.Action{}, err
		}
		return jsonapi.Action{NextHop: action_next_hop, SRH: *srh}, err
	} else {
		return jsonapi.Action{}, fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) GetDownlinkAction(ctx context.Context, ueIp netip.Addr) (jsonapi.Action, error) {
	var action_next_hop string
	var action_srh []string
	if stmt, ok := db.stmt["get_downlink_action"]; ok {
		err := stmt.QueryRowContext(ctx, ueIp.String()).Scan(&action_next_hop, pq.Array(&action_srh))
		if err != nil {
			return jsonapi.Action{}, err
		}
		srh, err := jsonapi.NewSRH(action_srh)
		if err != nil {
			return jsonapi.Action{}, err
		}
		action, err := jsonapi.NewNextHop(action_next_hop)
		if err != nil {
			return jsonapi.Action{}, err
		}
		return jsonapi.Action{NextHop: *action, SRH: *srh}, err
	} else {
		return jsonapi.Action{}, fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) SetUplinkAction(ctx context.Context, uplinkTeid uint32, srgwIp netip.Addr, gnbIp netip.Addr, ueIp netip.Addr) (jsonapi.Action, error) {
	var action_next_hop string
	var action_srh []string
	if stmt, ok := db.stmt["set_uplink_action"]; ok {
		err := stmt.QueryRowContext(ctx, uplinkTeid, srgwIp.String(), gnbIp.String(), ueIp.String()).Scan(&action_next_hop, pq.Array(&action_srh))
		if err != nil {
			return jsonapi.Action{}, err
		}
		srh, err := jsonapi.NewSRH(action_srh)
		if err != nil {
			return jsonapi.Action{}, err
		}
		action, err := jsonapi.NewNextHop(action_next_hop)
		if err != nil {
			return jsonapi.Action{}, err
		}
		return jsonapi.Action{NextHop: *action, SRH: *srh}, err
	} else {
		return jsonapi.Action{}, fmt.Errorf("Procedure not registered")
	}
}
