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


package peer

import (
	"net"
	"fmt"
	ycfg	"github.com/yeeco/gyee/p2p/config"
	sch		"github.com/yeeco/gyee/p2p/scheduler"
	yclog	"github.com/yeeco/gyee/p2p/logger"
	"sync"
)

//
// Listener manager
//
const PeerLsnMgrName = sch.PeerLsnMgrName

type ListenerManager struct {
	sdl			*sch.Scheduler			// pointer to scheduler
	name		string					// name
	tep			sch.SchUserTaskEp		// entry
	ptn			interface{}				// the listner task node pointer
	ptnPeerMgr	interface{}				// the peer manager task node pointer
	cfg			*ycfg.Cfg4PeerListener	// configuration
	listener	net.Listener			// listener of net
	listenAddr	*net.TCPAddr			// listen address
	accepter	*acceptTskCtrlBlock		// pointer to accepter
}

//
// Create listener manager
//
func NewLsnMgr() *ListenerManager {
	var lsnMgr = ListenerManager {
		name: PeerLsnMgrName,
	}

	lsnMgr.tep = lsnMgr.lsnMgrProc
	return &lsnMgr
}

//
// Entry point exported to shceduler
//
func (lsnMgr *ListenerManager)TaskProc4Scheduler(ptn interface{}, msg *sch.SchMessage) sch.SchErrno {
	return lsnMgr.tep(ptn, msg)
}

//
// Listen manager entry
//
func (lsnMgr *ListenerManager)lsnMgrProc(ptn interface{}, msg *sch.SchMessage) sch.SchErrno {

	//
	// Attention: this task would init the listener while EvSchPoweron event received,
	// and it would be started when EvPeLsnStartReq, which would be sent by peer manager,
	// whose name is sch.PeerMgrName, after EvSchPoweron received. This task and the
	// peer manager task are all static tasks that registered in the table TaskStaticTab,
	// and EvSchPoweron event would be sent by scheduler to all static tasks in order
	// they registered, so we must register this LsnMgrProc task at a position ahead of
	// where peer manager is registered.
	//

	yclog.LogCallerFileLine("LsnMgrProc: " +
		"scheduled, sender: %s, recver: %s, msg: %d",
		sch.SchinfGetMessageSender(msg), sch.SchinfGetMessageRecver(msg), msg.Id)

	var eno sch.SchErrno

	switch msg.Id {

	case sch.EvSchPoweron:
		eno = lsnMgr.lsnMgrPoweron(ptn)

	case sch.EvSchPoweroff:
		eno = lsnMgr.lsnMgrPoweroff()

	case sch.EvPeLsnStartReq:
		eno = lsnMgr.lsnMgrStart()

	case sch.EvPeLsnStopReq:
		eno = lsnMgr.lsnMgrStop()

	default:
		yclog.LogCallerFileLine("LsnMgrProc: invalid message: %d", msg.Id)
		eno = sch.SchEnoParameter
	}

	if eno != sch.SchEnoNone {
		yclog.LogCallerFileLine("LsnMgrProc: errors, eno: %d", eno)
	}

	return eno
}

//
// Poweron event handler
//
func (lsnMgr *ListenerManager)lsnMgrPoweron(ptn interface{}) sch.SchErrno {

	yclog.LogCallerFileLine("lsnMgrPoweron: poweron, carry out task initilization")

	//
	// Keep ourselves task node pointer;
	// Get peer mamager task node pointer;
	//

	var eno sch.SchErrno

	lsnMgr.ptn = ptn
	lsnMgr.sdl = sch.SchinfGetScheduler(ptn)

	if eno, lsnMgr.ptnPeerMgr = lsnMgr.sdl.SchinfGetTaskNodeByName(PeerMgrName); eno != sch.SchEnoNone {
		yclog.LogCallerFileLine("lsnMgrPoweron: invalid peer manager task node pointer")
		return eno
	}

	if lsnMgr.ptnPeerMgr == nil {
		yclog.LogCallerFileLine("lsnMgrPoweron: invalid peer manager task node pointer")
		return sch.SchEnoInternal
	}

	//
	// Get configuration
	//

	lsnMgr.cfg = ycfg.P2pConfig4PeerListener()

	if lsnMgr.cfg == nil {
		yclog.LogCallerFileLine("lsnMgrPoweron: invalid configuration pointer")
		return sch.SchEnoConfig
	}

	return sch.SchEnoNone
}

