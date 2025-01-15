// Code generated by gen.go; DO NOT EDIT.

// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package database

type procedureOrFunction struct {
	num_in       int
	num_out      int
	is_procedure bool
}

var procedures = map[string]procedureOrFunction{
	"insert_uplink_rule":   {is_procedure: true, num_in: 7, num_out: 1},
	"insert_downlink_rule": {is_procedure: true, num_in: 3, num_out: 1},
	"enable_rule":          {is_procedure: true, num_in: 1, num_out: 0},
	"disable_rule":         {is_procedure: true, num_in: 1, num_out: 0},
	"switch_rule":          {is_procedure: true, num_in: 2, num_out: 0},
	"delete_rule":          {is_procedure: true, num_in: 1, num_out: 0},
	"update_action":        {is_procedure: true, num_in: 2, num_out: 0},
	"get_uplink_action":    {is_procedure: false, num_in: 5, num_out: 0},
	"get_downlink_action":  {is_procedure: false, num_in: 1, num_out: 0},
	"get_rule":             {is_procedure: false, num_in: 1, num_out: 0},
	"get_all_rules":        {is_procedure: false, num_in: 0, num_out: 0},
}
