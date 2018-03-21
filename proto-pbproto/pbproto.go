// Package pbproto is implemented PROTOBUF socket communication protocol.
package pbproto

import (
	"bufio"
	"encoding/binary"
	"io"
	"sync"

	"github.com/henrylee2cn/teleport/codec"
	"github.com/henrylee2cn/teleport/socket"
	"github.com/henrylee2cn/teleport/utils"
	"github.com/henrylee2cn/tp-ext/proto-pbproto/pb"
)

// NewPbProtoFunc is creation function of PROTOBUF socket protocol.
var NewPbProtoFunc = func(rw io.ReadWriter) socket.Proto {
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
	return &pbproto{
		id:   'p',
		name: "protobuf",
		r:    bufio.NewReaderSize(rw, readBufioSize),
		w:    rw,
	}
}

type pbproto struct {
	id   byte
	name string
	r    *bufio.Reader
	w    io.Writer
	rMu  sync.Mutex
}

// Version returns the protocol's id and name.
func (pp *pbproto) Version() (byte, string) {
	return pp.id, pp.name
}

// Pack writes the Packet into the connection.
// Note: Make sure to write only once or there will be package contamination!
func (pp *pbproto) Pack(p *socket.Packet) error {
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

	b, err := codec.ProtoMarshal(&pb.Format{
		Seq:       p.Seq(),
		Ptype:     int32(p.Ptype()),
		Uri:       p.Uri(),
		Meta:      p.Meta().QueryString(),
		BodyCodec: int32(p.BodyCodec()),
		Body:      bodyBytes,
		XferPipe:  p.XferPipe().Ids(),
	})
	if err != nil {
		return err
	}

	p.SetSize(uint32(len(b)))
	var all = make([]byte, p.Size()+4)
	binary.BigEndian.PutUint32(all, p.Size())
	copy(all[4:], b)
	_, err = pp.w.Write(all)
	return err
}

// Unpack reads bytes from the connection to the Packet.
// Note: Concurrent unsafe!
func (pp *pbproto) Unpack(p *socket.Packet) error {
	pp.rMu.Lock()
	defer pp.rMu.Unlock()
	var size uint32
	err := binary.Read(pp.r, binary.BigEndian, &size)
	if err != nil {
		return err
	}
	if err = p.SetSize(size); err != nil {
		return err
	}
	if p.Size() == 0 {
		return nil
	}

	bb := utils.AcquireByteBuffer()
	defer utils.ReleaseByteBuffer(bb)
	bb.ChangeLen(int(p.Size()))
	_, err = io.ReadFull(pp.r, bb.B)
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
	err = p.UnmarshalBody(bodyBytes)
	return err
}
