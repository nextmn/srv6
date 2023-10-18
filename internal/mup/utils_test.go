// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFromSlice(t *testing.T) {
	res, err := fromSlice([]byte{0xFF, 192, 168, 0, 1}, 8, 4)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, []byte{192, 168, 0, 1}); diff != "" {
		t.Error(diff)
	}
	res, err = fromSlice([]byte{0xFF}, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, []byte{0xFE}); diff != "" {
		t.Error(diff)
	}
	res, err = fromSlice([]byte{0xFF, 0x55}, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res, []byte{0xFD, 0x54}); diff != "" {
		t.Error(diff)
	}
}

func TestAppendToSlice(t *testing.T) {
	b1 := []byte{0xFF, 0x00, 0x00, 0x00}
	if err := appendToSlice(b1, 8, []byte{0x00, 0xAA}); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(b1, []byte{0xFF, 0x00, 0xAA, 0x00}); diff != "" {
		t.Error(diff)
	}
	b2 := []byte{0xE0, 0x00, 0x00, 0x00}
	if err := appendToSlice(b2, 3, []byte{0x00, 0xAA, 0xFF}); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(b2, []byte{0xE0, 0x15, 0x5F, 0xE0}); diff != "" {
		t.Error(diff)
	}
}
