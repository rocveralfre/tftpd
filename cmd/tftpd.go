package main

import (
	"fmt"
	"log"
	"net"
	"tftpd/internal/tftp"
)

func main() {
	var addr net.UDPAddr = net.UDPAddr{Port: 69, IP: net.ParseIP("127.0.0.1")}
	var sessions map[string]tftp.Session = make(map[string]tftp.Session)

	log.Println("tftpd: starting on port: ", addr)
	// listen to incoming udp packets
	udpConn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("tftpd: socket correctly open")
	defer udpConn.Close()

	for {
		buf := make([]byte, 1024)
		n, remote, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		go handleRequest(udpConn, remote, buf[:n], &sessions)
	}

}

func handleRequest(pc *net.UDPConn, addr *net.UDPAddr, buf []byte, sessions *map[string]tftp.Session) {
	currentSession, presence := (*sessions)[addr.IP.String()]
	if presence == true {
		fmt.Println("Have Session for: ", addr)
		currentSession.HandleRequest(pc, buf)
	} else {
		fmt.Println("Missing Sessions for: ", addr, "\nRequested: \n", buf)
		var newSession *tftp.Session = tftp.CreateSession(addr)
		newSession.HandleRequest(pc, buf)
		(*sessions)[addr.IP.String()] = (*newSession)
	}
}
