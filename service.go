package main

import (
	"./Msg"
	"flag"
	"net"
	"strconv"
)

func main() {
	affectedNodePtr := flag.Int("node", 1, "affected node")
	actionPtr := flag.String("action", "send", "action: send or drop")
	destinationPortPtr := flag.Int("dst", 2, "destination port for send action")
	flag.Parse()
	udp := "udp"

	from, _ := net.ResolveUDPAddr(udp, "127.0.0.1:50000")
	to, _ := net.ResolveUDPAddr(udp, "127.0.0.1:"+strconv.Itoa(40000+*affectedNodePtr))
	connection, _ := net.ListenUDP(udp, from)
    msg := Msg.Msg {
        Type: *actionPtr,
        Dst: *destinationPortPtr,
        Data: "hi there",
        Src: 0,
        Ack: false }
    jsonMsg := msg.ToJson()
    buf := []byte(jsonMsg)
	connection.WriteToUDP(buf, to)
}
