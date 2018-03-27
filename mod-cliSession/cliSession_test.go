package cliSession_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	cliSession "github.com/henrylee2cn/tp-ext/mod-cliSession"
)

type Args struct {
	A int
	B int `param:"<range:1:>"`
}

type P struct{ tp.PullCtx }

func (p *P) Divide(args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

func TestCliSession(t *testing.T) {
	srv := tp.NewPeer(tp.PeerConfig{
		ListenAddress: ":9090",
	})
	srv.RoutePull(new(P))
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := cliSession.New(
		tp.NewPeer(tp.PeerConfig{}),
		":9090",
		100,
		time.Second*5,
	)
	go func() {
		for {
			t.Logf("%+v", cli.Stats())
			time.Sleep(time.Millisecond * 500)
		}
	}()
	go func() {
		var reply int
		for i := 0; ; i++ {
			rerr := cli.Pull("/p/divide", &Args{
				A: i,
				B: 2,
			}, &reply).Rerror()
			if rerr != nil {
				t.Log(rerr)
			} else {
				t.Logf("%d/2=%v", i, reply)
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
	time.Sleep(time.Second * 6)
	cli.Close()
	time.Sleep(time.Second * 3)
}
