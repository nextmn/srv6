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
	match_gnb_ip_prefix CIDR,
	match_uplink_teid INTEGER
);


CREATE OR REPLACE PROCEDURE insert_uplink_rule(
	IN in_enabled BOOL, IN in_ue_ip_prefix CIDR,
	IN in_gnb_ip_prefix CIDR, IN in_uplink_teid INTEGER,
	IN in_next_hop INET, IN in_srh INET ARRAY,
	OUT out_uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	INSERT INTO rule(type_uplink, enabled, match_ue_ip_prefix, match_gnb_ip_prefix, match_uplink_teid, action_next_hop, action_srh)
		VALUES(TRUE, in_enabled, in_ue_ip_prefix, in_gnb_ip_prefix, in_uplink_teid, in_next_hop, in_srh) RETURNING rule.uuid INTO out_uuid;
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
END;$$;

CREATE OR REPLACE PROCEDURE switch_rule(
	IN in_uuid_enable UUID,
	IN in_uuid_disable UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	UPDATE rule SET enabled = true WHERE rule.uuid = in_uuid_enable;
	UPDATE rule SET enabled = false WHERE rule.uuid = in_uuid_disable;
END;$$;

CREATE OR REPLACE PROCEDURE delete_rule(
	IN in_uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	DELETE FROM rule WHERE uuid = in_uuid;
END;$$;

CREATE OR REPLACE FUNCTION get_uplink_action(
	IN in_uplink_teid INTEGER, IN in_gnb_ip INET
)
RETURNS TABLE (
	t_action_next_hop INET,
	t_action_srh INET ARRAY
)
AS $$
BEGIN
	RETURN QUERY SELECT rule.action_next_hop AS "t_action_next_hop", rule.action_srh AS "t_action_srh"
		FROM rule
		WHERE (rule.match_uplink_teid = in_uplink_teid
			AND rule.match_gnb_ip_prefix && in_gnb_ip);
END;$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_downlink_action(
	IN in_ue_ip_address INET
)
RETURNS TABLE (
	t_action_next_hop INET,
	t_action_srh INET ARRAY
)
AS $$
BEGIN
	RETURN QUERY SELECT rule.action_next_hop AS "t_action_next_hop", rule.action_srh AS "t_action_srh"
		FROM rule
		WHERE (rule.type_uplink = FALSE AND rule.enabled = TRUE
			AND match_ue_ip_prefix && in_ue_ip_address);
END;$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_rule(
	IN in_uuid UUID
)
RETURNS TABLE (
	t_type_uplink BOOL,
	t_enabled BOOL,
	t_action_next_hop INET,
	t_action_srh INET ARRAY,
	t_match_ue_ip_prefix CIDR,
	t_match_gnb_ip_prefix CIDR,
	t_match_uplink_teid INTEGER
)
AS $$
BEGIN
	RETURN QUERY SELECT type_uplink AS "t_type_uplink", enabled AS "t_enabled", action_next_hop AS "t_action_next_hop",
		action_srh AS "t_action_srh", match_ue_ip_prefix AS "t_match_ue_ip_prefix", match_gnb_ip_prefix AS "t_match_gnb_ip_prefix",
		match_uplink_teid AS "t_match_uplink_teid"
		FROM rule
		WHERE (rule.uuid = in_uuid);
END;$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_all_rules()
RETURNS TABLE (
	t_uuid UUID,
	t_type_uplink BOOL,
	t_enabled BOOL,
	t_action_next_hop INET,
	t_action_srh INET ARRAY,
	t_match_ue_ip_prefix CIDR,
	t_match_gnb_ip_prefix CIDR,
	t_match_uplink_teid INTEGER
)
AS $$
BEGIN
	RETURN QUERY SELECT uuid AS "t_uuid", type_uplink AS "t_type_uplink",
		enabled AS "t_enabled", action_next_hop AS "t_action_next_hop",
		action_srh AS "t_action_srh", match_ue_ip_prefix AS "t_match_ue_ip_prefix", match_gnb_ip_prefix AS "t_match_gnb_ip_prefix",
		match_uplink_teid AS "t_match_uplink_teid"
		FROM rule;
END;$$ LANGUAGE plpgsql;
