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
		XyZ       string  `param:"<query><nonzero><rerr: 100002: Parameter cannot be empty>"`
		SwapValue float32 `param:"<swap><nonzero>"`
	}
	Query struct {
		X string `param:"<query:_x>"`
	}
)

type P struct{ tp.PullCtx }

func (p *P) Divide(args *Args) (int, *tp.Rerror) {
	tp.Infof("query args _x: %s, xy_z: %s, swap_value: %v", args.Query.X, args.XyZ, args.SwapValue)
	return args.A / args.B, nil
}

type SwapPlugin struct{}

func (s *SwapPlugin) Name() string {
	return "swap_plugin"
}
func (s *SwapPlugin) PostReadPullBody(ctx tp.ReadCtx) *tp.Rerror {
	ctx.Swap().Store("swap_value", 123)
	return nil
}

func TestBinder(t *testing.T) {
	bplugin := binder.NewStructArgsBinder(nil)
	srv := tp.NewPeer(
		tp.PeerConfig{ListenAddress: ":9090"},
	)
	srv.PluginContainer().AppendRight(bplugin)
	srv.RoutePull(new(P), new(SwapPlugin))
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := tp.NewPeer(tp.PeerConfig{})
	sess, err := cli.Dial(":9090")
	if err != nil {
		t.Fatal(err)
	}
	var reply int
	rerr := sess.Pull("/p/divide?_x=testquery_x&xy_z=testquery_xy_z", &Args{
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
