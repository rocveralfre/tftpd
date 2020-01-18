package tftp

import "net"

// UDPPacket : to distribute around
type UDPPacket struct {
	Peer     *net.UDPAddr
	Socket   *net.UDPConn
	Body     []byte
	BodySize int
}
