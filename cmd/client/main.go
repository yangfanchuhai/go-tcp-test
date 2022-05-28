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
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			defer wg.Done()
			startClient(fmt.Sprintf("conn%d", i))
		}(i + 1)
	}
	wg.Wait()
}

func startClient(connName string)  {
	done := make(chan struct{})
	quit := make(chan struct{})
	conn, err := net.Dial("tcp", "localhost:8888")

	if err != nil {
		fmt.Println("connect failed", err)
		return
	}

	defer conn.Close()
	codec := frame.NewCodec()
	var counter = 0
	//处理和服务器的交互
	go func() {
		for  {
			select {
			case <-quit:
				done <- struct{}{}
				fmt.Printf("conne %s processor exit\n", connName)
				return
			default:

			}
			conn.SetReadDeadline(time.Now().Add(time.Second))
			fPayload, err := codec.Decode(conn, "")
			if err != nil {
				if e, ok := err.(net.Error); ok  {
					if e.Timeout() {
						//fmt.Println("read timeout retry")
						continue
					}
				}
				panic(err)
			}

			p := &packet.PacPayload{}
			err = p.Decode(fPayload)

			if err != nil {
				fmt.Println("decode packet failed", err)
				panic(err)
			}
			counter ++
			fmt.Printf("%s receive from server msg=%s counter=%d\n", connName, p.Payload, counter)
		}

	}()

	p := packet.PacPayload{CommType: packet.CommConn, Payload: connName + "hello"}
	b, err := p.Encode()

	if err != nil {
		fmt.Println("encode connect packet failed", err)
		panic(err)
	}

	err = codec.Encode(conn, b)

	if err != nil {
		fmt.Println("connect failed", err)
		panic(err)
	}
	fmt.Printf("%s connect success\n", connName)
	for i := 0; i < 10; i++  {
		p := packet.PacPayload{CommType: packet.CommSubmit, Payload: fmt.Sprintf("%smsg%d", connName, i)}
		b, err := p.Encode()
		if err != nil {
			fmt.Println("encode submit packet failed", err)
			panic(err)
		}
		err = codec.Encode(conn, b)
		if err != nil {
			fmt.Println("client submit failed", err)
			panic(err)
		}
		fmt.Printf("%s submit %s success\n", connName, p.Payload)
		time.Sleep(time.Second)
	}
	fmt.Printf("%s send all msg success\n", connName)
	for {
		//fmt.Printf("%s sleep, counter=%d\n", connName, counter)
		if counter >= 11 {
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
	quit <- struct{}{}
	<-done
	fmt.Printf("%s client exit ok counter=%d\n", connName, counter)
}