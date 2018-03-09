## websocket

Websocket is an extension package that makes the Teleport framework compatible with websocket protocol as specified in RFC 6455.

### Usage

`import websocket "github.com/henrylee2cn/tp-ext/mod-websocket"`

#### Test

```go
package websocket_test

import (
	"net/http"
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	jsonproto "github.com/henrylee2cn/tp-ext/proto-jsonproto"
	websocket "github.com/henrylee2cn/tp-ext/mod-websocket"
)

type Args struct {
	A int
	B int `param:"<range:1:>"`
}

type P struct{ tp.PullCtx }

func (p *P) Divide(args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

func TestWebsocket(t *testing.T) {
	srv := tp.NewPeer(tp.PeerConfig{})
	http.Handle("/ws", websocket.NewServeHandler(srv, nil, jsonproto.NewProtoFunc))
	go http.ListenAndServe("0.0.0.0:9090", nil)
	srv.RoutePull(new(P))
	time.Sleep(time.Second * 1)

	cli := tp.NewPeer(tp.PeerConfig{}, websocket.NewDialPlugin("/ws"))
	sess, err := cli.Dial("127.0.0.1:9090", jsonproto.NewProtoFunc)
	if err != nil {
		t.Fatal(err)
	}
	var reply int
	rerr := sess.Pull("/p/divide", &Args{
		A: 10,
		B: 2,
	}, &reply).Rerror()
	if rerr != nil {
		t.Fatal(rerr)
	}
	t.Logf("10/2=%d", reply)
	time.Sleep(time.Second)
}
```

test command:

```sh
go test -v -run=TestWebsocket
```