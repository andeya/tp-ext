// Package pbSubProto is implemented PROTOBUF socket communication protocol.
package pbSubProto

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"

	"github.com/henrylee2cn/teleport/codec"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/henrylee2cn/teleport/utils"
	"github.com/henrylee2cn/tp-ext/mod-websocket/pbSubProto/pb"
)

// NewPbSubProtoFunc is creation function of JSON socket protocol.
var NewPbSubProtoFunc = func(rw io.ReadWriter) socket.Proto {
	var (
		readBufioSize             int
		readBufferSize, isDefault = socket.ReadBuffer()
	)
	if isDefault {
		readBufioSize = 1024 * 4
	} else if readBufferSize == 0 {
		readBufioSize = 1024 * 35
	} else {
		readBufioSize = readBufferSize / 2
	}
	return &pbSubProto{
		id:   'p',
		name: "protobuf",
		r:    bufio.NewReaderSize(rw, readBufioSize),
		w:    rw,
	}
}

type pbSubProto struct {
	id   byte
	name string
	r    *bufio.Reader
	w    io.Writer
	rMu  sync.Mutex
}

// Version returns the protocol's id and name.
func (psp *pbSubProto) Version() (byte, string) {
	return psp.id, psp.name
}

// Pack writes the Packet into the connection.
// Note: Make sure to write only once or there will be package contamination!
func (psp *pbSubProto) Pack(p *socket.Packet) error {
	// marshal body
	bodyBytes, err := p.MarshalBody()
	if err != nil {
		return err
	}
	// do transfer pipe
	bodyBytes, err = p.XferPipe().OnPack(bodyBytes)
	if err != nil {
		return err
	}
	// marshal transfer pipe ids
	var xferPipeIds = make([]int, p.XferPipe().Len())
	for i, id := range p.XferPipe().Ids() {
		xferPipeIds[i] = int(id)
	}
	xferPipeIdsBytes, err := json.Marshal(xferPipeIds)
	if err != nil {
		return err
	}

	b, err := codec.ProtoMarshal(&pb.Format{
		Seq:       p.Seq(),
		Ptype:     int32(p.Ptype()),
		Uri:       p.Uri(),
		Meta:      p.Meta().QueryString(),
		BodyCodec: int32(p.BodyCodec()),
		Body:      bodyBytes,
		XferPipe:  xferPipeIdsBytes,
	})
	if err != nil {
		return err
	}

	_, err = psp.w.Write(b)
	return err
}

// Unpack reads bytes from the connection to the Packet.
// Note: Concurrent unsafe!
func (psp *pbSubProto) Unpack(p *socket.Packet) error {
	psp.rMu.Lock()
	defer psp.rMu.Unlock()
	bb := utils.AcquireByteBuffer()
	defer utils.ReleaseByteBuffer(bb)
	var err error
	bb.B, err = ioutil.ReadAll(psp.r)
	if err != nil {
		return err
	}
	s := &pb.Format{}
	err = codec.ProtoUnmarshal(bb.B, s)
	if err != nil {
		return err
	}

	// read transfer pipe
	for _, r := range s.XferPipe {
		p.XferPipe().Append(r)
	}

	// read body
	p.SetBodyCodec(byte(s.BodyCodec))
	bodyBytes, err := p.XferPipe().OnUnpack(s.Body)
	if err != nil {
		return err
	}

	// read other
	p.SetSeq(s.Seq)
	p.SetPtype(byte(s.Ptype))
	p.SetUri(s.Uri)
	p.Meta().ParseBytes(s.Meta)

	// unmarshal new body
	err = p.UnmarshalNewBody(bodyBytes)
	return err
}
