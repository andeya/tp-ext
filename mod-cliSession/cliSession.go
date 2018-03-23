// Client session with a high efficient and load balanced connection pool.
//
// Copyright 2018 HenryLee. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cliSession

import (
	"time"

	"github.com/henrylee2cn/goutil/pool"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
)

// CliSession client session which is has connection pool
type CliSession struct {
	addr string
	peer tp.Peer
	pool *pool.Workshop
}

// New creates a client session which is has connection pool.
func New(peer tp.Peer, addr string, sessMaxQuota int, sessMaxIdleDuration time.Duration, protoFunc ...socket.ProtoFunc) *CliSession {
	newWorkerFunc := func() (pool.Worker, error) {
		sess, rerr := peer.Dial(addr, protoFunc...)
		return sess, rerr.ToError()
	}
	return &CliSession{
		addr: addr,
		peer: peer,
		pool: pool.NewWorkshop(sessMaxQuota, sessMaxIdleDuration, newWorkerFunc),
	}
}

// Addr returns the address.
func (c *CliSession) Addr() string {
	return c.addr
}

// Peer returns the peer.
func (c *CliSession) Peer() tp.Peer {
	return c.peer
}

// Close closes the session.
func (c *CliSession) Close() {
	c.pool.Close()
}

// Stats returns the current session pool stats.
func (c *CliSession) Stats() pool.WorkshopStats {
	return c.pool.Stats()
}

// AsyncPull sends a packet and receives reply asynchronously.
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name.
func (c *CliSession) AsyncPull(
	uri string,
	args interface{},
	reply interface{},
	pullCmdChan chan<- tp.PullCmd,
	setting ...socket.PacketSetting,
) tp.PullCmd {
	_sess, err := c.pool.Hire()
	if err != nil {
		pullCmd := tp.NewFakePullCmd(uri, args, reply, tp.ToRerror(err))
		if pullCmdChan != nil && cap(pullCmdChan) == 0 {
			tp.Panicf("*CliSession.AsyncPull(): pullCmdChan channel is unbuffered")
		}
		pullCmdChan <- pullCmd
		return pullCmd
	}
	sess := _sess.(tp.Session)
	defer c.pool.Fire(sess)
	return sess.AsyncPull(uri, args, reply, pullCmdChan, setting...)
}

// Pull sends a packet and receives reply.
// Note:
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *CliSession) Pull(uri string, args interface{}, reply interface{}, setting ...socket.PacketSetting) tp.PullCmd {
	pullCmd := c.AsyncPull(uri, args, reply, make(chan tp.PullCmd, 1), setting...)
	<-pullCmd.Done()
	return pullCmd
}

// Push sends a packet, but do not receives reply.
// Note:
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *CliSession) Push(uri string, args interface{}, setting ...socket.PacketSetting) *tp.Rerror {
	_sess, err := c.pool.Hire()
	if err != nil {
		return tp.ToRerror(err)
	}
	sess := _sess.(tp.Session)
	defer c.pool.Fire(sess)
	return sess.Push(uri, args, setting...)
}
