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
	"strings"
)

func SendMsgAndFile(msg *message.Message, conn net.Conn) error {
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

func GetMsgAndFile(path string, conn net.Conn) error {
	buffer := bufio.NewReader(conn)
	decoder := gob.NewDecoder(buffer)
	var msg message.Message
	err := decoder.Decode(&msg)

	if err != nil {
		return errors.New("cannot decode\n" + err.Error())
	}

	// able to get
	fnameArr := strings.Split(msg.FileName, "/")
	fileName := path + "/" + fnameArr[len(fnameArr) - 1]
	if msg.FileSize >= 0 {
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)

		defer file.Close()
		if err != nil {
			return errors.New("open file error\n" + err.Error())
		}

		sz, err := io.CopyN(file, buffer, msg.FileSize)

		if err != nil {
			return errors.New("send file error\n" + err.Error())
		}
		if sz != msg.FileSize {
			errorInfo := fmt.Sprintf("send file error\n %d actual sent while expected %d\n", sz, msg.FileSize)
			return errors.New(errorInfo)
		}
	} else {
		// error on server side
		return errors.New(msg.FileName)
	}
	return nil
}


func GetMsg(conn net.Conn) (string, error) {
	buffer := bufio.NewReader(conn)
	decoder := gob.NewDecoder(buffer)
	var msg message.Message
	err := decoder.Decode(&msg)

	if err != nil {
		return "", errors.New("cannot decode\n" + err.Error())
	} else if msg.FileSize < 0 {
		// other side failure
		return "", errors.New(msg.FileName)
	} else {
		// success
		return msg.FileName, nil
	}
}