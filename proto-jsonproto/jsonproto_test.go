package jsonproto_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	jsonproto "github.com/henrylee2cn/tp-ext/proto-jsonproto"
)

type Home struct {
	tp.PullCtx
}

func (h *Home) Test(args *map[string]interface{}) (map[string]interface{}, *tp.Rerror) {
	h.Session().Push("/push/test", map[string]interface{}{
		"your_id": h.Query().Get("peer_id"),
	})
	meta := h.CopyMeta()
	return map[string]interface{}{
		"args": *args,
		"meta": meta.String(),
	}, nil
}

func TestJsonProto(t *testing.T) {
	// Server
	srv := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"})
	srv.RoutePull(new(Home))
	go srv.ListenAndServe(jsonproto.NewJsonProtoFunc)
	time.Sleep(1e9)

	// Client
	cli := tp.NewPeer(tp.PeerConfig{})
	cli.RoutePush(new(Push))
	sess, err := cli.Dial(":9090", jsonproto.NewJsonProtoFunc)
	if err != nil {
		t.Error(err)
	}
	var reply interface{}
	rerr := sess.Pull("/home/test?peer_id=110",
		map[string]interface{}{
			"bytes": []byte("test bytes"),
		},
		&reply,
		tp.WithAddMeta("add", "1"),
		tp.WithXferPipe('g'),
	).Rerror()
	if rerr != nil {
		t.Error(rerr)
	}
	t.Logf("reply:%v", reply)
	time.Sleep(3e9)
}

type Push struct {
	tp.PushCtx
}

func (p *Push) Test(args *map[string]interface{}) *tp.Rerror {
	tp.Infof("receive push(%s):\nargs: %#v\n", p.Ip(), args)
	return nil
}
