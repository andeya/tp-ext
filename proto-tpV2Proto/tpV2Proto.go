// Package tpV2Proto compatible teleport v2 protocol
package tpV2Proto

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"strconv"
	"sync"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/codec"
	socket "github.com/henrylee2cn/teleport/socket"
	"github.com/henrylee2cn/teleport/utils"
	"github.com/henrylee2cn/tp-ext/proto-tpV2Proto/pb"
)

// NewProtoFunc Returns teleport v2 protocol
var NewProtoFunc = func(rw io.ReadWriter) socket.Proto {
	var (
		v2ProtoReadBufioSize      int
		readBufferSize, isDefault = socket.ReadBuffer()
	)
	if isDefault {
		v2ProtoReadBufioSize = 1024 * 4
	} else if readBufferSize == 0 {
		v2ProtoReadBufioSize = 1024 * 35
	} else {
		v2ProtoReadBufioSize = readBufferSize / 2
	}
	return &tpV2Proto{
		id:   '2',
		name: "tpV2",
		r:    bufio.NewReaderSize(rw, v2ProtoReadBufioSize),
		w:    rw,
	}
}

// tpV2Proto compatible socket communication protocol.
type tpV2Proto struct {
	id   byte
	name string
	r    io.Reader
	w    io.Writer
	rMu  sync.Mutex
}

// Version returns the protocol's id and name.
func (t *tpV2Proto) Version() (byte, string) {
	return t.id, t.name
}

func (t *tpV2Proto) Pack(p *socket.Packet) error {
	bb := utils.AcquireByteBuffer()
	defer utils.ReleaseByteBuffer(bb)
	seq, err := strconv.ParseUint(p.Seq(), 10, 64)
	if err != nil {
		return err
	}
	header := &pb.Header{
		Seq:  seq,
		Type: int32(p.Ptype()),
		Uri:  p.Uri(),
	}

	if p.Meta().Len() > 0 {
		// Get rerror struct
		newRerr := tp.NewRerrorFromMeta(p.Meta())
		if newRerr != nil {
			header.StatusCode = newRerr.Code
			header.Status = newRerr.Message
		}
	}

	// write header
	err = t.writeHeader(bb, header)
	if err != nil {
		return err
	}

	// write body
	err = t.writeBody(bb, p)
	if err == nil {
		// real write
		_, err = t.w.Write(bb.B)
	}
	return err
}

func (t *tpV2Proto) writeHeader(bb *utils.ByteBuffer, header *pb.Header) error {
	// fake size
	err := binary.Write(bb, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}

	headerBytes, err := header.Marshal()
	if err != nil {
		return err
	}
	// write headerLength
	binary.BigEndian.PutUint32(bb.B, uint32(len(headerBytes)+1))

	// write headerCodecId
	err = bb.WriteByte(codec.ID_PROTOBUF)
	if err != nil {
		return err
	}

	// write header
	_, err = bb.Write(headerBytes)
	return err
}

func (t *tpV2Proto) writeBody(bb *utils.ByteBuffer, p *socket.Packet) error {
	var (
		bodyBytes []byte
		err       error
	)

	switch bo := p.Body().(type) {
	case []byte:
		bodyBytes = bo
	case *[]byte:
		bodyBytes = *bo
	default:
		bodyBytes, err = json.Marshal(bo)
		if err != nil {
			return err
		}
	}

	var (
		bodyLength     = uint32(len(bodyBytes))
		hasBodyCodecId bool
	)

	// get bodyCodecId
	bodyCodec, err := codec.Get(p.BodyCodec())
	if err != nil {
		return err
	}
	if bodyCodec.Id() != 0 {
		bodyLength += 1
		hasBodyCodecId = true
	}

	// write bodyLength
	err = binary.Write(bb, binary.BigEndian, bodyLength)
	if err != nil {
		return err
	}

	if hasBodyCodecId {
		// write bodyCodecId
		err = bb.WriteByte(bodyCodec.Id())
		if err != nil {
			return err
		}
	}

	// write body
	_, err = bb.Write(bodyBytes)
	return err
}

func (t *tpV2Proto) Unpack(p *socket.Packet) error {
	t.rMu.Lock()
	defer t.rMu.Unlock()

	bb := utils.AcquireByteBuffer()
	defer utils.ReleaseByteBuffer(bb)
	// read header
	err := t.readHeader(bb, p)
	if err != nil {
		return err
	}

	// read body
	return t.readBody(bb, p)
}

func (t *tpV2Proto) readHeader(bb *utils.ByteBuffer, p *socket.Packet) error {
	var headerLength uint32
	err := binary.Read(t.r, binary.BigEndian, &headerLength)
	if err != nil {
		return err
	}
	if err = p.SetSize(headerLength); err != nil {
		return err
	}
	bb.ChangeLen(int(headerLength))
	_, err = io.ReadFull(t.r, bb.B)
	if err != nil {
		return err
	}

	header := &pb.Header{}
	err = header.Unmarshal(bb.B[1:headerLength])
	if err != nil {
		return err
	}
	p.SetSeq(strconv.FormatUint(header.Seq, 10))
	p.SetPtype(byte(header.Type))
	p.SetUri(header.Uri)

	if header.StatusCode != 0 {
		newRerr := tp.Rerror{
			Code:    header.StatusCode,
			Message: header.Status,
		}
		metaByts, err := newRerr.MarshalJSON()
		if err != nil {
			return err
		}
		p.Meta().ParseBytes(metaByts)
	}

	return err
}

func (t *tpV2Proto) readBody(bb *utils.ByteBuffer, p *socket.Packet) error {
	bb.Reset()
	var bodyLength uint32
	err := binary.Read(t.r, binary.BigEndian, &bodyLength)
	if err != nil {
		return err
	}
	if err = p.SetSize(bodyLength); err != nil {
		return err
	}
	bb.ChangeLen(int(bodyLength))

	_, err = io.ReadFull(t.r, bb.B)
	if err != nil {
		return err
	}

	bbLen := len(bb.B)
	// Array length judgment
	if bbLen == 0 {
		return nil
	}
	p.SetBodyCodec(bb.B[0])
	if bbLen > 1 {
		err = p.UnmarshalBody(bb.B[1:])
	}
	return err
}