//
// Setup net lsitener
//
func (lsnMgr *ListenerManager)lsnMgrSetupListener() sch.SchErrno {

	var err error

	lsnAddr := fmt.Sprintf("%s:%d", lsnMgr.cfg.IP.String(), lsnMgr.cfg.Port)

	if lsnMgr.listener, err = net.Listen("tcp", lsnAddr); err != nil {

		yclog.LogCallerFileLine("lsnMgrSetupListener: "+
			"listen failed, addr: %s, err: %s",
			lsnAddr, err.Error())

		return sch.SchEnoOS
	}

	lsnMgr.listenAddr = lsnMgr.listener.Addr().(*net.TCPAddr)

	yclog.LogCallerFileLine("lsnMgrSetupListener: "+
		"task inited ok, listening address: %s",
		lsnMgr.listenAddr.String())

	return sch.SchEnoNone
}

//
// Poweroff event handler
//
func (lsnMgr *ListenerManager)lsnMgrPoweroff() sch.SchErrno {

	yclog.LogCallerFileLine("lsnMgrPoweroff: poweroff, done")

	//
	// kill accepter task if needed
	//

	if _, ptn := lsnMgr.sdl.SchinfGetTaskNodeByName(PeerAccepterName); ptn != nil {
		lsnMgr.lsnMgrStop()
	}

	return lsnMgr.sdl.SchinfTaskDone(lsnMgr.ptn, sch.SchEnoKilled)
}

//
// Startup event handler
//
func (lsnMgr *ListenerManager)lsnMgrStart() sch.SchErrno {

	//
	// When startup signal rceived, we create task which would go into
	// a longlong loop to accept possible inbound connection. Notice that
	// this task would have no chance to receive any messages scheduled
	// to it since it's in a dead loop. To bring this task out, a stop
	// request must be sent to the manager(ourself), which would then try
	// to close the listener, so the task would get out.
	//

	yclog.LogCallerFileLine("lsnMgrStart: try to create accept task ...")

	if eno := lsnMgr.lsnMgrSetupListener(); eno != sch.SchEnoNone {

		yclog.LogCallerFileLine("lsnMgrStart: " +
			"setup listener failed, eno: %d",
			eno)

		return eno
	}

	var accepter = acceptTskCtrlBlock {
		sdl:		lsnMgr.sdl,
		lsnMgr:		lsnMgr,
		event:		sch.SchEnoNone,
	}
	accepter.tep = accepter.peerAcceptProc
	lsnMgr.accepter = &accepter

	var tskDesc = sch.SchTaskDescription{
		Name:		PeerAccepterName,
		MbSize:		0,
		Ep:			&accepter,
		Wd:			&sch.SchWatchDog{HaveDog:false,},
		Flag:		sch.SchCreatedGo,
		DieCb:		nil,
		UserDa:		nil,
	}

	if eno, ptn := lsnMgr.sdl.SchinfCreateTask(&tskDesc); eno != sch.SchEnoNone || ptn == nil {

		yclog.LogCallerFileLine("lsnMgrStart: " +
			"SchinfCreateTask failed, eno: %d, ptn: %X",
			eno, ptn.(*interface{}))

		if ptn == nil {
			return eno
		}

		return sch.SchEnoInternal
	}

	yclog.LogCallerFileLine("lsnMgrStart: accept task created")

	return sch.SchEnoNone
}

//
// Stop event handler
//
func (lsnMgr *ListenerManager)lsnMgrStop() sch.SchErrno {

	yclog.LogCallerFileLine("lsnMgrStop: listner will be closed")

	if lsnMgr.accepter != nil {

		lsnMgr.accepter.lockTcb.Lock()
		defer lsnMgr.accepter.lockTcb.Unlock()

		//
		// Close the listener to force the acceptor task out of the loop,
		// see function acceptProc for details please.
		//

		lsnMgr.accepter.event = sch.SchEnoKilled
		lsnMgr.accepter.listener = nil
	}

	if err := lsnMgr.listener.Close(); err != nil {

		yclog.LogCallerFileLine("lsnMgrStop: try to close listner fialed, err: %s", err.Error())
		return sch.SchEnoOS
	}

	yclog.LogCallerFileLine("lsnMgrStop: listner closed ok")

	lsnMgr.listener = nil

	return sch.SchEnoNone
}

