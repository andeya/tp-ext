package websocket_test

import (
	"net/http"
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	websocket "github.com/henrylee2cn/tp-ext/mod-websocket"
	"github.com/henrylee2cn/tp-ext/mod-websocket/jsonSubProto"
	"github.com/henrylee2cn/tp-ext/mod-websocket/pbSubProto"
)

type Arg struct {
	A int
	B int `param:"<range:1:>"`
}

type P struct{ tp.PullCtx }

func (p *P) Divide(arg *Arg) (int, *tp.Rerror) {
	return arg.A / arg.B, nil
}

func TestJsonSubWebsocket(t *testing.T) {
	srv := tp.NewPeer(tp.PeerConfig{})
	http.Handle("/ws", websocket.NewJsonServeHandler(srv, nil))
	go http.ListenAndServe("0.0.0.0:9090", nil)
	srv.RoutePull(new(P))
	time.Sleep(time.Second * 1)

	cli := tp.NewPeer(tp.PeerConfig{}, websocket.NewDialPlugin("/ws"))
	sess, err := cli.Dial("127.0.0.1:9090", jsonSubProto.NewJsonSubProtoFunc)
	if err != nil {
		t.Fatal(err)
	}
	var result int
	rerr := sess.Pull("/p/divide", &Arg{
		A: 10,
		B: 2,
	}, &result,
	// tp.WithXferPipe('g'),
	).Rerror()
	if rerr != nil {
		t.Fatal(rerr)
	}
	t.Logf("10/2=%d", result)
	time.Sleep(time.Second)
}

func TestPbSubWebsocket(t *testing.T) {
	srv := tp.NewPeer(tp.PeerConfig{})
	http.Handle("/ws", websocket.NewPbServeHandler(srv, nil))
	go http.ListenAndServe("0.0.0.0:9090", nil)
	srv.RoutePull(new(P))
	time.Sleep(time.Second * 1)

	cli := tp.NewPeer(tp.PeerConfig{}, websocket.NewDialPlugin("/ws"))
	sess, err := cli.Dial("127.0.0.1:9090", pbSubProto.NewPbSubProtoFunc)
	if err != nil {
		t.Fatal(err)
	}
	var result int
	rerr := sess.Pull("/p/divide", &Arg{
		A: 10,
		B: 2,
	}, &result,
	// tp.WithXferPipe('g'),
	).Rerror()
	if rerr != nil {
		t.Fatal(rerr)
	}
	t.Logf("10/2=%d", result)
	time.Sleep(time.Second)
}
