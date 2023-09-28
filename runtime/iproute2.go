// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

func ipRoute2Init() error {
	if err := createRtTable(); err != nil {
		return err
	}
	if err := createRtProto(); err != nil {
		return err
	}
	if err := addIPRules(); err != nil {
		return err
	}
	if err := addDefaultRoute(); err != nil {
		return err
	}
	return nil
}

func ipRoute2Exit() error {
	if err := removeDefaultRoute(); err != nil {
		return err
	}
	if err := removeIPRules(); err != nil {
		return err
	}
	if err := removeRtProto(); err != nil {
		return err
	}
	if err := removeRtTable(); err != nil {
		return err
	}
	return nil
}

func addIPRules() error {
	return runIP("-6", "rule", "add", "to", SRv6.Locator, "lookup", RTTableName, "protocol", RTProtoName)
}

func addDefaultRoute() error {
	// This default route will be replaced later with a route to sr TUN interface
	return runIP("-6", "route", "add", "blackhole", "default", "table", RTTableName, "proto", RTProtoName)
}

func removeDefaultRoute() error {
	return runIP("-6", "route", "del", "blackhole", "default", "table", RTTableName, "proto", RTProtoName)
}

func removeIPRules() error {
	return runIP("-6", "rule", "del", "to", SRv6.Locator, "lookup", RTTableName, "protocol", RTProtoName)
}
