package gcahe

import (
	"encoding/binary"
	"io"
)

type Node struct {
	Type   int
	Length int64
	Data   []byte
}

func (this *Node) Encode(writer io.Writer) error {

	var xError error

	xError = binary.Write(writer, binary.BigEndian, &this.Type)
	if xError != nil {
		return xError
	}

	xError = binary.Write(writer, binary.BigEndian, &this.Length)
	if xError != nil {
		return xError
	}

	if this.Length > 0 {
		xError = binary.Write(writer, binary.BigEndian, &this.Data)
	}

	return xError

}

func (this *Node) Decode(reader io.Reader) error {

	var xError error

	xError = binary.Read(reader, binary.BigEndian, &this.Type)
	if xError != nil {
		return xError
	}

	xError = binary.Read(reader, binary.BigEndian, &this.Length)
	if xError != nil {
		return xError
	}

	if this.Length > 0 {
		this.Data = make([]byte, this.Length)
		xError = binary.Read(reader, binary.BigEndian, &this.Data)
	}

	return xError

}
