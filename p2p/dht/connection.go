/*
 *  Copyright (C) 2017 gyee authors
 *
 *  This file is part of the gyee library.
 *
 *  the gyee library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  the gyee library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with the gyee library.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package dht

import (
	sch	"github.com/yeeco/gyee/p2p/scheduler"
)

const ConMgrName = sch.DhtConMgrName

type ConMgr struct {
	name	string				// my name
	tep		sch.SchUserTaskEp	// task entry
	ptnMe	interface{}			// pointer to task node of myself
}

//
// Create route manager
//
func NewConMgr() *ConMgr {

	conMgr := ConMgr{
		name:	ConMgrName,
		tep:	nil,
		ptnMe:	nil,
	}

	conMgr.tep = conMgr.conMgrProc

	return &conMgr
}

//
// Entry point exported to shceduler
//
func (conMgr *ConMgr)TaskProc4Scheduler(ptn interface{}, msg *sch.SchMessage) sch.SchErrno {
	return conMgr.tep(ptn, msg)
}

//
// Connection manager entry
//
func (conMgr *ConMgr)conMgrProc(ptn interface{}, msg *sch.SchMessage) sch.SchErrno {

	eno := sch.SchEnoUnknown

	switch msg.Id {
	case sch.EvSchPoweron:
	case sch.EvSchPoweroff:
	case sch.EvDhtLsnMgrAcceptInd:
	case sch.EvDhtConMgrConnectReq:
	case sch.EvDhtLsnMgrStatusInd:
	case sch.EvDhtConMgrSendReq:
	case sch.EvDhtConMgrCloseReq:
	case sch.EvDhtConInstMsgInd:
	case sch.EvDhtConInstStatusInd:
	case sch.EvDhtConInstCloseRsp:
	default:
		eno = sch.SchEnoParameter
	}

	return eno
}

