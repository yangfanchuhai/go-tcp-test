package frame

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestNewCodec(t *testing.T) {
	codec := NewCodec()

	if codec == nil {
		t.Errorf("nil codec")
	}
}

func TestEncode(t *testing.T) {
	codec := NewCodec()
	buf := make([]byte, 0, 128)
	bw := bytes.NewBuffer(buf)

	err := codec.Encode(bw, []byte("hello"))

	if err != nil {
		t.Errorf("want nil, actual %s", err.Error())
	}

	var totalLen int32
	err = binary.Read(bw, binary.BigEndian, &totalLen)

	if err != nil {
		t.Error(err.Error())
	}

	if totalLen != 9 {
		t.Errorf("want 9, actual %d", totalLen)
	}

	left := bw.Bytes()

	if string(left) != "hello" {
		t.Errorf("want hello, actual %s", string(left))
	}
}

func TestDecode(t *testing.T)  {
	codec := NewCodec()

	data := []byte{0x0, 0x0, 0x0, 0x9, 'h', 'e', 'l', 'l', 'o'}
	payload, err := codec.Decode(bytes.NewReader(data), "")

	if err != nil {
		t.Error(err.Error())
	}

	if string(payload) != "hello" {
		t.Errorf("want hello, actual %s", string(payload))
	}
}
