## heartbeat

A generic timing heartbeat plugin.

During a heartbeat, if there is no communication, send a heartbeat packet;
When the connection is idle more than 3 times the heartbeat time, take the initiative to disconnect.

### Usage

`import heartbeat "github.com/henrylee2cn/tp-ext/plugin-heartbeat"`

#### Test

```go
func TestHeartbeat1(t *testing.T) {
	srv := tp.NewPeer(
		tp.PeerConfig{ListenAddress: ":9090"},
		heartbeat.NewPong(time.Second),
	)
	go srv.Listen()
	time.Sleep(time.Second)

	cli := tp.NewPeer(
		tp.PeerConfig{},
		heartbeat.NewPing(time.Second),
	)
	cli.Dial(":9090")
	time.Sleep(time.Second * 10)
}
```

test command:

```sh
go test -v -run=TestHeartbeat1
```