// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

//func linuxSRCreateEndpoints() error {
//	for _, endpoint := range SRv6.Endpoints {
//		switch endpoint.Behavior {
//		case "End":
//			if err := runIP("-6", "route", "add", endpoint.Sid, "encap", "seg6local", "action", "End",
//				"dev", LinuxSRLinkName, "table", RTTableName, "proto", RTProtoName); err != nil {
//				return err
//			}
//		case "End.DX4":
//			if err := runIP("-6", "route", "add", endpoint.Sid, "encap", "seg6local", "action", "End.DX4", "nh4", "0.0.0.0",
//				"dev", LinuxSRLinkName, "table", RTTableName, "proto", RTProtoName); err != nil {
//				return err
//			}
//		default:
//		}
//	}
//	return nil
//}
