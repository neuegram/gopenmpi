package ompi

import (
	"bufio"
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
	comm []byte	// Process group to which src and dest both belong
	tag  int	// Message classifier
	// Message Body
	buf *[]byte		// Message data
	datatype []byte // Type of message data
	count int		// Number of items in buf of type datatype
}

// Construct and initialize a new concrete type that implements Codable
// The concrete type constructedby NewCodable() will be dependent upon other factors as
// more codables are implemented. For now, strictly Message is used
func NewCodable(src, dest int, comm []byte, tag int, buf *[]byte, datatype []byte, count int) Codable {
	return Message{src, dest, comm, tag, buf, datatype, count}
}

func (m Message) Encode() ([]byte, error) {
	var errorList []error
	var buf bytes.Buffer

	// Dirty closure hack
	writeUint32 := func(val interface{}, maxVarintLen int) {
		var uvarint uint64
		if maxVarintLen == binary.MaxVarintLen32 {
			varint32, ok := val.(int)
			if !ok {
				return
			}
			uvarint = uint64(varint32)
		}

		uvarintBytes := make([]byte, maxVarintLen)
		binary.PutUvarint(uvarintBytes, uvarint)

		_, err := buf.Write(uvarintBytes)
		if err != nil {
			errorList = append(errorList, err)
		}
	}

	if _, err := buf.Write(m.comm); err != nil {
		return buf.Bytes(), err
	}

	length := binary.MaxVarintLen32
	writeUint32(m.count, length)
	writeUint32(m.src, length)
	writeUint32(m.tag, length)

	if len(errorList) > 0 {
		return buf.Bytes(), errorList[0]
	}

	if _, err := buf.Write(m.datatype); err != nil {
		return buf.Bytes(), err
	}

	if _, err := buf.Write(*m.buf); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func (m Message) Decode(buf *[]byte) ([]byte, error) {
	var body []byte

	br := bytes.NewReader(*buf)
	reader := bufio.NewReader(br)

	comm := make([]byte, len(MPI_COMM_WORLD))
	if _, err := reader.Read(comm); err != nil {
		return body, err
	}

	var errorList []error
	readUint32 := func() int {
		uvarint := make([]byte, 4)
		if _, err := reader.Read(uvarint); err != nil {
			errorList = append(errorList, err)
		}

		val, _ := binary.Uvarint(uvarint)
		return int(val)
	}

	m.count = readUint32()
	m.src = readUint32()
	m.tag = readUint32()

	datatype := make([]byte, 4)
	if _, err := reader.Read(datatype); err != nil {
		return body, err
	}
	reader.Discard(3)

	body = make([]byte, reader.Buffered())
	if _, err := reader.Read(body); err != nil {
		return body, err
	}

	return body, nil
}
