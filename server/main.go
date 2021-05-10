/*
This file handles clients' requests including:
- put file to the storage
- get file from the storage
- delete file from the storage
- search for string
Notes:
- the system would support concurrent requests and multiple clients
- replicate files (if replica goes down, reject put operation, but other options are ok)
- if client put file that already exists in the system, reject operation or overwrite the file
- operation success/failure acknowledgement after replicating
- detect and repair file corruption
*/

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
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
	check(err)

	fmt.Println(backupServer)

	for {
		if conn, err := listener.Accept(); err == nil {
			go handleConnection(conn)
		}
	}
}

/*
Function to handle different operation requests from client: put, get, delete, and search. 
Also establish connection with backup server to replicate the same operation on the backup server 
and detect and handle file corruption 
*/
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

		fileHash := make(map[string]string) // Map to store a file checksum 

		// Put operation: store received files in "storj" folder, if file already exist then reject the request 
		if strings.EqualFold(operation, "put") {
			fileSize := queryList[2]
			filePath := strings.Split(fileInfo, "/")
			if _, err := os.Stat("./storj"); os.IsNotExist(err) {
    			os.Mkdir("./storj", 0755)
			}
			fileName := "./storj/" + filePath[len(filePath) - 1]	

			if _, err := os.Stat(fileName); err == nil {
				fmt.Println("The file already exists. Please delete the file if you still want to proceed with the operation.")
			} else {
				// Store the file
				file, err := os.OpenFile(fileName, os.O_CREATE | os.O_TRUNC | os.O_RDWR, 0666)
				check(err)
				defer file.Close()
				n, err := strconv.ParseInt(fileSize, 10, 64)
				check(err)
				if _, err := io.CopyN(file, conn, n); err != nil {
					log.Fatalln(err.Error())
					return
				}
				
				// Hash the file 
				fileCheck, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
				check(err)
				defer fileCheck.Close()
				hasher := sha256.New()
				if _, err := io.Copy(hasher, fileCheck); err != nil {
					log.Fatalln(err.Error())
					return
				}
				value := hex.EncodeToString(hasher.Sum(nil))
				fileHash[fileName] = value
				fmt.Println(fileHash)
			}
		}
	}
}

/*
Function to handle error by logging error message
*/
func check(err error) {
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
}

func check(err error) {
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
}
