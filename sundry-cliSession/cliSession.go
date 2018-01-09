// CliSession client session which has connection pool.
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

	"github.com/henrylee2cn/goutil"
	"github.com/henrylee2cn/goutil/pool"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
)

// CliSession client session which is has connection pool
type CliSession struct {
	peer *tp.Peer
	pool *pool.Workshop
}

// New creates a client session which is has connection pool.
func New(peer *tp.Peer, addr string, sessMaxQuota int, sessMaxIdleDuration time.Duration, protoFunc ...socket.ProtoFunc) *CliSession {
	newWorkerFunc := func() (pool.Worker, error) {
		sess, rerr := peer.Dial(addr, protoFunc...)
		return sess, rerr.ToError()
	}
	return &CliSession{
		peer: peer,
		pool: pool.NewWorkshop(sessMaxQuota, sessMaxIdleDuration, newWorkerFunc),
	}
}

// AsyncPull sends a packet and receives reply asynchronously.
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name.
func (c *CliSession) AsyncPull(uri string, args interface{}, reply interface{}, done chan tp.PullCmd, setting ...socket.PacketSetting) {
	_sess, err := c.pool.Hire()
	if err != nil {
		done <- c.fakePullCmd(uri, args, reply, tp.ToRerror(err), setting...)
		return
	}
	sess := _sess.(tp.Session)
	defer c.pool.Fire(sess)
	sess.AsyncPull(uri, args, reply, done, setting...)
}

// Pull sends a packet and receives reply.
// Note:
// If the args is []byte or *[]byte type, it can automatically fill in the body codec name;
// If the session is a client role and PeerConfig.RedialTimes>0, it is automatically re-called once after a failure.
func (c *CliSession) Pull(uri string, args interface{}, reply interface{}, setting ...socket.PacketSetting) tp.PullCmd {
	doneChan := make(chan tp.PullCmd, 1)
	c.AsyncPull(uri, args, reply, doneChan, setting...)
	pullCmd := <-doneChan
	close(doneChan)
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

// Close closes the session.
func (c *CliSession) Close() {
	c.pool.Close()
}

// Stats returns the current session pool stats.
func (c *CliSession) Stats() pool.WorkshopStats {
	return c.pool.Stats()
}

func (c *CliSession) fakePullCmd(uri string, args, reply interface{}, rerr *tp.Rerror, setting ...socket.PacketSetting) tp.PullCmd {
	output := socket.NewPacket(
		socket.WithPtype(tp.TypePull),
		socket.WithUri(uri),
		socket.WithBody(args),
	)
	for _, fn := range setting {
		fn(output)
	}
	return &fakePullCmd{
		peer:   c.peer,
		reply:  reply,
		rerr:   rerr,
		output: output,
	}
}

type fakePullCmd struct {
	peer   *tp.Peer
	reply  interface{}
	rerr   *tp.Rerror
	output *socket.Packet
}

// Peer returns the peer.
func (c *fakePullCmd) Peer() *tp.Peer {
	return c.peer
}

// Session returns the session.
func (c *fakePullCmd) Session() tp.Session {
	return nil
}

// Ip returns the remote addr.
func (c *fakePullCmd) Ip() string {
	return ""
}

// Public returns temporary public data of context.
func (c *fakePullCmd) Public() goutil.Map {
	return nil
}

// PublicLen returns the length of public data of context.
func (c *fakePullCmd) PublicLen() int {
	return 0
}

// Output returns writed packet.
func (c *fakePullCmd) Output() *socket.Packet {
	return c.output
}

// Result returns the pull result.
func (c *fakePullCmd) Result() (interface{}, *tp.Rerror) {
	return c.reply, c.rerr
}

// *Rerror returns the pull error.
func (c *fakePullCmd) Rerror() *tp.Rerror {
	return c.rerr
}

// CostTime returns the pulled cost time.
// If PeerConfig.CountTime=false, always returns 0.
func (c *fakePullCmd) CostTime() time.Duration {
	return 0
}
