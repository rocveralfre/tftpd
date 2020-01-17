package tftp

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

//Session :
//Due to the stateless nature of UDP
//We have to create a State using this struct
type Session struct {
	Peer          *net.UDPAddr
	FileName      string
	CurrentOffset int64
	CurrentChunk  []byte
	FileReader    *os.File
	TotalChunks   int64
	FileSize      int64
	Mode          string
	ReqType       int
}

// CreateSession : returns a new valid object
func CreateSession(peer *net.UDPAddr) *Session {
	return &Session{Peer: peer, ReqType: 0}
}

// HandleRequest : handles binary requests from peers
func (session *Session) HandleRequest(pc *net.UDPConn, message []byte) {
	if len(message) > 2 {
		var opCode uint16 = binary.BigEndian.Uint16(message[:2])
		var payload []byte
		switch opCode {
		case 0x01:
			fmt.Println("READ REQUEST")
			session.FileName, payload = getFileName(message[2:])
			session.Mode = getMode(payload)
			session.ReqType = 1
			session.CurrentOffset = 0
			fmt.Println("REQUESTED:", session.FileName, "\nMode:", session.Mode)
			readRequest(pc, session)
		case 0x02:
			fmt.Println("WRITE REQUEST")
		case 0x04:
			if ack(pc, binary.BigEndian.Uint16(message[2:]), session) {
				session.FileReader.Close()
			}
		default:
			fmt.Println("Unknown OP Code:", opCode, "\nBody:", message)
		}
	}
}

func ack(pc *net.UDPConn, block uint16, session *Session) bool {
	if session.ReqType == 1 && session.CurrentOffset == int64(block) {

		session.CurrentOffset = session.CurrentOffset + 1
		n, err := session.FileReader.Read(session.CurrentChunk)
		fmt.Println("READ: ", n, " ERR:", err)
		var response []byte = []byte{0x00, 0x03, 0x00}
		offset := make([]byte, 8)
		binary.BigEndian.PutUint64(offset, uint64(session.CurrentOffset))
		response = append(response, offset...)
		response = append(response, session.CurrentChunk[:n]...)
		pc.WriteToUDP(response, session.Peer)
		time.Sleep(10000)
		if n < 512 {
			fmt.Println("DONE: 100%")
			return true
		} else {
			percentage := (session.CurrentOffset / session.TotalChunks) * 100
			fmt.Println(percentage, "%")
			return false
		}
	}
	return false
}

func readRequest(pc *net.UDPConn, session *Session) {
	if session.ReqType == 1 && session.CurrentOffset == 0 {
		if file, err := os.Stat(session.FileName); err == nil {
			session.FileSize = file.Size()
			if file.Size()%512 == 0 {
				session.TotalChunks = (file.Size() / 512)
			} else {
				session.TotalChunks = (file.Size() / 512) + 1
			}
			fmt.Println("File Exists", "\nSize: ", session.FileSize, "\nTotalChunks: ", session.TotalChunks)
			file, err := os.Open(session.FileName)
			if err != nil {
				var response []byte = []byte{0x00, 0x05, 0x00, 0x01}
				response = append(response, "FileNotFound"...)
				response = append(response, 0x00)
				pc.WriteToUDP(response, session.Peer)
			} else {
				session.CurrentChunk = make([]byte, 512)
				session.FileReader = file
				session.CurrentOffset = session.CurrentOffset + 1
				n, err := file.Read(session.CurrentChunk)
				fmt.Println("READ: ", n, " ERR:", err)
				var response []byte = []byte{0x00, 0x03, 0x00, 0x01}
				response = append(response, session.CurrentChunk[:n]...)
				pc.WriteToUDP(response, session.Peer)
				if n < 512 {
					fmt.Println("DONE: 100%")
				} else {
					percentage := (session.CurrentOffset / session.TotalChunks) * 100
					fmt.Println(percentage, "%")
				}
			}

		} else {
			var response []byte = []byte{0x00, 0x05, 0x00, 0x01}
			response = append(response, "FileNotFound"...)
			response = append(response, 0x00)
			pc.WriteToUDP(response, session.Peer)
		}
	}
}

func getFileName(binary []byte) (string, []byte) {
	var count int = 0

	for binary[count] != 0x00 {
		count++
	}
	return string(binary[:count]), binary[count+1:]
}

func getMode(binary []byte) string {
	var count int = 0

	for binary[count] != 0x00 {
		count++
	}

	return string(binary[:count])
}
