package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	CommConn = iota + 0x01
	CommSubmit
)

const (
	CommConnAck = iota + 0x80
	CommSubmitAck
)

type PacPayload struct {
	CommType int32
	Payload string
}

type Packet interface {
	Decode([]byte) error
	Encode() ([]byte, error)
}



func (p *PacPayload) Decode(b []byte) error  {
	if b == nil || len(b) <= 4 {
		return errors.New("empty")
	}

	buff := bytes.NewBuffer(b)

	var commType int32
	err := binary.Read(buff, binary.BigEndian, &commType)
	if err != nil {
		return err
	}

	p.CommType = commType

	payloadB := make([]byte, len(b) - 4)

	n, err := buff.Read(payloadB)
	if err != nil {
		return err
	}

	if n < len(b) - 4 {
		return errors.New("short read")
	}
	p.Payload = string(payloadB)
	return nil
}

func (p *PacPayload) Encode() ([]byte, error) {
	buff := &bytes.Buffer{}
	binary.Write(buff, binary.BigEndian, &p.CommType)
	buff.Write([]byte(p.Payload))
	return buff.Bytes(), nil
}

