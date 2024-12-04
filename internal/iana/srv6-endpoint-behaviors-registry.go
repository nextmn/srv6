// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package iana

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type EndpointBehavior uint32

// Full registry is available at https://www.iana.org/assignments/segment-routing/segment-routing.xhtml
// Only implemented behaviors are defined in this file
const (
	NotToBeAllocated EndpointBehavior = 0x0000
	End              EndpointBehavior = 0x0001
	End_DX4          EndpointBehavior = 0x0011
	End_MAP          EndpointBehavior = 0x0028
	End_Limit        EndpointBehavior = 0x0029
	End_M_GTP6_D     EndpointBehavior = 0x0045
	End_M_GTP6_Di    EndpointBehavior = 0x0046
	End_M_GTP6_E     EndpointBehavior = 0x0047
	End_M_GTP4_E     EndpointBehavior = 0x0048
)

// Convert a string to an EndpointBehavior
func ToEndpointBehavior(s string) (EndpointBehavior, error) {
	switch strings.ToLower(s) {
	case "end":
		return End, nil
	case "end.dx4":
		return End_DX4, nil
	case "end.map":
		return End_MAP, nil
	case "end.limit":
		return End_Limit, nil
	case "end.m.gtp6.d":
		return End_M_GTP6_D, nil
	case "end.m.gtp6.di", "end.m.gtp6.d.di": // Di in iana registry, but D.Di in RFC9433
		return End_M_GTP6_Di, nil
	case "end.m.gtp6.e":
		return End_M_GTP6_E, nil
	case "end.m.gtp4.e":
		return End_M_GTP4_E, nil
	default:
		return NotToBeAllocated, fmt.Errorf("The value %s cannot be converted to EndpointBehavior. It may not be implemented, or contain a typo.", s)
	}
}

func (e EndpointBehavior) String() string {
	switch e {
	case End:
		return "End"
	case End_DX4:
		return "End.DX4"
	case End_MAP:
		return "End.MAP"
	case End_Limit:
		return "End.Limit"
	case End_M_GTP6_D:
		return "End.M.GTP6.D"
	case End_M_GTP6_Di:
		return "End.M.GTP6.Di"
	case End_M_GTP6_E:
		return "End.M.GTP6.E"
	case End_M_GTP4_E:
		return "End.M.GTP4.E"
	default:
		return "Unknown behavior"
	}
}

func (e *EndpointBehavior) ToIPRoute2Action() (string, error) {
	switch *e {
	case End:
		return "End", nil
	case End_DX4:
		return "End.DX4", nil
	default:
		return "", fmt.Errorf("Not implemented")
	}
}

// Unmarshal YAML to EndpointBehavior
func (e *EndpointBehavior) UnmarshalYAML(n *yaml.Node) error {
	eb, err := ToEndpointBehavior(n.Value)
	if err != nil {
		return err
	}
	*e = eb
	return nil
}
