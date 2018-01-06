## heartbeat

A generic timing heartbeat package.

- Demo

```go
srv := tp.NewPeer(tp.PeerConfig{ListenAddress: ":9090"})
heartbeat.WithPong(srv)
go srv.Listen()

cli := tp.NewPeer(tp.PeerConfig{})
heartbeat.WithPing(cli, time.Second)
cli.Dial(":9090")
time.Sleep(time.Second * 10)
```