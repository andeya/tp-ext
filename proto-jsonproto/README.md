## jsonproto

jsonproto is implemented JSON protocol.


### Data Packet 

`Length``JSON`

- `Length`: uint32, 4 bytes, big endian
- `JSON`: {"seq":%d,"ptype":%d,"uri":%q,"meta":%q,"body_codec":%d,"body":"%s","xfer_pipe":%s}

Demo:

```
83{"seq":%d,"ptype":%d,"uri":%q,"meta":%q,"body_codec":%d,"body":"%s","xfer_pipe":%s}
```

### Usage

`import jsonproto "github.com/henrylee2cn/tp-ext/proto-jsonproto"`

#### Test

```go
package jsonproto_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
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

func TestJsonproto(t *testing.T) {
	// Server
	svr := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"})
	svr.PullRouter.Reg(new(Home))
	go svr.Listen(jsonproto.NewJsonproto)
	time.Sleep(1e9)

	// Client
	cli := tp.NewPeer(tp.PeerConfig{})
	cli.PushRouter.Reg(new(Push))
	sess, err := cli.Dial(":9090", jsonproto.NewJsonproto)
	if err != nil {
		if err != nil {
			t.Error(err)
		}
	}
	var reply interface{}
	rerr := sess.Pull("/home/test?peer_id=110",
		map[string]interface{}{
			"bytes": []byte("test bytes"),
		},
		&reply,
		socket.WithAddMeta("add", "1"),
		socket.WithXferPipe('g'),
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
```

test command:

```sh
go test -v -run=TestJsonproto
```