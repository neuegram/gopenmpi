package ompi

import (
	"net"
)

const PORT = ":1337"

type Comm interface {
	ResolveHost(string) (string, error)
	Connect(string) (net.Conn, error)
	Connected(string) bool
	Read(string, *[]byte) (int, error)
	Write(string, []byte) (int, error)
	Finalize() error
}

// Implements Comm
type TCPComm struct {
	conn map[string]*net.Conn
	listener net.Listener
}

// Construct and initialize a new concrete type that implements Comm
// The concrete type constructed by NewComm() will be dependent upon other factors as
// more protocols are implemented. For now, strictly TCPComm is used
func NewComm() (Comm, error)  {
	var comm TCPComm
	comm.conn = make(map[string]*net.Conn)

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		return comm, err
	}
	comm.listener = listener

	return comm, nil
}

func (tcp TCPComm) ResolveHost(host string) (string, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "[" + host + "]" + PORT)
	if err != nil {
		return "", nil
	}

	return tcpAddr.IP.String(), nil
}

func (tcp TCPComm) Connect(host string) (net.Conn, error) {
	if !tcp.Connected(host) {
		conn, err := net.Dial("tcp", "[" + host + "]" + PORT)
		if err != nil {
			return conn, err
		}

		tcp.conn[host] = &conn
	}

	return *tcp.conn[host], nil
}

func (tcp TCPComm) Connected(host string) bool {
	if _, ok := tcp.conn[host]; ok {
		return true
	}
	return false
}

func (tcp TCPComm) Read(host string, buf *[]byte) (int, error) {
	if !tcp.Connected(host) {
		if _, err := tcp.Connect(host); err != nil {
			return 0, err
		}
	}

	n, err := (*tcp.conn[host]).Read(*buf)
	return n, err
}

func (tcp TCPComm) Write(host string, buf []byte) (int, error) {
	if !tcp.Connected(host) {
		if _, err := tcp.Connect(host); err != nil {
			return 0, err
		}
	}

	n, err := (*tcp.conn[host]).Write(buf)
	return n, err
}

func (tcp TCPComm) Finalize() error {
	for _, v := range tcp.conn {
		if err := (*v).Close(); err != nil {
			return err
		}
	}
	return nil
}

