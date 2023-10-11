// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"reflect"
	"testing"
)

func TestFromSlice(t *testing.T) {
	res, err := fromSlice([]byte{0xFF, 192, 168, 0, 1}, 8, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res, []byte{192, 168, 0, 1}) {
		t.Fatalf("Failed multiple of 8 extraction")
	}
	res, err = fromSlice([]byte{0xFF}, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res, []byte{0xFE}) {
		t.Fatalf("Failed shift 1, %x", res[0])
	}
	res, err = fromSlice([]byte{0xFF, 0x55}, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res, []byte{0xFD, 0x54}) {
		t.Fatalf("Failed shift 2 %x, %x", res[0], res[1])
	}
}
