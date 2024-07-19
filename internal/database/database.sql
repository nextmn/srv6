-- Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
-- Use of this source code is governed by a MIT-style license that can be
-- SPDX-License-Identifier: MIT
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS rule (
	uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	type_uplink BOOL NOT NULL,
	enabled BOOL NOT NULL,
	action_next_hop INET NOT NULL,
	action_srh INET ARRAY NOT NULL,
	match_ue_ip_prefix CIDR NOT NULL,
	match_gnb_ip_prefix CIDR
);

CREATE TABLE IF NOT EXISTS uplink_gtp4 (
	uplink_teid INTEGER,
	srgw_ip INET,
	gnb_ip INET,
	action_uuid UUID REFERENCES rule (uuid) ON DELETE CASCADE,
	PRIMARY KEY (uplink_teid, srgw_ip, gnb_ip)
);

CREATE OR REPLACE PROCEDURE insert_uplink_rule(
	IN enabled BOOL, IN ue_ip_prefix CIDR,
	IN gnb_ip_prefix CIDR, IN next_hop INET, IN srh INET ARRAY,
	OUT uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	INSERT INTO rule(type_uplink, enabled, match_ue_ip_prefix, match_gnb_ip_prefix, action_next_hop, action_srh)
		VALUES(TRUE, enabled, ue_ip_prefix, gnb_ip_prefix, next_hop, srh) RETURNING uuid INTO uuid;
END;$$;

CREATE OR REPLACE PROCEDURE insert_downlink_rule(
	IN enabled BOOL, IN ue_ip_prefix CIDR,
	IN next_hop INET, IN srh INET ARRAY,
	OUT uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	INSERT INTO rule(type_uplink, enabled, match_ue_ip_prefix, action_next_hop, action_srh)
		VALUES(FALSE, enabled, ue_ip_prefix, next_hop, srh) RETURNING uuid INTO uuid;
END;$$;


CREATE OR REPLACE PROCEDURE enable_rule(
	IN rule_id UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	UPDATE rule SET enabled = true WHERE uuid = rule_id;
END;$$;

CREATE OR REPLACE PROCEDURE disable_rule(
	IN rule_id UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	UPDATE rule SET enabled = false WHERE uuid = rule_id;
	DELETE FROM uplink_gtp4 WHERE action_uuid = rule_id;
END;$$;

CREATE OR REPLACE PROCEDURE delete_rule(
	IN rule_id UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	DELETE FROM rule WHERE uuid = rule_id;
END;$$;

CREATE OR REPLACE PROCEDURE get_uplink_action(
	IN uplink_teid INTEGER, IN srgw_ip INET, IN gnb_ip INET,
	OUT action_next_hop INET, OUT action_srh INET ARRAY
)
LANGUAGE plpgsql AS $$
BEGIN
	SELECT rule.action_next_hop, rule.action_srh FROM uplink_gtp4, rule
		WHERE (uplink_gtp4.uplink_teid = uplink_teid
			AND uplink_gtp4.srgw_ip = srgw_ip
			AND uplink_gtp4.gnb_ip = gnb_ip
			AND rule.uuid = uplink_gtp4.action_uuid)
		INTO action_next_hop, action_srh;
END;$$;

CREATE OR REPLACE PROCEDURE set_uplink_action(
	IN uplink_teid INTEGER, IN srgw_ip INET, gnb_ip INET, IN ue_ip_address INET,
	OUT action_next_hop INET, OUT action_srh INET ARRAY
)
LANGUAGE plpgsql AS $$
DECLARE
	action_uuid UUID;
BEGIN
	SELECT uuid, action_next_hop, action_srh FROM rule
		WHERE (type_uplink = TRUE AND enabled = TRUE
			AND gnb_ip << match_gnb_ip_prefix AND ue_ip << match_ue_ip_prefix)
		INTO action_uuid, action_next_hop, action_srh;
	INSERT INTO uplink_gtp4(uplink_teid, srgw_ip, gnb_ip, action_uuid)
		VALUES(uplink_teid, srgw_ip, gnb_ip, action_uuid);
END;$$;

CREATE OR REPLACE PROCEDURE get_downlink_action(
	IN ue_ip_address INET,
	OUT action_next_hop INET, OUT action_srh INET ARRAY
)
LANGUAGE plpgsql AS $$
BEGIN
	SELECT rule.action_next_hop, rule.action_srh FROM rule
		WHERE (type_uplink = FALSE AND enabled = TRUE
			AND ue_ip << match_ue_ip_prefix)
		INTO action_next_hop, action_srh;
END;$$;

CREATE OR REPLACE PROCEDURE get_rule(
	IN uuid UUID,
	OUT type_uplink BOOL,
	OUT enabled BOOL,
	OUT action_next_hop INET,
	OUT action_srh INET ARRAY,
	OUT match_ue_ip_prefix CIDR,
	OUT match_gnb_ip_prefix CIDR
)
LANGUAGE plpgsql AS $$
BEGIN
	SELECT type_uplink, enabled, action_next_hop,
		action_srh, match_ue_ip_prefix, match_gnb_ip_prefix
		FROM rule
		WHERE (rule.uuid = uuid)
		INTO type_uplink, enabled, action_next_hop, action_srh,
			match_ue_ip_prefix, match_gnb_ip_prefix;
END;$$;

CREATE OR REPLACE PROCEDURE get_all_rules(
	OUT uuid UUID,
	OUT type_uplink BOOL,
	OUT enabled BOOL,
	OUT action_next_hop INET,
	OUT action_srh INET ARRAY,
	OUT match_ue_ip_prefix CIDR,
	OUT match_gnb_ip_prefix CIDR
)
LANGUAGE plpgsql AS $$
BEGIN
	SELECT uuid, type_uplink, enabled, action_next_hop,
		action_srh, match_ue_ip_prefix, match_gnb_ip_prefix
		FROM rule
		INTO uuid, type_uplink, enabled, action_next_hop, action_srh,
			match_ue_ip_prefix, match_gnb_ip_prefix;
END;$$;
