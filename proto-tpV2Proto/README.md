## tpV2Proto

 Compatible teleport v2 protocol

```
HeaderLength | HeaderCodecId | Header | BodyLength | BodyCodecId | Body
```


**Notes:**
- `HeaderLength`: uint32, 4 bytes, big endian
- `HeaderCodecId`: uint8, 1 byte, header use protobuf(constant: `p`)
- `Header`: header bytes
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
	srv := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"})
	srv.RoutePull(new(Home))
	go srv.ListenAndServe(tpV2Proto.NewProtoFunc)
	time.Sleep(1e9)

	// Client
	cli := tp.NewPeer(tp.PeerConfig{})
	sess, err := cli.Dial(":9090", tpV2Proto.NewProtoFunc)
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
		tp.WithAddMeta("add", "1"),
		tp.WithSetMeta("set", "2"),
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
