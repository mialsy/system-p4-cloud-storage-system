/*
This file defines Message struct which includes information about requested operation,
file name, file size, and number of copy to make
*/
package message

import (
	"encoding/gob"
	"net"
)

type Message struct {
	Operation string
	FileName string
	FileSize int64
	CopyRemain int
}

// Constructor to create struct Message, initializing copy remain to 1
func New(operation string, fileName string) *Message {
	msg := Message{operation, fileName, 0, 1}
	return &msg
}

/*
Function to send Message over the network
*/
func (m *Message) Send(conn net.Conn) error {
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(m)
	return err
}