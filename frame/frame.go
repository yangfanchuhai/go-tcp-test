package frame

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type FPayload []byte

type StreamFrameCodec interface {
	Encode(io.Writer, FPayload) error
	Decode(io.Reader, string) (FPayload, error)
}

type myFrameCodec struct {}

var ErrShortRead = errors.New("short read")
var ErrShortWrite = errors.New("short write")

func NewCodec() StreamFrameCodec {
	return &myFrameCodec{}
}

func (codec *myFrameCodec ) Encode(w io.Writer, fPayload FPayload) error  {
	fmt.Printf("encode frame payload:%X\n", fPayload)
	var buff bytes.Buffer
	var totalLen = int32(len(fPayload)) + 4
	fmt.Printf("encode frame len of f %d totallen:%d\n", len(fPayload), totalLen)
	err := binary.Write(w, binary.BigEndian, &totalLen)
	binary.Write(&buff, binary.BigEndian, &totalLen)
	if err != nil {
		return err
	}
	
	n, err := w.Write(fPayload)
	buff.Write(fPayload)
	if err != nil {
		return err
	}

	if n != len(fPayload) {
		return ErrShortWrite
	}

	buffB := buff.Bytes()
	fmt.Printf("client write %X %X %s\n", buffB[0:4], buffB[4:8], buffB[8:])
	return nil
}

func (codec *myFrameCodec ) Decode(r io.Reader, connName string) (FPayload, error) {
	var totalLen int32
	var buff bytes.Buffer
	totalLenB, err := readN(connName, r, 4)
	buff.Write(totalLenB)
	//如果读取长度失败，包括超时，交由外层重试
	if err != nil {
		return nil, err
	}
	totalLen = int32(binary.BigEndian.Uint32(totalLenB))

	//如果先读取了长度，这里一定要继续，直到读完，直到出现非超时错误
	for {
		if conn, ok := r.(net.Conn); ok {
			//要重新设置超时时间
			conn.SetReadDeadline(time.Now().Add(time.Second))
		}
		p, err := readN(connName, r, int(totalLen - 4))
		if err != nil {
			if e, ok := err.(net.Error); ok {
				if e.Timeout() {
					continue
				}
			}
			return nil, err
		}
		buff.Write(p)
		buffB := buff.Bytes()
		fmt.Printf("%s read from peer %X %X %s\n", connName, buffB[0:4], buffB[4:8], buffB[8:])
		return p, nil
	}
}

//读取N个字节
func readN(connName string, r io.Reader, n int) ([]byte, error)  {
	var leftLen = n
	b := make([]byte, 0, n)
	//var retryTimes = 0
	for leftLen > 0 {
		tempByte := make([]byte, leftLen)
		readN, err := r.Read(tempByte)
		//retryTimes ++
		if err != nil {
			if e, ok := err.(net.Error); ok {
				if e.Timeout() {
					if leftLen == n && readN == 0 {
						//如果一个字节都未读，允许退出，在外层重试
						//fmt.Printf("read n timeout and max time, readN=%d error\n", readN)
						return nil, err
					}

					//如果已经读了部分字节，则阻塞一直读完
					//fmt.Printf("read n timeout, readN=%d retry\n", readN)
					time.Sleep(time.Millisecond * 10)
					if conn, ok := r.(net.Conn); ok {
						conn.SetReadDeadline(time.Now().Add(time.Second))
					}
					continue
				}
			}
			return nil, err
		}

		if readN > 0 {
			leftLen -= readN
			rB := tempByte[0:readN]
			b = append(b, rB...)
		}
	}
	fmt.Printf("%s frame decode read bytes %X\n", connName, b)
	return b, nil
}
