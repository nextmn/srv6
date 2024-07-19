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
	IN in_enabled BOOL, IN in_ue_ip_prefix CIDR,
	IN in_gnb_ip_prefix CIDR, IN in_next_hop INET, IN in_srh INET ARRAY,
	OUT out_uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	INSERT INTO rule(type_uplink, enabled, match_ue_ip_prefix, match_gnb_ip_prefix, action_next_hop, action_srh)
		VALUES(TRUE, in_enabled, in_ue_ip_prefix, in_gnb_ip_prefix, in_next_hop, in_srh) RETURNING rule.uuid INTO out_uuid;
END;$$;

CREATE OR REPLACE PROCEDURE insert_downlink_rule(
	IN in_enabled BOOL, IN in_ue_ip_prefix CIDR,
	IN in_next_hop INET, IN in_srh INET ARRAY,
	OUT out_uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	INSERT INTO rule(type_uplink, enabled, match_ue_ip_prefix, action_next_hop, action_srh)
		VALUES(FALSE, in_enabled, in_ue_ip_prefix, in_next_hop, in_srh) RETURNING rule.uuid INTO out_uuid;
END;$$;


CREATE OR REPLACE PROCEDURE enable_rule(
	IN in_uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	UPDATE rule SET enabled = true WHERE rule.uuid = in_uuid;
END;$$;

CREATE OR REPLACE PROCEDURE disable_rule(
	IN in_uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	UPDATE rule SET enabled = false WHERE rule.uuid = in_uuid;
	DELETE FROM uplink_gtp4 WHERE uplink_gtp4.action_uuid = in_uuid;
END;$$;

CREATE OR REPLACE PROCEDURE delete_rule(
	IN in_uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	DELETE FROM rule WHERE uuid = in_uuid;
END;$$;

CREATE OR REPLACE PROCEDURE get_uplink_action(
	IN in_uplink_teid INTEGER, IN in_srgw_ip INET, IN in_gnb_ip INET,
	OUT out_action_next_hop INET, OUT out_action_srh INET ARRAY
)
LANGUAGE plpgsql AS $$
BEGIN
	SELECT rule.action_next_hop, rule.action_srh FROM uplink_gtp4, rule
		WHERE (uplink_gtp4.uplink_teid = in_uplink_teid
			AND uplink_gtp4.srgw_ip = in_srgw_ip
			AND uplink_gtp4.gnb_ip = in_gnb_ip
			AND rule.uuid = uplink_gtp4.action_uuid)
		INTO out_action_next_hop, out_action_srh;
END;$$;

CREATE OR REPLACE PROCEDURE set_uplink_action(
	IN in_uplink_teid INTEGER, IN in_srgw_ip INET, IN in_gnb_ip INET, IN in_ue_ip_address INET,
	OUT out_action_next_hop INET, OUT out_action_srh INET ARRAY
)
LANGUAGE plpgsql AS $$
DECLARE
	var_uuid UUID;
BEGIN
	SELECT uuid, action_next_hop, action_srh FROM rule
		WHERE (rule.type_uplink = TRUE AND rule.enabled = TRUE
			AND in_gnb_ip << rule.match_gnb_ip_prefix AND in_ue_ip << rule.match_ue_ip_prefix)
		INTO var_uuid, out_action_next_hop, out_action_srh;
	INSERT INTO uplink_gtp4(uplink_teid, srgw_ip, gnb_ip, var_uuid)
		VALUES(in_uplink_teid, in_srgw_ip, in_gnb_ip, var_uuid);
END;$$;

CREATE OR REPLACE PROCEDURE get_downlink_action(
	IN in_ue_ip_address INET,
	OUT out_action_next_hop INET, OUT out_action_srh INET ARRAY
)
LANGUAGE plpgsql AS $$
BEGIN
	SELECT rule.action_next_hop, rule.action_srh FROM rule
		WHERE (rule.type_uplink = FALSE AND rule.enabled = TRUE
			AND in_ue_ip << match_ue_ip_prefix)
		INTO out_action_next_hop, out_action_srh;
END;$$;

CREATE OR REPLACE PROCEDURE get_rule(
	IN in_uuid UUID,
	OUT out_type_uplink BOOL,
	OUT out_enabled BOOL,
	OUT out_action_next_hop INET,
	OUT out_action_srh INET ARRAY,
	OUT out_match_ue_ip_prefix CIDR,
	OUT out_match_gnb_ip_prefix CIDR
)
LANGUAGE plpgsql AS $$
BEGIN
	SELECT type_uplink, enabled, action_next_hop,
		action_srh, match_ue_ip_prefix, match_gnb_ip_prefix
		FROM rule
		WHERE (rule.uuid = in_uuid)
		INTO out_type_uplink, out_enabled, out_action_next_hop, out_action_srh,
			out_match_ue_ip_prefix, out_match_gnb_ip_prefix;
END;$$;

CREATE OR REPLACE PROCEDURE get_all_rules(
	OUT out_uuid UUID,
	OUT out_type_uplink BOOL,
	OUT out_enabled BOOL,
	OUT out_action_next_hop INET,
	OUT out_action_srh INET ARRAY,
	OUT out_match_ue_ip_prefix CIDR,
	OUT out_match_gnb_ip_prefix CIDR
)
LANGUAGE plpgsql AS $$
BEGIN
	SELECT uuid, type_uplink, enabled, action_next_hop,
		action_srh, match_ue_ip_prefix, match_gnb_ip_prefix
		FROM rule
		INTO out_uuid, out_type_uplink, out_enabled, out_action_next_hop, out_action_srh,
			out_match_ue_ip_prefix, out_match_gnb_ip_prefix;
END;$$;
