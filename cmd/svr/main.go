package main

import (
	"fmt"
	"github.com/yangfanchuhai/go-tcp-test/frame"
	"github.com/yangfanchuhai/go-tcp-test/packet"
	"net"
	"sync"
	"time"
)

func main() {
	listen, err := net.Listen("tcp", ":8888")

	if err != nil  {
		fmt.Println("server listen failed", err)
	}
	fmt.Println("server listen success")
	for  {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accept conn failed", err)
		}
		fmt.Println("accept conn success")

		go processConn(conn)
	}
}

//处理单个连接
func processConn(conn net.Conn)  {
	defer conn.Close()
	codec := frame.NewCodec()
	var connName = ""
	for  {
		conn.SetReadDeadline(time.Now().Add(time.Second))
		f, err := codec.Decode(conn, connName)
		if err != nil {
			if e, ok := err.(net.Error); ok {
				if e.Timeout() {
					//fmt.Println("wait request from server timeout, retry")
					continue
				}
			}
			fmt.Println(connName + "server decode frame failed", err)
			break
		}
		fmt.Printf("conn %s receive from client bytes %X", connName, f)
		connName, err = processPacket(conn, codec, f)
		if err != nil {
			fmt.Println("process packet failed", err)
			break
		}
	}
}

var packetPool = sync.Pool{New: func() interface{} {
		return packet.PacPayload{}
	},
}

func processPacket(conn net.Conn, codec frame.StreamFrameCodec, f frame.FPayload) (string, error)  {
	p := packetPool.Get().(packet.PacPayload)
	err := p.Decode(f)
	if err != nil {
		fmt.Println("decode packet failed", err)
		return "", err
	}
	fmt.Printf("receive request from client:%s\n", p.Payload)
	var connName = p.Payload[0:5]
	switch p.CommType {
	case packet.CommConn:
		p.CommType = packet.CommConnAck
		p.Payload = fmt.Sprintf("ok")
		//p = packet.PacPayload{CommType: packet.CommConnAck, Payload: fmt.Sprintf("ok")}
	case packet.CommSubmit:
		p.CommType = packet.CommSubmitAck
		p.Payload = fmt.Sprintf("return%s", p.Payload)
		//p = packet.PacPayload{CommType: packet.CommSubmitAck, Payload: fmt.Sprintf("return%s", p.Payload)}
	}

	b, err := p.Encode()

	if err != nil {
		fmt.Println("encode packet failed", err)
		return "", err
	}
	err = codec.Encode(conn, b)
	if err != nil {
		fmt.Println("send reply to client failed")
		return "", err
	}
	fmt.Printf("send reply to client success %s\n", p.Payload)
	packetPool.Put(p)
	return connName, nil
}