package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	var listenPort int = 69
	// var sessions []tftp.TftpSession = make([]tftp.TftpSession, 5)

	log.Println("tftpd: starting on port: ", listenPort)
	// listen to incoming udp packets
	pc, err := net.ListenPacket("udp", ":69")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("tftpd: socket correctly open")
	defer pc.Close()

	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		go serve(pc, addr, buf[:n])
	}

}

func serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	fmt.Println(addr)
}
