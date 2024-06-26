package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		panic(err)
	}

	connect := []byte{
		0b00010000,
		11 + 3 + 2 + 4,
		0, 0, 'M', 'Q', 'T', 'T', 0, 0, 0, 0, 3,
		1, 1, 1,
		0, 4,
		'h', 'e', 'n', 'k',
	}

	n, err := conn.Write(connect)
	if err != nil {
		panic(err)
	}
	if n != len(connect) {
    panic(fmt.Sprintf("connect n: %d", n))
	}

  publish := []byte{
		0b00110000,
		7 + 1 + 4, // Remaining length
    0, 5, 't', 'e', '/', 's', 't', // Topic name
    0, // Properties length
    't', 'e', 's', 't',
  }

	n, err = conn.Write(publish)
	if err != nil {
		panic(err)
	}
	if n != len(publish) {
    panic(fmt.Sprintf("publish n: %d", n))
	}

  subscribe := []byte{
		0b10000000,
		2 + 1 + 8, // Remaining length
    0x80, 0x08, // Packet identifer
    0, // Properties length
    0, 5, 't', 'e', '/', 's', 't', 0, // Topic filter + subscription options
  }

	n, err = conn.Write(subscribe)
	if err != nil {
		panic(err)
	}
	if n != len(subscribe) {
    panic(fmt.Sprintf("subscribe n: %d", n))
	}

  puback := make([]byte, 100)
  n, err = conn.Read(puback)
  fmt.Println("puback", puback)

  fmt.Println("done")
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
