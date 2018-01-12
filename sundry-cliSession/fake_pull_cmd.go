package cliSession

import (
	"time"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
)

// NewFakePullCmd creates a fake tp.PullCmd
func NewFakePullCmd(peer *tp.Peer, uri string, args, reply interface{}, rerr *tp.Rerror, setting ...socket.PacketSetting) tp.PullCmd {
	output := socket.NewPacket(
		socket.WithPtype(tp.TypePull),
		socket.WithUri(uri),
		socket.WithBody(args),
	)
	for _, fn := range setting {
		fn(output)
	}
	return &fakePullCmd{
		peer:   peer,
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
