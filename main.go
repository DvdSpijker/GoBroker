package main

import (
  "fmt"
	"github.com/DvdSpijker/GoBroker/packet"
)

func main() {

  fmt.Println("packet tests")

  // packet := packet.ControlPacket{}
  //
  // bytes := []byte{18, 129, 1}
  //
  // n, err := packet.Decode(bytes)
  //
  // fmt.Println(n, err, packet)

  connectPacket := packet.ConnectPacket{}

  bytes := []byte{}

  n, err := connectPacket.Decode(bytes)

  fmt.Println(n, err, connectPacket)
}
