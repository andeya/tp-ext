## md5Hash

Provides a integrity check transfer filter

### Usage

`import md5Hash "github.com/henrylee2cn/tp-ext/xfer-md5Hash"`

#### Test

```go
package md5Hash_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/xfer"
	md5Hash "github.com/henrylee2cn/tp-ext/xfer-md5Hash"
)

func TestSeparate(t *testing.T) {
	// Register filter(custom)
	md5Hash.Reg('m', "md5")
	md5Check, _ := xfer.Get('m')
	input := []byte("md5")
	b, err := md5Check.OnPack(input)
	if err != nil {
		t.Fatalf("Onpack: %v", err)
	}

	_, err = md5Check.OnUnpack(b)
	if err != nil {
		t.Fatalf("Md5 check failed: %v", err)
	}

	// Tamper with data
	b = append(b, "viruses"...)
	_, err = md5Check.OnUnpack(b)
	if err == nil {
		t.Fatal("Md5 check failed:")
	}
}

func TestCombined(t *testing.T) {
	// Register filter(custom)
	md5Hash.Reg('m', "md5")
	// Server
	srv := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"})
	srv.RoutePull(new(Home))
	go srv.ListenAndServe()
	time.Sleep(1e9)

	// Client
	cli := tp.NewPeer(tp.PeerConfig{})
	sess, err := cli.Dial(":9090")
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
		// Use custom filter
		tp.WithXferPipe('m'),
	).Rerror()
	if rerr != nil {
		t.Error(rerr)
	}
	t.Logf("=========reply:%v", reply)
}

type Home struct {
	tp.PullCtx
}

func (h *Home) Test(args *map[string]interface{}) (map[string]interface{}, *tp.Rerror) {
	return map[string]interface{}{
		"your_id": h.Query().Get("peer_id"),
	}, nil
}
```

test command:
```sh
# Separate test
go test -v -run=TestSeparate
# Combined with teleport test
go test -v -run=TestCombined
```
