package main

import (
	"net"
	"sync"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	bin := []byte{
		0b00010000,
		11 + 3 + 2 + 4,
		0, 0, 'M', 'Q', 'T', 'T', 0, 0, 0, 0, 3,
		1, 1, 1,
		0, 4,
		'h', 'e', 'n', 'k',
	}

	n, err := conn.Write(bin)
	if err != nil {
		panic(err)
	}
	if n != len(bin) {
		panic(n)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
