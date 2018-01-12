## tpV2Proto

 Compatible teleport v2 protocol

```
HeaderLength | HeaderCodecId | Header | BodyLength | BodyCodecId | Body
```

**Notes:**
- `HeaderLength`: uint32, 4 bytes, big endian
- `HeaderCodecId`: uint8, 1 byte
- `Header`: header bytes(`use protobuf`)
- `BodyLength`: uint32, 4 bytes, big endian
	* may be 0, meaning that the `Body` is empty and does not indicate the `BodyCodecId`
	* may be 1, meaning that the `Body` is empty but indicates the `BodyCodecId`
- `BodyCodecId`: uint8, 1 byte
- `Body`: body bytes

### Usage

`import tpV2Proto "github.com/henrylee2cn/tp-ext/proto-tpV2Proto"`

#### Test

```go
package tpV2Proto_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	tpV2Proto "github.com/henrylee2cn/tp-ext/proto-tpV2Proto"
)

type Home struct {
	tp.PullCtx
}

func (h *Home) Test(args *map[string]interface{}) (map[string]interface{}, *tp.Rerror) {
	return map[string]interface{}{
		"your_id": h.Query().Get("peer_id"),
	}, nil
}

func TestTpV2Proto(t *testing.T) {
	// Server
	svr := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"})
	svr.PullRouter.Reg(new(Home))
	go svr.Listen(tpV2Proto.DefaultProtoFunc())
	time.Sleep(1e9)

	// Client
	cli := tp.NewPeer(tp.PeerConfig{})
	sess, err := cli.Dial(":9090", tpV2Proto.DefaultProtoFunc())
	if err != nil {
		if err != nil {
			t.Error(err)
		}
	}
	var reply interface{}
	rerr := sess.Pull("/home/test?peer_id=110",
		// map[string]interface{}{
		// 	"bytes": []byte("test bytes"),
		// },
		nil,
		&reply,
		socket.WithAddMeta("add", "1"),
		socket.WithSetMeta("set", "2"),
	).Rerror()
	if rerr != nil {
		t.Error(rerr)
	}
	t.Logf("=========reply:%v", reply)
}
```

test command:

```sh
go test -v -run=TestTpV2Proto
```
