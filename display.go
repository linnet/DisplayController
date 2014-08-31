package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

var displayId int
var shouldReply bool = true

func main() {
	displayId, _ = strconv.Atoi(os.Args[1])
	if len(os.Args) > 2 {
		shouldReply = false
		fmt.Println("Set to ignore requests")
	}

	port := 6900 + displayId

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", port))
	checkError(err)

	sock, err := net.ListenUDP("udp4", addr)
	checkError(err)

	fmt.Printf("Display %d listening on port %d\n", displayId, port)
	for {
		respondToCommand(sock)
	}
}

func respondToCommand(sock *net.UDPConn) {
	var buf [1024]byte
	rlen, remote, err := sock.ReadFromUDP(buf[:])
	checkError(err)

	fmt.Printf("%d bytes received from %s: %s\n", rlen, remote, buf)

	if shouldReply {
		replyToController("OK")
	} else {
		fmt.Println("Not replying")
	}
}

func replyToController(status string) {
	serverAddr, err := net.ResolveUDPAddr("udp4", ":60001")
	checkError(err)

	con, err := net.DialUDP("udp4", nil, serverAddr)
	checkError(err)
	defer con.Close()

	reply := fmt.Sprintf("%d:%s", displayId, status)
	_, err = con.Write([]byte(reply))
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
