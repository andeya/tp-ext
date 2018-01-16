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
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/henrylee2cn/teleport/utils"
)

// NewFakePullCmd creates a fake tp.PullCmd
func NewFakePullCmd(peer tp.Peer, uri string, args, reply interface{}, rerr *tp.Rerror, setting ...socket.PacketSetting) tp.PullCmd {
	output := socket.NewPacket(
		socket.WithPtype(tp.TypePull),
		socket.WithUri(uri),
		socket.WithBody(args),
	)
	for _, fn := range setting {
		fn(output)
	}
	return &fakePullCmd{
		peer:      peer,
		reply:     reply,
		rerr:      rerr,
		output:    output,
		inputMeta: utils.AcquireArgs(),
	}
}

type fakePullCmd struct {
	peer      tp.Peer
	reply     interface{}
	rerr      *tp.Rerror
	output    *socket.Packet
	inputMeta *utils.Args
}

// Peer returns the peer.
func (c *fakePullCmd) Peer() tp.Peer {
	return c.peer
}

// Session returns the session.
func (c *fakePullCmd) Session() tp.Session {
	return nil
}

// Id returns the session id.
func (c *fakePullCmd) Id() string {
	return ""
}

// RealId returns the current real remote id.
func (c *fakePullCmd) RealId() string {
	return ""
}

// Ip returns the remote addr.
func (c *fakePullCmd) Ip() string {
	return ""
}

// RealIp returns the the current real remote addr.
func (c *fakePullCmd) RealIp() string {
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

// InputMeta returns the header metadata of input packet.
func (c *fakePullCmd) InputMeta() *utils.Args {
	return c.inputMeta
}

// CostTime returns the pulled cost time.
// If PeerConfig.CountTime=false, always returns 0.
func (c *fakePullCmd) CostTime() time.Duration {
	return 0
}
