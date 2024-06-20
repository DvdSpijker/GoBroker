package main

import (
	"net"
	"sync"
  "fmt"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
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
		14, // rem len
    0, 5, 't', 'e', '/', 's', 't', // Topic name
    0, // Properties length
    0, 4, // Payload length
    't', 'e', 's', 't',
  }

	n, err = conn.Write(publish)
	if err != nil {
		panic(err)
	}
	if n != len(publish) {
    panic(fmt.Sprintf("publish n: %d", n))
	}

  fmt.Println("done")
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
