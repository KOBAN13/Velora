package main

import (
	"Velora/server/pkg/packets"
	"fmt"

	"google.golang.org/protobuf/proto"
)

func main() {
	createPacket()

	var data = []byte{8, 69, 18, 14, 10, 12, 208, 159, 209, 128, 208, 184, 208, 178, 208, 181, 209, 130}
	var packet = &packets.Packet{}

	var err = proto.Unmarshal(data, packet)

	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println(packet)
}

func createPacket() *packets.Packet {
	var packet = &packets.Packet{
		SenderId: 69,
		Msg:      packets.NewChat("Привет"),
	}

	var data, err = proto.Marshal(packet)

	if err != nil {
		panic(err)
	}

	fmt.Println(data)

	return packet
}
