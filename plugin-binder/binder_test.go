package binder_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	binder "github.com/henrylee2cn/tp-ext/plugin-binder"
)

type (
	Args struct {
		A int
		B int `param:"<range:1:100>"`
		Query
		XyZ string `param:"<query><nonzero><rerr: 100002: Parameter cannot be empty>"`
	}
	Query struct {
		X string `param:"<query>"`
	}
)

type P struct{ tp.PullCtx }

func (p *P) Divide(args *Args) (int, *tp.Rerror) {
	tp.Infof("query args x: %s, xy_z: %s", args.Query.X, args.XyZ)
	return args.A / args.B, nil
}

func TestBinder(t *testing.T) {
	bplugin := binder.NewStructArgsBinder(nil)
	srv := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"})
	srv.RoutePull(new(P), bplugin)
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := tp.NewPeer(tp.PeerConfig{})
	sess, err := cli.Dial(":9090")
	if err != nil {
		t.Fatal(err)
	}
	var reply int
	rerr := sess.Pull("/p/divide?x=testquery_x&xy_z=testquery_xy_z", &Args{
		A: 10,
		B: 2,
	}, &reply).Rerror()
	if rerr != nil {
		t.Fatal(rerr)
	}
	t.Logf("10/2=%d", reply)
	rerr = sess.Pull("/p/divide", &Args{
		A: 10,
		B: 5,
	}, &reply).Rerror()
	if rerr == nil {
		t.Fatal(rerr)
	}
	t.Logf("10/5 error:%v", rerr)
	rerr = sess.Pull("/p/divide", &Args{
		A: 10,
		B: 0,
	}, &reply).Rerror()
	if rerr == nil {
		t.Fatal(rerr)
	}
	t.Logf("10/0 error:%v", rerr)
}
