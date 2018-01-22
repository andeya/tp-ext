package heartbeat_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	heartbeat "github.com/henrylee2cn/tp-ext/plugin-heartbeat"
)

func TestHeartbeat1(t *testing.T) {
	srv := tp.NewPeer(
		tp.PeerConfig{ListenAddress: ":9090"},
		heartbeat.NewPong(),
	)
	go srv.Listen()
	time.Sleep(time.Second)

	cli := tp.NewPeer(
		tp.PeerConfig{},
		heartbeat.NewPing(time.Second),
	)
	cli.Dial(":9090")
	time.Sleep(time.Second * 10)
}

func TestHeartbeat2(t *testing.T) {
	srv := tp.NewPeer(
		tp.PeerConfig{ListenAddress: ":9090"},
		heartbeat.NewPong(),
	)
	go srv.Listen()
	time.Sleep(time.Second)

	cli := tp.NewPeer(
		tp.PeerConfig{},
		heartbeat.NewPing(time.Second*3),
	)
	sess, _ := cli.Dial(":9090")
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * 2)
		sess.Pull("uri", nil, nil)
	}
}
