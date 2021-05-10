/*
This file handles clients' requests including:
- put file to the storage
- get file from the storage
- delete file from the storage
- search for string
Notes:
- the system would support concurrent requests and multiple clients (// go routines)
- replicate files (if replica goes down, reject put operation, but other options are ok)
- if client put file that already exists in the system, reject operation or overwrite the file
- operation success/failure acknowledgement after replicating
- detect and repair file corruption (// checksum)
*/

package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	argv := os.Args
	if len(argv) != 3 {
		fmt.Println("Incorrect number of arguments!")
		fmt.Println("Usage: ./server hostname:portnumber backup_hostname:portnumber")
		return
	}

	defaultServer := argv[1]
	backupServer := argv[2]

	listener, err := net.Listen("tcp", defaultServer)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	fmt.Println(backupServer)

	for {
		if conn, err := listener.Accept(); err == nil {
			go handleConnection(conn)
		}
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		// Receive message from client and parse relevant information 
		message := make([]byte, 128)
		_, err := conn.Read(message)
		check(err)
		message = bytes.Trim(message, "\x00")
		queryList := strings.Split(string(message), " ")
		operation := queryList[0]
		fileInfo := queryList[1]

		fmt.Println(queryList)
		fmt.Println(operation)
		fmt.Println(fileInfo)
	}
}

func check(err error) {
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
}
