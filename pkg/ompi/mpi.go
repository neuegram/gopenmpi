package ompi

import (
	"os"
	"bufio"
	"errors"
	"fmt"
	"strings"
)

var MPI_COMM_WORLD = []byte{0x41, 0x01, 0x00, 0x00, 0x39, 0x00, 0x00, 0x00}
var MPI_CHAR = []byte{0x01, 0x00, 0x00, 0x0}

type MPI interface {
	Initialize() error
	Send(*[]byte, int, int, int) error
	Receive(*[]byte, int, int, int) error
	Finalize() error
}

// Implements MPI
type node struct {
	hosts []string  // Assume sequential hosts (index == rank)
	rank int		// How will we determine own rank?
	comm Comm
}

// Construct and initialize a new node
func NewMPI(hostfile string) (MPI, error) {
	var mpi MPI
	hosts, err := parseHostFile(hostfile); if err != nil {
		return mpi, err
	}

	var hostname string
	if hostname, err = os.Hostname(); err != nil {
		return mpi, err
	}

	rank := -1
	for i, h := range hosts {
		if strings.Compare(hostname, h) == 0 {
			rank = i
			break
		}
		fmt.Println(h)
	}

	comm, err := NewComm()
	if err != nil {
		return mpi, err
	}

	return node{hosts:hosts, rank:rank, comm:comm}, nil
}

// Parse new-line delimited hostlist
func parseHostFile(hostfile string) ([]string, error) {
	var hostlist []string
	var file *os.File
	file, err := os.Open(hostfile); if err != nil {
		return hostlist, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hostlist = append(hostlist, scanner.Text())
	}

	return hostlist, scanner.Err()
}

func (n node) Initialize() error {
	for _, host := range n.hosts {
		if _, err := n.comm.Connect(host); err != nil {
			return err
		}
	}

	return nil
}

func (n node) Rank() int {
	return n.rank
}

// Send the tagged contents of buf to the node with rank dest
func (n node) Send(buf *[]byte, count, dest, tag int) error {
	if len(n.hosts) < dest {
		return errors.New("Destination index out of bounds")
	}

	codable := NewCodable(n.rank, dest, MPI_COMM_WORLD, 0x0, buf, MPI_CHAR, 0x41)
	encoded, err := codable.Encode()
	if err != nil {
		return err
	}

	_, err = n.comm.Write(n.hosts[dest], encoded)
	return err
}

// Receive data with matching tag from src into buf 
func (n node) Receive(buf *[]byte, count, src, tag int) error {
	if len(n.hosts) < src {
		return errors.New("Source index out of bounds")
	}

	encoded := make([]byte, count)
	_, err := n.comm.Read(n.hosts[src], &encoded)
	if err != nil {
		return err
	}

	var codable Codable = Message{}
	decoded, err := codable.Decode(&encoded)
	if err != nil {
		return err
	}

	*buf = decoded

	return nil
}

func (n node) Finalize() error {
	return n.comm.Finalize()
}
