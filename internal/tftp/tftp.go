package tftp

//TftpSession :
//Due to the stateless nature of UDP
//We have to create a State using this struct
type TftpSession struct {
	fileName     string
	currentChunk int
	fileSize     int
}
