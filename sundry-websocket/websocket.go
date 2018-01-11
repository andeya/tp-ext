// Websocket is an extension package that makes the Teleport framework compatible
// with websocket protocol as specified in RFC 6455.
//
// Copyright 2018 HenryLee. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package websocket

import (
	"io"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strings"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/henrylee2cn/teleport/utils"
	ws "github.com/henrylee2cn/tp-ext/sundry-websocket/websocket"
)

func NewDialPlugin(pattern string) tp.PostDialPlugin {
	pattern = path.Join("/", strings.TrimRight(pattern, "/"))
	if pattern == "/" {
		pattern = ""
	}
	return &clientPlugin{pattern}
}

type clientPlugin struct {
	pattern string
}

func (*clientPlugin) Name() string {
	return "websocket"
}

func (c *clientPlugin) PostDial(sess tp.PreSession) *tp.Rerror {
	var location, origin string
	if sess.Peer().TlsConfig() == nil {
		location = "ws://" + sess.RemoteIp() + c.pattern
		origin = "ws://" + sess.LocalIp() + c.pattern
	} else {
		location = "wss://" + sess.RemoteIp() + c.pattern
		origin = "wss://" + sess.LocalIp() + c.pattern
	}
	cfg, err := ws.NewConfig(location, origin)
	if err != nil {
		return tp.NewRerror(tp.CodeDialFailed, "upgrade to websocket failed", err.Error())
	}
	conn, err := ws.NewClient(cfg, sess.Conn())
	if err != nil {
		return tp.NewRerror(tp.CodeDialFailed, "upgrade to websocket failed", err.Error())
	}
	sess.ResetConn(conn, NewWsProtoFunc(sess.GetProtoFunc()))
	return nil
}

// NewServeHandler creates a websocket handler.
func NewServeHandler(peer *tp.Peer, handshake func(*ws.Config, *http.Request) error, protoFunc ...socket.ProtoFunc) http.Handler {
	w := &serverHandler{
		peer:      peer,
		Server:    new(ws.Server),
		protoFunc: NewWsProtoFunc(protoFunc...),
	}
	var scheme string
	if peer.TlsConfig() == nil {
		scheme = "ws"
	} else {
		scheme = "wss"
	}
	if handshake != nil {
		w.Server.Handshake = func(cfg *ws.Config, r *http.Request) error {
			cfg.Origin = &url.URL{
				Host:   r.RemoteAddr,
				Scheme: scheme,
			}
			return handshake(cfg, r)
		}
	} else {
		w.Server.Handshake = func(cfg *ws.Config, r *http.Request) error {

			cfg.Origin = &url.URL{
				Host:   r.RemoteAddr,
				Scheme: scheme,
			}
			return nil
		}
	}
	w.Server.Handler = w.handler
	w.Server.Config = ws.Config{
		TlsConfig: peer.TlsConfig(),
	}
	return w
}

type serverHandler struct {
	peer      *tp.Peer
	protoFunc socket.ProtoFunc
	*ws.Server
}

func (w *serverHandler) handler(conn *ws.Conn) {
	sess, err := w.peer.ServeConn(conn, w.protoFunc)
	if err != nil {
		tp.Errorf("serverHandler: %v", err)
	}
	for sess.Health() {
		runtime.Gosched()
	}
}

// NewWsProtoFunc wraps a protocol to a new websocket protocol.
func NewWsProtoFunc(subProto ...socket.ProtoFunc) socket.ProtoFunc {
	return func(rw io.ReadWriter) socket.Proto {
		conn, ok := rw.(*ws.Conn)
		if !ok {
			tp.Warnf("connection does not support websocket protocol")
			if len(subProto) > 0 {
				return subProto[0](rw)
			} else {
				return socket.DefaultProtoFunc()(rw)
			}
		}
		buf := &utils.ByteBuffer{}
		p := &wsProto{
			id:   'w',
			name: "websocket",
			conn: conn,
			buf:  buf,
		}
		if len(subProto) > 0 {
			p.subProto = subProto[0](buf)
		} else {
			p.subProto = socket.DefaultProtoFunc()(buf)
		}
		return p
	}
}

type wsProto struct {
	id       byte
	name     string
	conn     *ws.Conn
	subProto socket.Proto
	buf      *utils.ByteBuffer
}

// Version returns the protocol's id and name.
func (w *wsProto) Version() (byte, string) {
	return w.id, w.name
}

// Pack writes the Packet into the connection.
// Note: Make sure to write only once or there will be package contamination!
func (w *wsProto) Pack(p *socket.Packet) error {
	w.buf.Reset()
	err := w.subProto.Pack(p)
	if err != nil {
		return err
	}
	return ws.Message.Send(w.conn, w.buf.Bytes())
}

// Unpack reads bytes from the connection to the Packet.
// Note: Concurrent unsafe!
func (w *wsProto) Unpack(p *socket.Packet) error {
	w.buf.Reset()
	err := ws.Message.Receive(w.conn, &w.buf.B)
	if err != nil {
		return err
	}
	return w.subProto.Unpack(p)
}
