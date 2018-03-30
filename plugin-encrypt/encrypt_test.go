package encrypt_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	encrypt "github.com/henrylee2cn/tp-ext/plugin-encrypt"
)

type Args struct {
	A int
	B int
}

type Result struct {
	C int
}

type math struct{ tp.PullCtx }

func (*math) Add(args *Args) (*Result, *tp.Rerror) {
	return &Result{C: args.A + args.B}, nil
}

func TestEncryptPlugin(t *testing.T) {
	p := encrypt.NewEncryptPlugin(100001, "cipherkey1234567")

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
	var reply Result
	rerr := sess.Pull("/math/add", &Args{
		A: 10,
		B: 2,
	}, &reply).Rerror()
	if rerr != nil {
		t.Fatal(rerr)
	}
	if reply.C != 12 {
		t.Fatalf("expect 12, but get %d", reply.C)
	}
	t.Logf("10+2=%d", reply.C)
}
