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

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/henrylee2cn/teleport/utils"
	ws "github.com/henrylee2cn/tp-ext/mod-websocket/websocket"
)

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
