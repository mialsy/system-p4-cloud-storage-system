/*
This file defines acknowlegdement that would be returned upon request
*/
package acknowledgement

import (
	"encoding/gob"
	"net"
	"log"
)

type Acknowledgement struct {
	IsSuccess    bool
	FeedbackInfo string
}

// Constructor to create struct Message, initializing copy remain to 1
func New(isSuccess bool, feedbackInfo string) *Acknowledgement {
	msg := Acknowledgement{isSuccess, feedbackInfo}
	return &msg
}

/*
Function to send Message over the network
*/
func (a *Acknowledgement) Send(conn net.Conn) error {
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(a)
	return err
}

func (a *Acknowledgement) ReadAccknowledge(conn net.Conn) error{
	decoder := gob.NewDecoder(conn)

	err := decoder.Decode(&a)
	if err != nil {
		log.Fatal("decode error:", err) 
	}
	return err
}