//
// Accept task
//
const PeerAccepterName = sch.PeerAccepterName

type acceptTskCtrlBlock struct {
	sdl			*sch.Scheduler		// pointer to scheduler
	tep			sch.SchUserTaskEp	// entry
	lsnMgr		*ListenerManager	// pointer to listener manager
	ptnPeMgr	interface{}			// pointer to peer manager task node
	ptnLsnMgr	interface{}			// pointer to listener manager task node
	listener	net.Listener		// the listener
	event		sch.SchErrno		// event fired
	curError	error				// current error fired
	lockTcb		sync.Mutex			// lock to protect this control block
	lockAccept	sync.Mutex			// lock to pause/resume acception
}

//
// message for sch.EvPeLsnConnAcceptedInd
//
type msgConnAcceptedInd struct {
	conn		net.Conn
	localAddr	*net.TCPAddr
	remoteAddr	*net.TCPAddr
}

//
// Entry point exported to shceduler
//
func (accepter *acceptTskCtrlBlock)TaskProc4Scheduler(ptn interface{}, msg *sch.SchMessage) sch.SchErrno {
	return accepter.tep(ptn, msg)
}

//
// Accept task entry
//
func (accepter *acceptTskCtrlBlock)peerAcceptProc(ptn interface{}, msg *sch.SchMessage) sch.SchErrno {

	_ = msg

	//
	// Go into a longlong loop to accept peer connections. Please see
	// comments in function lsnMgrStart for more.
	//

	_, accepter.ptnLsnMgr = accepter.sdl.SchinfGetTaskNodeByName(PeerLsnMgrName)

	if accepter.ptnLsnMgr == nil {
		yclog.LogCallerFileLine("PeerAcceptProc: invalid listener manager task pointer")
		accepter.sdl.SchinfTaskDone(ptn, sch.SchEnoInternal)
		return sch.SchEnoInternal
	}

	_, accepter.ptnPeMgr = accepter.sdl.SchinfGetTaskNodeByName(PeerMgrName)

	if accepter.ptnPeMgr == nil {
		yclog.LogCallerFileLine("PeerAcceptProc: invalid peer manager task pointer")
		accepter.sdl.SchinfTaskDone(ptn, sch.SchEnoInternal)
		return sch.SchEnoInternal
	}

	accepter.listener = accepter.lsnMgr.listener

	if accepter.listener == nil {
		yclog.LogCallerFileLine("PeerAcceptProc: invalid listener, done accepter")
		accepter.sdl.SchinfTaskDone(ptn, sch.SchEnoInternal)
		return sch.SchEnoInternal
	}

	accepter.event = sch.EvSchNull
	accepter.curError = nil

	yclog.LogCallerFileLine("PeerAcceptProc: inited ok, tring to accept ...")

acceptLoop:

	for {

		//
		// lock: to know if we are allowed to accept
		//

		accepter.lockAccept.Lock()

		//
		// Check if had been kill by manager: we first obtain the lock then check the listener and
		// event to see if we hav been killed, if not, we backup the listener for later accept operation
		// and free the lock. See function lsnMgrStop for more please.
		//
		// Notice: seems we can apply "chan" to implement the "stop" logic moer better than what we
		// do currently.
		//

		accepter.lockTcb.Lock()

		if accepter.listener == nil || accepter.event != sch.SchEnoNone {
			yclog.LogCallerFileLine("PeerAcceptProc: break the loop, for we might have been killed")
			break acceptLoop
		}

		listener := accepter.listener

		accepter.lockTcb.Unlock()

		//
		// Get lock to accept: unlock it at once since we just want to know if we
		// are allowed to accept, and we might be blocked in calling to Accept().
		//

		accepter.lockAccept.Unlock()

		//
		// Try to accept. Since we had never set deadline for the listener, we
		// would work in a blocked mode; and if here the manager had close the
		// listener, accept would get errors from underlying network.
		//

		yclog.LogCallerFileLine("PeerAcceptProc: try Accept()")

		conn, err := listener.Accept()

		yclog.LogCallerFileLine("PeerAcceptProc: get out from Accept()")

		//
		// Lock the control block since following statements need to access it
		//

		accepter.lockTcb.Lock()

		//
		// Check errors
		//

		if err != nil && !err.(net.Error).Temporary() {

			yclog.LogCallerFileLine("PeerAcceptProc: " +
				"break loop for non-temporary error while accepting, err: %s", err.Error())

			accepter.curError = err
			break acceptLoop
		}

		//
		// Check connection accepted
		//

		if conn == nil {

			yclog.LogCallerFileLine("PeerAcceptProc: " +
				"break loop for null connection accepted without errors")

			accepter.event = sch.EvSchException

			break acceptLoop
		}

		yclog.LogCallerFileLine("PeerAcceptProc: " +
			"accept one: %s",
			conn.RemoteAddr().String())

		//
		// Connection got, hand it up to peer manager task, notice that we will continue the loop
		// event when we get errors to make and send the message to peer manager, see bellow.
		//

		var msg = sch.SchMessage{}
		var msgBody = msgConnAcceptedInd {
			conn: 		conn,
			localAddr:	conn.LocalAddr().(*net.TCPAddr),
			remoteAddr:	conn.RemoteAddr().(*net.TCPAddr),
		}

		eno := accepter.sdl.SchinfMakeMessage(&msg, ptn, accepter.ptnPeMgr, sch.EvPeLsnConnAcceptedInd, &msgBody)
		if eno != sch.SchEnoNone {

			yclog.LogCallerFileLine("PeerAcceptProc: " +
				"SchinfMakeMessage for EvPeLsnConnAcceptedInd failed, eno: %d",
				eno)

			accepter.lockTcb.Unlock()

			continue
		}

		eno = accepter.sdl.SchinfSendMessage(&msg)
		if eno != sch.SchEnoNone {

			yclog.LogCallerFileLine("PeerAcceptProc: " +
				"SchinfSendMessage for EvPeLsnConnAcceptedInd failed, target: %s",
				accepter.sdl.SchinfGetTaskName(accepter.ptnPeMgr))

			accepter.lockTcb.Unlock()

			continue
		}

		yclog.LogCallerFileLine("PeerAcceptProc: " +
			"send EvPeLsnConnAcceptedInd ok, target: %s",
			accepter.sdl.SchinfGetTaskName(accepter.ptnPeMgr))

		accepter.lockTcb.Unlock()
	}

	//
	// Notice: when loop is broken to here, the Lock is still obtained by us,
	// see above pls, do not lock again.
	//
	// Here we get out! We should check what had happened to break the loop
	// for accepting above.
	//

	if accepter.curError != nil && accepter.event != sch.SchEnoNone {

		//
		// This is the normal case: the loop is broken by manager task, or
		// errors fired from underlying network.
		//

		yclog.LogCallerFileLine("PeerAcceptProc: broken for event: %d", accepter.event)

		accepter.sdl.SchinfTaskDone(ptn, accepter.event)
		accepter.lockTcb.Unlock()

		return accepter.event
	}

	//
	// Abnormal case, we should never come here, debug out and then done the
	// accepter task.
	//

	if accepter.curError != nil {

		yclog.LogCallerFileLine("PeerAcceptProc: abnormal exit, event: %d, err: %s",
			accepter.event, accepter.curError.Error())

	} else {

		yclog.LogCallerFileLine("PeerAcceptProc: abnormal exit, event: %d, err: nil",
			accepter.event)
	}

	accepter.lockTcb.Unlock()
	accepter.sdl.SchinfTaskDone(ptn, sch.SchEnoUnknown)

	return sch.SchEnoUnknown
}

//
// Pause accept
//
func (accepter *acceptTskCtrlBlock)PauseAccept() bool {
	yclog.LogCallerFileLine("PauseAccept: try to pause accepting inbound")
	accepter.lockAccept.Lock()
	return true
}

//
// Resume accept
//
func (accepter *acceptTskCtrlBlock)ResumeAccept() bool {
	yclog.LogCallerFileLine("PauseAccept: try to resume accepting inbound")
	accepter.lockAccept.Unlock()
	return true
}
