package internal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"
)

const (
	Conn_timeout time.Duration = time.Second * 3
	Dead_timeout time.Duration = time.Second * 10
)

var (
	ZabbixHeader []byte = []byte("ZBXD\x01")

	ErrBadHeader error = errors.New("read zabbix header value is not ZBXD\x01 from tcp connection.")
	ErrBadBody   error = errors.New("read zabbix response boody failed.")
)

type Connection struct {
	conn net.Conn

	lock *sync.RWMutex

	Reader *bufio.Reader
	Writer *bufio.Writer
}

func NewConn(destination string) (*Connection, error) {
	conn, err := net.DialTimeout("tcp", destination, Conn_timeout)
	if err != nil {
		return nil, err
	}

	return &Connection{
		conn:   conn,
		lock:   new(sync.RWMutex),
		Reader: bufio.NewReader(conn),
		Writer: bufio.NewWriter(conn),
	}, nil
}

func (conn *Connection) Read(data []byte) error {
	conn.lock.Lock()
	conn.conn.SetDeadline(time.Now().Add(Dead_timeout))
	_, err := conn.Reader.Read(data)
	conn.lock.Unlock()

	return err
}

func (conn *Connection) ReadResp() ([]byte, error) {
	metadata := make([]byte, 13)
	err := conn.Read(metadata)
	if err != nil {
		return []byte{}, err
	}

	if !bytes.Equal(metadata[0:5], ZabbixHeader) {
		return []byte{}, ErrBadHeader
	}

	length := binary.LittleEndian.Uint64(metadata[4:12])

	data := make([]byte, int64(length))
	err = conn.Read(data)
	if err != nil {
		return []byte{}, ErrBadBody
	}

	return data, nil
}

func (conn *Connection) Write(data []byte) error {
	conn.lock.Lock()
	conn.conn.SetReadDeadline(time.Now().Add(Dead_timeout))
	_, err := conn.Writer.Write(data)
	if err != nil {
		conn.lock.Unlock()
		return err
	}

	err = conn.Writer.Flush()

	conn.lock.Unlock()
	return err
}

func (conn *Connection) Close() error {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	return conn.conn.Close()
}
