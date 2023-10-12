// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

// slice: input slice
// startBit: offset in bits
// length: length of result in Bytes
func fromSlice(slice []byte, startBit uint, length uint) ([]byte, error) {
	if uint(len(slice)) < length {
		return nil, ErrTooShortToParse
	}
	startByte := startBit / 8
	offset := startBit % 8
	ret := make([]byte, length)
	if offset == 0 {
		copy(ret, slice[startByte:startByte+length])
		return ret, nil
	}

	// init left
	for i, b := range slice[startByte : startByte+length] {
		ret[i] = (b << offset)
	}
	// init right
	for i, b := range slice[startByte+1 : startByte+length] {
		ret[i] |= b >> (8 - offset)
	}
	return ret, nil
}
