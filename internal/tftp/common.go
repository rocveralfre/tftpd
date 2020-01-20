package tftp

type ackPacket struct {
	opCode  uint16
	blockNo uint16
}

type errPacket struct {
	opCode    uint16
	errorCode uint16
	errMsg    string
}

func getFileName(binary []byte) (string, []byte) {
	var count int

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
