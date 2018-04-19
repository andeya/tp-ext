package secure_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	secure "github.com/henrylee2cn/tp-ext/plugin-secure"
)

type Args struct {
	A int
	B int
}

type Result struct {
	C int
}

type math struct{ tp.PullCtx }

func (m *math) Add(args *Args) (*Result, *tp.Rerror) {
	// all encrypted
	// secure.UseSecure(m.Output())

	return &Result{C: args.A + args.B}, nil
}

func newSession(t *testing.T) tp.Session {
	p := secure.NewSecurePlugin(100001, "cipherkey1234567")
	srv := tp.NewPeer(tp.PeerConfig{
		ListenAddress: ":9090",
		PrintBody:     true,
	})
	srv.RoutePull(new(math), p)
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := tp.NewPeer(tp.PeerConfig{
		PrintBody: true,
	}, p)
	sess, err := cli.Dial(":9090")
	if err != nil {
		t.Fatal(err)
	}
	return sess
}

func TestSecurePlugin(t *testing.T) {
	sess := newSession(t)
	// test secure
	var reply Result
	rerr := sess.Pull(
		"/math/add",
		&Args{A: 10, B: 2},
		&reply,
		secure.WithSecureMeta(),
	).Rerror()
	if rerr != nil {
		t.Fatal(rerr)
	}
	if reply.C != 12 {
		t.Fatalf("expect 12, but get %d", reply.C)
	}
	t.Logf("test secure10+2=%d", reply.C)
}

func TestAcceptSecurePlugin(t *testing.T) {
	sess := newSession(t)
	// test accept secure
	var reply Result
	rerr := sess.Pull(
		"/math/add",
		&Args{A: 20, B: 4},
		&reply,
		secure.WithAcceptSecureMeta(),
	).Rerror()
	if rerr != nil {
		t.Fatal(rerr)
	}
	if reply.C != 24 {
		t.Fatalf("expect 24, but get %d", reply.C)
	}
	t.Logf("test accept secure: 20+4=%d", reply.C)
}
