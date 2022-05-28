package packet

import (
	"testing"
)

func TestEncode(t *testing.T) {
	packet := PacPayload{CommType: CommConn, Payload: "hello"}

	data, err := packet.Encode()

	if err != nil {
		t.Error(err)
	}

	if len(data) != 9 {
		t.Errorf("want 9 actual %d", len(data))
	}

	var payload = string(data[4:])
	if payload != "hello" {
		t.Errorf("want hello, actual %s", payload)
	}
}

func TestDecode(t *testing.T) {
	data := []byte{0x0, 0x0, 0x0, 0x1, 'h', 'e', 'l', 'l', 'o'}

	packet := &PacPayload{}

	packet.Decode(data)

	if packet.CommType != 1 {
		t.Errorf("want 1, actual %d", packet.CommType)
	}

	if packet.Payload != "hello" {
		t.Errorf("want hello, actual %s", packet.Payload)
	}
}
