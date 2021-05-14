/*
This file includes utils for the TCP storage system.
*/
package utils

import (
	"P4-siri/message"
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func SendMsgAndFile(msg *message.Message, conn net.Conn) error{
	file, err := os.OpenFile(msg.FileName, os.O_RDONLY, 0666)
	if err != nil {
		return errors.New("open file error\n" + err.Error())
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return errors.New("stat error\n" + err.Error())
	}

	size := stat.Size()
	msg.FileSize = size
	buffer := bufio.NewWriter(conn)
	encoder := gob.NewEncoder(buffer)

	// send message
	err = encoder.Encode(msg)
	if err != nil {
		return errors.New("send message error\n" + err.Error())
	}

	// send file
	sz, err := io.Copy(buffer, file)
	if err != nil {
		return errors.New("send file error\n" + err.Error())
	}
	if sz != msg.FileSize {
		errorInfo := fmt.Sprintf("send file error\n %d actual sent while expected %d\n", sz, msg.FileSize)
		return errors.New(errorInfo)
	}
	buffer.Flush()
	return nil
}
