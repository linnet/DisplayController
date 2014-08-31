package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

var msgId int

func main() {
	var displays [4]chan string
	var displayResponses [4]chan string
	for i := 0; i < 4; i++ {
		displays[i] = make(chan string, 10)
		displayResponses[i] = make(chan string)
	}

	go listenForResponse(displayResponses)

	for i := 0; i < len(displays); i++ {
		go sendToDisplay(i+1, displays[i], displayResponses[i])
	}

	for {
		sendCommandToRandomDisplay(displays)
		//sendCommandToAllDisplays(displays)
		//time.Sleep(1 * time.Second)
	}
}

func sendCommandToRandomDisplay(displays [4]chan string) {
	displayId := rand.Intn(4)
	display := displays[displayId]

	display <- fmt.Sprintf("Msg %d to random display", msgId)

	msgId++
}

func sendCommandToAllDisplays(displays [4]chan string) {
	for _, display := range displays {
		display <- fmt.Sprintf("Show text number %d", msgId)
		display <- fmt.Sprintf("Other cmd %d", msgId)
	}
	msgId++
}

func listenForResponse(displayResponses [4]chan string) {
	addr, err := net.ResolveUDPAddr("udp4", ":60001")
	checkError(err)

	sock, err := net.ListenUDP("udp4", addr)
	checkError(err)

	fmt.Printf("Controller listening on %s\n", addr)
	for {
		readFromUdp(sock, displayResponses)
	}

}

func sendToDisplay(displayId int, cmds chan string, responses chan string) {
	serverAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", 6900+displayId))
	checkError(err)

	con, err := net.DialUDP("udp4", nil, serverAddr)
	checkError(err)
	defer con.Close()

	for {
		cmd := <-cmds

		if _, err := con.Write([]byte(cmd)); err != nil {
			fmt.Printf("%d: Error sending to display\n", displayId)
		}

		fmt.Printf("%d: Sent command %s\n", displayId, cmd)

		select {
		case status := <-responses:
			fmt.Printf("%d: Status %s\n", displayId, status)
		case <-time.After(time.Millisecond * 200):
			fmt.Printf("%d: Timed out\n", displayId)
		}
	}
}

func readFromUdp(sock *net.UDPConn, displayResponses [4]chan string) {
	var buf [1024]byte
	rlen, _, err := sock.ReadFromUDP(buf[:])
	checkError(err)

	//	fmt.Printf("%d bytes received from %s: %s\n", rlen, remote, buf)

	displayReplying := buf[0] - '1'
	status := string(buf[2:rlen])

	displayResponses[displayReplying] <- status
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
