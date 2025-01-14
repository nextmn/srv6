// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
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

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n4tosrv6"

	"github.com/gofrs/uuid"
	"github.com/lib/pq"
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
		for range v.num_out {
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

func (db *Database) InsertRule(ctx context.Context, r n4tosrv6.Rule) (*uuid.UUID, error) {
	srh := []string{}
	for _, ip := range r.Action.SRH {
		srh = append(srh, ip.String())
	}
	switch r.Type {
	case "uplink":
		var inneripsrc string
		var inneripdst string
		var outeripsrc []string
		if r.Match.Header.InnerIpSrc == nil {
			inneripsrc = "0.0.0.0/0"
		} else {
			inneripsrc = r.Match.Header.InnerIpSrc.String() + "/32"
		}
		if r.Match.Payload == nil {
			inneripdst = "0.0.0.0/0"
		} else {
			inneripdst = r.Match.Payload.Dst.String() + "/32"
		}
		for _, i := range r.Match.Header.OuterIpSrc {
			outeripsrc = append(outeripsrc, i.String())
		}

		if stmt, ok := db.stmt["insert_uplink_rule"]; ok {
			var id uuid.UUID
			err := stmt.QueryRowContext(ctx, r.Enabled, inneripsrc, pq.Array(outeripsrc), r.Match.Header.FTeid.Teid, r.Match.Header.FTeid.Addr.String(), inneripdst, pq.Array(srh)).Scan(&id)
			return &id, err
		} else {
			return nil, fmt.Errorf("Procedure not registered")
		}
	case "downlink":
		if stmt, ok := db.stmt["insert_downlink_rule"]; ok {
			var id uuid.UUID
			var dst string
			if r.Match.Payload == nil {
				dst = "0.0.0.0/0"
			} else {
				dst = r.Match.Payload.Dst.String() + "/32"
			}
			err := stmt.QueryRowContext(ctx, r.Enabled, dst, pq.Array(srh)).Scan(&id)
			return &id, err
		} else {
			return nil, fmt.Errorf("Procedure not registered")
		}
	default:
		return nil, fmt.Errorf("Wrong type for the rule")
	}
}

func (db *Database) GetRule(ctx context.Context, uuid uuid.UUID) (n4tosrv6.Rule, error) {
	var type_uplink bool
	var enabled bool
	var action_srh []string
	var match_ue_ip string
	var match_gnb_ip []string
	var match_service_ip *string
	var match_uplink_teid *uint32
	var match_uplink_upf *string
	if stmt, ok := db.stmt["get_rule"]; ok {
		err := stmt.QueryRowContext(ctx, uuid.String()).Scan(&type_uplink, &enabled, pq.Array(&action_srh), &match_ue_ip, pq.Array(&match_gnb_ip), &match_uplink_teid, &match_uplink_upf, &match_service_ip)
		if err != nil {
			return n4tosrv6.Rule{}, err
		}
		rule := n4tosrv6.Rule{
			Enabled: enabled,
			Match:   n4tosrv6.Match{},
		}
		if type_uplink {
			rule.Type = "uplink"
			rule.Match.Header = &n4tosrv6.GtpHeader{}

			rule.Match.Header.OuterIpSrc = make([]netip.Prefix, 0)
			for _, i := range match_gnb_ip {
				p, err := netip.ParsePrefix(i)
				if err != nil {
					return n4tosrv6.Rule{}, err
				}
				rule.Match.Header.OuterIpSrc = append(rule.Match.Header.OuterIpSrc, p)
			}
			if match_uplink_upf != nil && match_uplink_teid != nil {
				addr, err := netip.ParseAddr(*match_uplink_upf)
				if err != nil {
					return n4tosrv6.Rule{}, err
				}
				rule.Match.Header.FTeid = jsonapi.Fteid{
					Teid: *match_uplink_teid,
					Addr: addr,
				}
			}
			if match_service_ip != nil {
				p, err := netip.ParsePrefix(*match_service_ip)
				if err == nil && p.Bits() == 32 {
					rule.Match.Payload = &n4tosrv6.Payload{
						Dst: p.Addr(),
					}
				}
			}
		} else {
			rule.Type = "downlink"
		}
		p, err := netip.ParsePrefix(match_ue_ip)
		if err == nil && p.Bits() == 32 {
			if type_uplink {
				a := p.Addr()
				rule.Match.Header.InnerIpSrc = &a
			} else {
				rule.Match.Payload = &n4tosrv6.Payload{
					Dst: p.Addr(),
				}
			}
		}
		srh, err := n4tosrv6.NewSRH(action_srh)
		if err != nil {
			return n4tosrv6.Rule{}, err
		}

		rule.Action = n4tosrv6.Action{
			SRH: *srh,
		}

		return rule, err
	} else {
		return n4tosrv6.Rule{}, fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) GetRules(ctx context.Context) (n4tosrv6.RuleMap, error) {
	var uuid uuid.UUID
	var type_uplink bool
	var enabled bool
	var action_srh []string
	var match_ue_ip string
	var match_gnb_ip []string
	var match_uplink_teid *uint32
	var match_uplink_upf *string
	var match_service_ip *string
	m := n4tosrv6.RuleMap{}
	if stmt, ok := db.stmt["get_all_rules"]; ok {
		rows, err := stmt.QueryContext(ctx)
		if err != nil {
			return m, nil
		}
		for rows.Next() {
			select {
			case <-ctx.Done():
				// avoid looping if no longer necessary
				return n4tosrv6.RuleMap{}, ctx.Err()
			default:
				err := rows.Scan(&uuid, &type_uplink, &enabled, pq.Array(&action_srh), &match_ue_ip, pq.Array(&match_gnb_ip), &match_uplink_teid, &match_uplink_upf, &match_service_ip)
				if err != nil {
					return m, err
				}
				rule := n4tosrv6.Rule{
					Enabled: enabled,
					Match:   n4tosrv6.Match{},
				}
				if type_uplink {
					rule.Type = "uplink"
					rule.Match.Header = &n4tosrv6.GtpHeader{}
					rule.Match.Header.OuterIpSrc = make([]netip.Prefix, 0)
					for _, i := range match_gnb_ip {
						p, err := netip.ParsePrefix(i)
						if err != nil {
							return n4tosrv6.RuleMap{}, err
						}
						rule.Match.Header.OuterIpSrc = append(rule.Match.Header.OuterIpSrc, p)
					}
					if match_uplink_upf != nil && match_uplink_teid != nil {
						addr, err := netip.ParseAddr(*match_uplink_upf)
						if err != nil {
							return n4tosrv6.RuleMap{}, err
						}
						rule.Match.Header.FTeid = jsonapi.Fteid{
							Teid: *match_uplink_teid,
							Addr: addr,
						}
					}
					if match_service_ip != nil {
						p, err := netip.ParsePrefix(*match_service_ip)
						if err == nil && p.Bits() == 32 {
							rule.Match.Payload = &n4tosrv6.Payload{
								Dst: p.Addr(),
							}
						}
					}
				} else {
					rule.Type = "downlink"
				}
				p, err := netip.ParsePrefix(match_ue_ip)
				if err == nil && p.Bits() == 32 {
					if type_uplink {
						a := p.Addr()
						rule.Match.Header.InnerIpSrc = &a
					} else {
						rule.Match.Payload = &n4tosrv6.Payload{
							Dst: p.Addr(),
						}
					}
				}

				srh, err := n4tosrv6.NewSRH(action_srh)
				if err != nil {
					return n4tosrv6.RuleMap{}, err
				}

				rule.Action = n4tosrv6.Action{
					SRH: *srh,
				}
				m[uuid] = rule
			}
		}
		return m, nil

	} else {
		return n4tosrv6.RuleMap{}, fmt.Errorf("Procedure not registered")
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

func (db *Database) SwitchRule(ctx context.Context, uuidEnable uuid.UUID, uuidDisable uuid.UUID) error {
	if stmt, ok := db.stmt["switch_rule"]; ok {
		_, err := stmt.ExecContext(ctx, uuidEnable.String(), uuidDisable.String())
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

func (db *Database) GetUplinkAction(ctx context.Context, uplinkFTeid jsonapi.Fteid, gnbIp netip.Addr, ueIp netip.Addr, serviceIp netip.Addr) (n4tosrv6.Action, error) {
	var action_srh []string
	if stmt, ok := db.stmt["get_uplink_action"]; ok {
		err := stmt.QueryRowContext(ctx, uplinkFTeid.Teid, uplinkFTeid.Addr.String(), gnbIp.String(), ueIp.String(), serviceIp.String()).Scan(pq.Array(&action_srh))
		if err != nil {
			return n4tosrv6.Action{}, err
		}
		srh, err := n4tosrv6.NewSRH(action_srh)
		if err != nil {
			return n4tosrv6.Action{}, err
		}
		return n4tosrv6.Action{SRH: *srh}, err
	} else {
		return n4tosrv6.Action{}, fmt.Errorf("Procedure not registered")
	}
}

func (db *Database) GetDownlinkAction(ctx context.Context, ueIp netip.Addr) (n4tosrv6.Action, error) {
	var action_srh []string
	if stmt, ok := db.stmt["get_downlink_action"]; ok {
		err := stmt.QueryRowContext(ctx, ueIp.String()).Scan(pq.Array(&action_srh))
		if err != nil {
			return n4tosrv6.Action{}, err
		}
		srh, err := n4tosrv6.NewSRH(action_srh)
		if err != nil {
			return n4tosrv6.Action{}, err
		}
		return n4tosrv6.Action{SRH: *srh}, err
	} else {
		return n4tosrv6.Action{}, fmt.Errorf("Procedure not registered")
	}
}
