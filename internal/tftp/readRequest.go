package tftp

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"os"
)

// ReadRequest :
//Due to the stateless nature of UDP
//We have to create a State using this struct
type ReadRequest struct {
	id                 int
	Peer               *net.UDPAddr
	FileName           string
	CurrentChunkOffset int64
	CurrentChunk       []byte
	CurrentChunkSize   int
	FileReader         *os.File
	TotalChunks        int64
	FileSize           int64
	Mode               string
	ReqType            int
}

// PrepareSession :
// Creates the correct fields in the struct
func (session *ReadRequest) PrepareSession(requestMessage []byte, udpConn *net.UDPConn) {
	session.FileName, requestMessage = getFileName(requestMessage[2:])
	session.Mode = getMode(requestMessage)
	session.CurrentChunkOffset = 0
	session.CurrentChunk = make([]byte, 512)
	session.prepareFile(udpConn)
}

// ServeSession : we handle as process
func (session *ReadRequest) ServeSession(peer *net.UDPAddr, inputChannel chan UDPPacket, control chan string) {
	done := false
	for !done {
		select {
		case newPacket := <-inputChannel:
			// Handle Request returns true only when we're done
			done = session.HandleRequest(newPacket.Socket, newPacket.Body)
			// log.Println("Continue: ", done)
		}
	}
	control <- session.Peer.IP.String()
}

// HandleRequest : handles binary requests from peers
func (session *ReadRequest) HandleRequest(pc *net.UDPConn, message []byte) bool {
	if len(message) > 2 {
		opCode := binary.BigEndian.Uint16(message[:2])
		switch opCode {
		case 0x02:
			log.Println("WRITE REQUEST")
			return false
		case 0x04:
			log.Println("[", session.id, "] ACK")
			return ack(pc, binary.BigEndian.Uint16(message[2:]), session)
		default:
			log.Println("Unknown OP Code:", opCode, "\nBody:", message)
		}
	}
	return false
}

func (session *ReadRequest) prepareFile(pc *net.UDPConn) {
	// log.Println("Session ID: ", session.id)
	if file, err := os.Stat(session.FileName); err == nil {
		session.FileSize = file.Size()
		if file.Size()%512 == 0 {
			session.TotalChunks = (file.Size() / 512)
		} else {
			session.TotalChunks = (file.Size() / 512) + 1
		}
		log.Println("File Exists", "\nSize: ", session.FileSize, "\nTotalChunks: ", session.TotalChunks)
		file, err := os.Open(session.FileName)
		if err != nil {
			sendError(1, "File not found.", pc, session)
		} else {
			session.CurrentChunk = make([]byte, 512)
			session.FileReader = file
			sendNextChunk(pc, session)
		}

	} else {
		sendError(1, "File not found.", pc, session)
	}

}

func ack(pc *net.UDPConn, block uint16, session *ReadRequest) bool {
	// log.Println("ACK Block ", uint64(block))
	// log.Println("ReqType ", session.ReqType)
	// log.Println("CurrentChunkOffset ", session.CurrentChunkOffset)
	// log.Println("CurrentChunkOffset ", int64(block) == session.CurrentChunkOffset)
	// log.Println("Total Chunks ", session.TotalChunks)

	if session.CurrentChunkOffset == int64(block) &&
		session.CurrentChunkOffset != session.TotalChunks {
		sendNextChunk(pc, session)
		return false
	} else if int64(block) == (session.CurrentChunkOffset - 1) {
		log.Println("Peer acknowledged old block, resending...")
		sendCurrentChunk(pc, session)
		return false
	} else {
		log.Println("Transfer successful.")
		session.FileReader.Close()
		return true
	}
}

func sendCurrentChunk(socket *net.UDPConn, session *ReadRequest) {
	// log.Println("Session ID: ", session.id)
	log.Println("[RESEND] Current chunk:", session.CurrentChunkOffset)
	response := new(bytes.Buffer)
	binary.Write(response, binary.BigEndian, []byte{0x00, 0x03})
	binary.Write(response, binary.BigEndian, uint16(session.CurrentChunkOffset))
	binary.Write(response, binary.BigEndian, session.CurrentChunk[:session.CurrentChunkSize])
	log.Println("[RESEND] About to send ", len(response.Bytes()), "bytes to peer")
	socket.WriteToUDP(response.Bytes(), session.Peer)
	if session.CurrentChunkSize < 512 {
		log.Println("DONE: 100%")
	} else {
		percentage := (session.CurrentChunkOffset / session.TotalChunks) * 100
		log.Println(percentage, "%")
	}
}

func sendNextChunk(socket *net.UDPConn, session *ReadRequest) {
	// log.Println("Session ID: ", session.id)
	log.Println("[SEND] Current chunk:", session.CurrentChunkOffset)
	session.CurrentChunkOffset++
	// TODO:
	// use binary.Write in order to avoid this madness of
	// creating slices by hand
	n, err := session.FileReader.Read(session.CurrentChunk)
	session.CurrentChunkSize = n
	log.Println("[SEND] READ: ", n, " ERR:", err)
	response := []byte{0x00, 0x03}
	binaryChunk := make([]byte, 2)
	binary.BigEndian.PutUint16(binaryChunk, uint16(session.CurrentChunkOffset))
	response = append(response, binaryChunk...)
	response = append(response, session.CurrentChunk[:n]...)
	log.Println("[SEND] About to send ", len(response), "bytes to peer")
	log.Println("[SEND] Chunk ", session.CurrentChunkOffset, " sent")
	socket.WriteToUDP(response, session.Peer)
	if n < 512 {
		log.Println("[SEND] DONE: 100%")
	} else {
		percentage := (session.CurrentChunkOffset / session.TotalChunks) * 100
		log.Println(percentage, "%")
	}
}

func sendError(errorCode int, description string, socket *net.UDPConn, session *ReadRequest) {
	log.Println("Session ID: ", session.id)
	// TODO:
	// use binary.Write in order to avoid this madness of
	// creating slices by hand
	response := []byte{0x00, 0x05}
	binaryCode := make([]byte, 2)
	binary.BigEndian.PutUint16(binaryCode, uint16(errorCode))
	response = append(response, binaryCode...)
	response = append(response, description...)
	response = append(response, 0x00)
	log.Println("Returning:", response)
	socket.WriteToUDP(response, session.Peer)

}
