package tftp

import (
	"encoding/binary"
	"errors"
	"net"
)

// Request :
// Basic interface for handling read requests and write requests
type Request interface {
	ServeSession(peer *net.UDPAddr, inputChannel chan UDPPacket, control chan string)
	PrepareSession(requestMessage []byte, udpConn *net.UDPConn)
}

// NewSession :
// Factory Function
func NewSession(packet []byte, addr *net.UDPAddr) (Request, error) {
	opCode := binary.BigEndian.Uint16(packet[:2])
	switch opCode {
	case 0x01:
		return &ReadRequest{Peer: addr}, nil
	default:
		return nil, errors.New("Missing Request Type:" + string(packet[0]))
	}
}
