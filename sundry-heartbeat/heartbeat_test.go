package heartbeat_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	heartbeat "github.com/henrylee2cn/tp-ext/sundry-heartbeat"
)

func TestHeartbeat(t *testing.T) {
	srv := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"})
	heartbeat.WithPong(srv)
	go srv.Listen()

	cli := tp.NewPeer(tp.PeerConfig{})
	heartbeat.WithPing(cli, time.Second)
	cli.Dial(":9090")
	time.Sleep(time.Second * 10)
}
