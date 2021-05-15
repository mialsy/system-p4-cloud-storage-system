/*
This file includes utils for the TCP storage system.
Include funtions to send and receiving messages or message and files combo over the connection,
and also error checking function
*/
package utils

import (
	"P4-siri/message"
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

/*
Function to send Message and file over the network
@param msg: the msg to send over the connection
@param conn: the connection
@return: error message
*/
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


/*
Function to get Message and file over the network
@param msg: the path to store the file
@param conn: the connection
@return: error message
*/
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


/*
Function to get Message over the network
@param conn: the connection
@return: feedback information from msg, error if there is any
*/
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


/*
Function to handle error by logging error message
@param err: the error being checked
*/
func Check(err error) {
	if err != nil {
		log.Fatalln(err.Error())
	}
}