// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package mup

import "encoding/binary"

// Args.Mob.Session as defined in RFC 9433, section 6.1:
//
//	 0                   1                   2                   3
//	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|   QFI     |R|U|                PDU Session ID                 |
//	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	|PDU Sess(cont')|
//	+-+-+-+-+-+-+-+-+
//	Figure 8: Args.Mob.Session Format
type ArgsMobSession struct {
	qfi          uint8  // QoS Flow Identifier (6 bits)
	r            uint8  // Reflective QoS Indication (1 bit)
	u            uint8  // Unused and for future use (1 bit)
	pduSessionID uint32 // Identifier of PDU Session. The GTP-U equivalent is TEID (32 bits)
}

// NewArgsMobSession creates an ArgsMobSession.
func NewArgsMobSession(qfi uint8, r bool, u bool, pduSessionID uint32) *ArgsMobSession {
	var ruint uint8 = 0
	if r {
		ruint = 1
	}
	var uuint uint8 = 0
	if u {
		uuint = 1
	}
	return &ArgsMobSession{
		qfi:          qfi,
		r:            ruint,
		u:            uuint,
		pduSessionID: pduSessionID,
	}
}

// ParseArgsMobSession parses given byte sequence as an ArgsMobSession.
func ParseArgsMobSession(b []byte) (*ArgsMobSession, error) {
	a := &ArgsMobSession{}
	if err := a.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return a, nil
}

// QFI returns the Qos Flow Identifier for this ArgsMobSession.
func (a *ArgsMobSession) QFI() uint8 {
	return a.qfi
}

// R returns the Reflective QoS Indication for this ArgsMobSession.
func (a *ArgsMobSession) R() bool {
	if a.r == 0 {
		return false
	}
	return true
}

// U returns the U bit for this ArgsMobSession.
func (a *ArgsMobSession) U() bool {
	if a.u == 0 {
		return false
	}
	return true
}

// PDUSessionID returns the PDU Session Identifier for this ArgsMobSession. The GTP-U equivalent is TEID.
func (a *ArgsMobSession) PDUSessionID() uint32 {
	return a.pduSessionID
}

// MarshalLen returns the serial length of ArgsMobSession.
func (a *ArgsMobSession) MarshalLen() int {
	return ARGS_MOB_SESSION_SIZE_BYTE
}

// Marshal returns the byte sequence generated from ArgsMobSession.
func (a *ArgsMobSession) Marshal() ([]byte, error) {
	b := make([]byte, a.MarshalLen())
	if err := a.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (a *ArgsMobSession) MarshalTo(b []byte) error {
	if len(b) < a.MarshalLen() {
		return ErrTooShortToMarshal
	}
	b[QFI_POS_BYTE] |= (QFI_MASK & a.qfi) << QFI_POS_BIT
	b[R_POS_BYTE] |= (R_MASK & a.r) << R_POS_BIT
	b[U_POS_BYTE] |= (U_MASK & a.u) << U_POS_BIT
	binary.BigEndian.PutUint32(b[TEID_POS_BYTE:TEID_POS_BYTE+TEID_SIZE_BYTE], a.pduSessionID)
	return nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in an ArgsMobSession.
func (a *ArgsMobSession) UnmarshalBinary(b []byte) error {
	if len(b) < ARGS_MOB_SESSION_SIZE_BYTE {
		return ErrTooShortToParse
	}
	a.qfi = QFI_MASK & (b[QFI_POS_BYTE] >> QFI_POS_BIT)
	a.r = R_MASK & (b[R_POS_BYTE] >> R_POS_BIT)
	a.u = U_MASK & (b[U_POS_BYTE] >> U_POS_BIT)
	a.pduSessionID = binary.BigEndian.Uint32(b[TEID_POS_BYTE : TEID_POS_BYTE+TEID_SIZE_BYTE])
	return nil
}
