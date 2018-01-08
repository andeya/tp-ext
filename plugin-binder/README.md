## binder

Parameter Binding Verification Plugin for Struct Handler.

### Usage

`import binder "github.com/henrylee2cn/tp-ext/plugin-binder"`

#### Param-Tags

tag   |   key    | required |     value     |   desc
------|----------|----------|---------------|----------------------------------
param |    query    | no |     -      | It indicates that the parameter is from the URI query part, else the parameter is from body. e.g. `/a/b?x={query}`
param |   desc   |      no      |     (e.g.`id`)   | Parameter Description
param |   len    |      no      |   (e.g.`3:6``3`)  | The length of the string type parameter
param |   range  |      no      |   (e.g.`0:10`)   | The range of parameters for the numeric type
param |  nonzero |      no      |    -    | Not allowed to zero
param |  regexp  |      no      |   (e.g.`^\w+$`)  | Regular expression validation
param |   err    |      no      |(e.g.`wrong password format`)| Custom error message

NOTES:
* `param:"-"` means ignore
* Encountered untagged exportable anonymous structure field, automatic recursive resolution

- Field-Types

base    |   slice    | special
--------|------------|------------
string  |  []string  | [][]byte
byte    |  []byte    | [][]uint8
uint8   |  []uint8   | struct
bool    |  []bool    |
int     |  []int     |
int8    |  []int8    |
int16   |  []int16   |
int32   |  []int32   |
int64   |  []int64   |
uint8   |  []uint8   |
uint16  |  []uint16  |
uint32  |  []uint32  |
uint64  |  []uint64  |
float32 |  []float32 |
float64 |  []float64 |


#### Test

```go
package binder_test

import (
	"testing"
	"time"

	tp "github.com/henrylee2cn/teleport"
	binder "github.com/henrylee2cn/tp-ext/plugin-binder"
)

type Args struct {
	A int
	B int `param:"<range:1:>"`
}

type P struct{ tp.PullCtx }

func (p *P) Divide(args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

func TestBinder(t *testing.T) {
	bplugin := binder.NewStructArgsBinder(10001, "error request parameter")
	srv := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"})
	srv.PullRouter.Reg(new(P), bplugin)
	go srv.Listen()
	time.Sleep(time.Second)

	cli := tp.NewPeer(tp.PeerConfig{})
	sess, err := cli.Dial(":9090")
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
	rerr = sess.Pull("/p/divide", &Args{
		A: 10,
		B: 0,
	}, &reply).Rerror()
	if rerr == nil {
		t.Fatal(rerr)
	}
	t.Logf("10/0 error:%v", rerr)
}
```

```sh
go test -v -run=TestBinder
```