package main

import (
	"log"
	"net"
	"tftpd/internal/tftp"
)

func main() {
	addr := net.UDPAddr{Port: 69, IP: net.ParseIP("127.0.0.1")}
	sessions := make(map[string]chan tftp.UDPPacket)
	control := make(chan string, 20)
	input := make(chan tftp.UDPPacket, 20)
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("tftpd: starting on port: ", addr)
	// listen to incoming udp packets
	udpConn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("tftpd: socket correctly open")
	defer udpConn.Close()
	go readUDP(udpConn, input)
	for {
		select {
		case udpPacket := <-input:
			handleRequest(udpConn, udpPacket, control, &sessions)
		case dead := <-control:
			log.Println("DELETING: ", dead)
			delete(sessions, dead)
		}
	}

}

func readUDP(udpConn *net.UDPConn, input chan tftp.UDPPacket) {
	buf := make([]byte, 1024)
	for {
		n, remote, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		input <- tftp.UDPPacket{Socket: udpConn, Peer: remote, Body: buf, BodySize: n}
	}
}

func handleRequest(pc *net.UDPConn, udpPacket tftp.UDPPacket, controlChannel chan string, sessions *map[string]chan tftp.UDPPacket) {
	addr := udpPacket.Peer
	buf := udpPacket.Body[:udpPacket.BodySize]
	currentSession, presence := (*sessions)[addr.IP.String()]

	if presence == true {
		// log.Println("Have Session for: ", addr)
		currentSession <- tftp.UDPPacket{Socket: pc, Peer: addr, Body: buf}
	} else {
		// log.Println("Missing Sessions for: ", addr, "\nRequested: \n", buf)
		newSession := &tftp.Session{Peer: addr}
		newSession.HandleRequest(pc, buf)
		channel := make(chan tftp.UDPPacket, 10)
		(*sessions)[addr.IP.String()] = channel
		go tftp.RunSession(addr, channel, controlChannel, newSession)
	}
}
