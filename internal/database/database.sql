-- Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
-- Use of this source code is governed by a MIT-style license that can be
-- SPDX-License-Identifier: MIT
CREATE TABLE IF NOT EXISTS uplink_gtp4 (
	uplink_teid INTEGER,
	srgw_ip INET,
	gnb_ip INET,
	action_uuid UUID NOT NULL,
	PRIMARY KEY (uplink_teid, srgw_ip, gnb_ip)
);

CREATE TABLE IF NOT EXISTS rule (
	uuid UUID PRIMARY KEY,
	type_uplink BOOL NOT NULL,
	enabled BOOL NOT NULL,
	action_next_hop INET NOT NULL,
	action_srh INET ARRAY NOT NULL,
	match_ue_ip_prefix CIDR NOT NULL,
	match_gnb_ip_prefix CIDR
);

CREATE OR REPLACE PROCEDURE insert_uplink_rule(
	IN uuid UUID, IN enabled BOOL, IN ue_ip_prefix CIDR,
	IN gnb_ip_prefix CIDR, IN next_hop INET, IN srh INET ARRAY
)
LANGUAGE plpgsql AS $$
BEGIN
	INSERT INTO rule(uuid, type_uplink, enabled, match_ue_ip_prefix, match_gnb_ip_prefix, action_next_hop, action_srh)
		VALUES(uuid, TRUE, enabled, ue_ip_prefix, gnb_ip_prefix, next_hop, srh);
END;$$;

CREATE OR REPLACE PROCEDURE insert_downlink_rule(
	IN uuid UUID, IN enabled BOOL, IN ue_ip_prefix CIDR,
	IN next_hop INET, IN srh INET ARRAY
)
LANGUAGE plpgsql AS $$
BEGIN
	INSERT INTO rule(uuid, type_uplink, enabled, match_ue_ip_prefix, action_next_hop, action_srh)
		VALUES(uuid, FALSE, enabled, ue_ip_prefix, next_hop, srh);
END;$$;

CREATE OR REPLACE PROCEDURE insert_action(
	IN uplink_teid INTEGER, IN srgw_ip INET, IN gnb_ip INET, IN action_uuid UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	INSERT INTO uplink_gtp4 (uplink_teid, srgw_ip, gnb_ip, action_uuid)
		VALUES(uplink_teid, srgw_ip, gnb_ip, action_uuid);
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
	UPDATE rule SET enabled = false WHERE uuid = rule_id
END;$$;

CREATE OR REPLACE PROCEDURE delete_rule(
	IN rule_id UUID
)
LANGUAGE plpgsql AS $$
BEGIN
	DELETE FROM rule WHERE uuid = rule_id
END;$$;
