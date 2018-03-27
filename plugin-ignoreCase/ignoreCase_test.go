package ignoreCase_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	ignoreCase "github.com/henrylee2cn/tp-ext/plugin-ignoreCase"
)

type Home struct {
	tp.PullCtx
}

func (h *Home) Test(args *map[string]interface{}) (map[string]interface{}, *tp.Rerror) {
	h.Session().Push("/push/tesT", map[string]interface{}{
		"your_id": h.Query().Get("peer_id"),
	})
	meta := h.CopyMeta()
	time.Sleep(5e9)

	return map[string]interface{}{
		"args": *args,
		"meta": meta.String(),
	}, nil
}

func TestIngoreCase(t *testing.T) {
	// Server
	srv := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"}, ignoreCase.NewIgnoreCase())
	srv.RoutePull(new(Home))
	go srv.ListenAndServe()
	time.Sleep(1e9)

	// Client
	cli := tp.NewPeer(tp.PeerConfig{}, ignoreCase.NewIgnoreCase())
	cli.RoutePush(new(Push))
	sess, err := cli.Dial(":9090")
	if err != nil {
		if err != nil {
			t.Error(err)
		}
	}
	var reply interface{}
	rerr := sess.Pull("/home/tesT?peer_id=110",
		map[string]interface{}{
			"bytes": []byte("test bytes"),
		},
		&reply,
		tp.WithAddMeta("add", "1"),
	).Rerror()
	if rerr != nil {
		t.Error(rerr)
	}
	t.Logf("reply:%v", reply)
}

type Push struct {
	tp.PushCtx
}

func (p *Push) Test(args *map[string]interface{}) *tp.Rerror {
	tp.Infof("receive push(%s):\nargs: %#v\n", p.Ip(), args)
	return nil
}
