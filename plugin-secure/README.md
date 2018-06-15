## secure

Package secure encrypting/decrypting the packet body.

### Usage

`import secure "github.com/henrylee2cn/tp-ext/plugin-secure"`

Ciphertext struct:

```go
package secure_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	secure "github.com/henrylee2cn/tp-ext/plugin-secure"
)

type Arg struct {
	A int
	B int
}

type Result struct {
	C int
}

type math struct{ tp.PullCtx }

func (m *math) Add(arg *Arg) (*Result, *tp.Rerror) {
	// enforces the body of the encrypted reply packet.
	// secure.EnforceSecure(m.Output())

	return &Result{C: arg.A + arg.B}, nil
}

func newSession(t *testing.T) tp.Session {
	p := secure.NewSecurePlugin(100001, "cipherkey1234567")
	srv := tp.NewPeer(tp.PeerConfig{
		ListenAddress: ":9090",
		PrintDetail:   true,
	})
	srv.RoutePull(new(math), p)
	go srv.ListenAndServe()
	time.Sleep(time.Second)

	cli := tp.NewPeer(tp.PeerConfig{
		PrintDetail: true,
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
	var result Result
	rerr := sess.Pull(
		"/math/add",
		&Arg{A: 10, B: 2},
		&result,
		secure.WithSecureMeta(),
		// secure.WithAcceptSecureMeta(false),
	).Rerror()
	if rerr != nil {
		t.Fatal(rerr)
	}
	if result.C != 12 {
		t.Fatalf("expect 12, but get %d", result.C)
	}
	t.Logf("test secure10+2=%d", result.C)
}

func TestAcceptSecurePlugin(t *testing.T) {
	sess := newSession(t)
	// test accept secure
	var result Result
	rerr := sess.Pull(
		"/math/add",
		&Arg{A: 20, B: 4},
		&result,
		secure.WithAcceptSecureMeta(true),
	).Rerror()
	if rerr != nil {
		t.Fatal(rerr)
	}
	if result.C != 24 {
		t.Fatalf("expect 24, but get %d", result.C)
	}
	t.Logf("test accept secure: 20+4=%d", result.C)
}
```

test command:

```sh
go test -v -run=TestSecurePlugin
go test -v -run=TestAcceptSecurePlugin
```