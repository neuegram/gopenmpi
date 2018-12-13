package ompi

import (
	"bytes"
	"encoding/binary"
)

type Codable interface {
	Encode() ([]byte, error)
	Decode(*[]byte) ([]byte, error)
}

type Message struct {
	// Message envelope
	src int		// Sending process rank
	dest int	// Receiving process rank
	comm uint64	// Process group to which src and dest both belong
	tag  int	// Message classifier
	// Message Body
	buf *[]byte		// Message data
	datatype int // Type of message data
	count int		// Number of items in buf of type datatype
}

// Construct and initialize a new concrete type that implements Codable
// The concrete type constructedby NewCodable() will be dependent upon other factors as
// more codables are implemented. For now, strictly Message is used
func NewCodable(src, dest int, comm uint64, tag int, buf *[]byte, datatype, count int) Codable {
	return Message{src, dest, comm, tag, buf, datatype, count}
}

func (m Message) Encode() ([]byte, error) {
	var errorList []error
	var buf bytes.Buffer

	// Dirty closure hack
	writeUint := func(val interface{}, maxVarintLen int) int {
		uvarint, ok := val.(uint64)
		if !ok {
			return 0
		}

		uvarintBytes := make([]byte, maxVarintLen)
		binary.PutUvarint(uvarintBytes, uvarint)

		n, err := buf.Write(uvarintBytes)
		if err != nil {
			errorList = append(errorList, err)
		}

		return n
	}

	length := binary.MaxVarintLen64
	writeUint(m.comm, length)

	length = binary.MaxVarintLen32
	writeUint(m.count, length)
	writeUint(m.src, length)
	writeUint(m.tag, length)
	writeUint(m.datatype, length)

	if len(errorList) > 0 {
		return buf.Bytes(), errorList[0]
	}

	if _, err := buf.Write(*m.buf); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func (m Message) Decode(buf *[]byte) ([]byte, error) {
	return []byte{}, nil
}
