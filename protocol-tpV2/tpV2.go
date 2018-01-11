// Compatible teleport v2 protocol
package tpV2

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"sync"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/codec"
	socket "github.com/henrylee2cn/teleport/socket"
	"github.com/henrylee2cn/teleport/utils"
)

type tpV2Proto struct {
	id   byte
	name string
	r    io.Reader
	w    io.Writer
	rMu  sync.Mutex
}

// newV2Proto Returns teleport v2 protocol
func newV2Proto(rw io.ReadWriter) socket.Proto {
	var (
		fastProtoReadBufioSize    int
		readBufferSize, isDefault = socket.ReadBuffer()
	)
	if isDefault {
		fastProtoReadBufioSize = 1024 * 4
	} else if readBufferSize == 0 {
		fastProtoReadBufioSize = 1024 * 35
	} else {
		fastProtoReadBufioSize = readBufferSize / 2
	}
	return &tpV2Proto{
		id:   '2',
		name: "tpV2",
		r:    bufio.NewReaderSize(rw, fastProtoReadBufioSize),
		w:    rw,
	}
}

var (
	errProtoUnmatch = errors.New("Mismatched protocol")
	lengthSize      = int64(binary.Size(uint32(0)))
)

// Version returns the protocol's id and name.
func (t *tpV2Proto) Version() (byte, string) {
	return t.id, t.name
}

func (t *tpV2Proto) Pack(p *socket.Packet) error {
	bb := utils.AcquireByteBuffer()
	defer utils.ReleaseByteBuffer(bb)
	// Get rerror struct
	newRerr := tp.NewRerrorFromMeta(p.Meta())
	header := &Header{
		Seq:        p.Seq(),
		Type:       int32(p.Ptype()),
		Uri:        p.Uri(),
		StatusCode: newRerr.Code,
		Status:     newRerr.Message,
	}

	// write header
	err := t.writeHeader(bb, header)
	if err != nil {
		return err
	}

	// write body
	switch bo := p.Body().(type) {
	case nil:
		err = binary.Write(bb, binary.BigEndian, uint32(1))
		if err == nil {
			bodyCodec, err := codec.Get(p.BodyCodec())
			if err != nil {
				return err
			}
			err = bb.WriteByte(bodyCodec.Id())
		}
	case []byte:
		err = t.writeBytesBody(bb, bo)
	case *[]byte:
		err = t.writeBytesBody(bb, *bo)
	default:
		err = t.writeBody(bb, p, bo)
	}

	if err == nil {
		// real write
		_, err = t.w.Write(bb.B)
		if err != nil {
			return err
		}
	}

	return err
}

func (t *tpV2Proto) writeHeader(bb *utils.ByteBuffer, header *Header) error {
	// fake size
	err := binary.Write(bb, binary.BigEndian, uint32(0))
	if err != nil {
		return err
	}

	// write headerCodecId
	err = bb.WriteByte(codec.ID_PROTOBUF)
	if err != nil {
		return err
	}

	headerBytes, err := header.Marshal()
	if err != nil {
		return err
	}
	// write headerLength
	err = binary.Write(bb, binary.BigEndian, uint32(len(headerBytes)))
	if err != nil {
		return err
	}

	// write header
	_, err = bb.Write(headerBytes)
	return err
}

func (t *tpV2Proto) readHeader(data []byte, p *socket.Packet) (string, error) {
	var headerLength uint32
	// read headerLength
	err := binary.Read(t.r, binary.BigEndian, &headerLength)
	if err != nil {
		return "", err
	}

	return "", err
}

func (t *tpV2Proto) writeBytesBody(bb *utils.ByteBuffer, body []byte) error {
	bodyLength := uint32(len(body))
	err := binary.Write(bb, binary.BigEndian, bodyLength)
	if err != nil {
		return err
	}

	_, err = bb.Write(body)
	return err
}

func (t *tpV2Proto) writeBody(bb *utils.ByteBuffer, p *socket.Packet, body interface{}) error {
	// write bodyCodecId
	bodyCodec, err := codec.Get(p.BodyCodec())
	if err != nil {
		return err
	}
	err = bb.WriteByte(bodyCodec.Id())

	// write bodyLength
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	err = binary.Write(bb, binary.BigEndian, uint32(len(bodyBytes)))
	if err != nil {
		return err
	}

	// write body
	_, err = bb.Write(bodyBytes)
	return err
}

func (t *tpV2Proto) Unpack(p *socket.Packet) error {
	bb := utils.AcquireByteBuffer()
	defer utils.ReleaseByteBuffer(bb)

	// read packet
	err := t.readPacket(bb, p)
	if err != nil {
		return err
	}

	header := &Header{}
	err = json.Unmarshal(bb.B, header)
	if err != nil {
		return err
	}

	return nil
}

func (t *tpV2Proto) readPacket(bb *utils.ByteBuffer, p *socket.Packet) error {
	t.rMu.Lock()
	defer t.rMu.Unlock()
	// size
	var size uint32
	err := binary.Read(t.r, binary.BigEndian, &size)
	if err != nil {
		return err
	}
	if err = p.SetSize(size); err != nil {
		return err
	}
	// protocol
	bb.ChangeLen(1024)
	_, err = t.r.Read(bb.B[:1])
	if err != nil {
		return err
	}
	if bb.B[0] != t.id {
		return errProtoUnmatch
	}

	_, err = io.ReadFull(t.r, bb.B)
	return err
}

func (t *tpV2Proto) readBody(data []byte, p *socket.Packet) error {
	return nil
}
