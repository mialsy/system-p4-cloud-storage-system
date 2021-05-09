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
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	argv := os.Args
	if len(argv) != 2 {
		fmt.Println("Incorrect number of arguments!")
		fmt.Println("Usage: ./server hostname:portnumber")
		return
	}

	listener, err := net.Listen("tcp", argv[1])
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	defaultServer := "192.168.122.201:9998"
	if strings.EqualFold(defaultServer, argv[1]) {
		defaultServer := "192.168.122.203:9998"
	}

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
		if err != nil {
			log.Println(err.Error())
			break
		}
		queryList := strings.Split(string(message), " ")
		operation := queryList[0]
		fileInfo := queryList[1]
	}
}