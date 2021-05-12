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
	"P4-siri/message"
	"bufio"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const storj = "./storj"
const checkFile = "checkFile.txt"

func main() {
	argv := os.Args
	if len(argv) != 3 {
		fmt.Println("Incorrect number of arguments!")
		fmt.Println("Usage: ./server hostname:portnumber backup_hostname:portnumber")
		return
	}

	fileHash := make(map[string]string) // Map to store a file checksum
	// Read checkFile to a map
	if _, err := os.Stat(checkFile); err == nil {
		file, err := os.OpenFile(checkFile, os.O_RDONLY, 0666)
		check(err)
		fscanner := bufio.NewScanner(file)
		for fscanner.Scan() {
			line := fscanner.Text()
			stringList := strings.Split(line, " ")
			fileHash[stringList[0]] = stringList[1]
		}
	}

	defaultServer := argv[1]
	backupServer := argv[2]

	listener, err := net.Listen("tcp", defaultServer)
	check(err)

	for {
		if conn, err := listener.Accept(); err == nil {
			go handleConnection(conn, backupServer, fileHash)
		}
	}
}

/*
Function to handle different operation requests from client: put, get, delete, and search.
Also establish connection with backup server to replicate the same operation on the backup server
and detect and handle file corruption
@param conn: the connection
@param backupServer: hostname and port number of the backup server
@param fileHash: map to store file name and its checksum value
*/
func handleConnection(conn net.Conn, backupServer string, fileHash map[string]string) {
	fmt.Println("I am handle connection")

	defer conn.Close()

	for {
		// Receive message from client
		decoder := gob.NewDecoder(conn)
		msg := &message.Message{}
		decoder.Decode(msg)

		if strings.EqualFold(msg.Operation, "put") {
			handlePut(msg, conn, backupServer, fileHash)
		}
	}
}

/*
Function to handle error by logging error message
@param err: the error being checked
*/
func check(err error) {
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
}

/*
Function to handle put operation: store received files in "storj" folder, if file already exist then reject the request
@param msg: message sent from client side
@param conn: the connection
@param fileHash: map to store file name and its checksum value
*/
func handlePut(msg *message.Message, conn net.Conn, backupServer string, fileHash map[string]string) bool {

	// Send same message to backup server
	fmt.Println(msg.CopyRemain)
	var bconn net.Conn
	// ensure all data is written out to the socket
	if msg.CopyRemain > 0 {
		bconn = connectBackup(backupServer)
		if bconn == nil {
			fmt.Println("backup off")
			return false
		}	
	}
	
	filePath := strings.Split(msg.FileName, "/")
	if _, err := os.Stat(storj); os.IsNotExist(err) {
		os.Mkdir(storj, 0755)
	}
	fileName := storj + "/" + filePath[len(filePath) - 1]

	val, present := fileHash[fileName]
	if present && !strings.EqualFold(val, "deleted") {
		fmt.Println("File already exists. Please delete the file to proceed with the operation.")
		return false
	} 
		// Store the file
		file, err := os.OpenFile(fileName, os.O_CREATE | os.O_TRUNC | os.O_RDWR, 0666)
		check(err)
	 	
		if _, err := io.CopyN(file, conn, msg.FileSize); err != nil {
			fmt.Println(conn)
			fmt.Println(msg.FileName)
			fmt.Println(msg.FileSize)
			log.Fatalln(err)
			return false
		}
		file.Close()

		// Hash the file
		fileCheck, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
		check(err)
		hasher := sha256.New()
		if _, err := io.Copy(hasher, fileCheck); err != nil {
			log.Fatalln(err)
			return false
		}
		value := hex.EncodeToString(hasher.Sum(nil))
		fileCheck.Close()
		fileHash[fileName] = value

		fileStored, err := os.OpenFile(checkFile, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
		check(err)
		line := fileName + " " + value + "\n"
		fileStored.WriteString(line)
		fileStored.Close()
		fmt.Println("File stored in Storj")

		if msg.CopyRemain > 0 {
			defer bconn.Close()
			fmt.Println("send message")
			msgCp := message.New(msg.Operation, msg.FileName)
			msgCp.CopyRemain -= 1

			fileCp, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
			check(err)
			
			statCp, err := fileCp.Stat()
			check(err)
			sizeCp := statCp.Size()
			msgCp.FileSize = sizeCp
			msgCp.Send(bconn)
			
			if _, err := io.Copy(bconn, fileCp); err != nil {
				log.Fatalln(err.Error())
				fileCp.Close()
				return false
			}
			fileCp.Close()
		}		

	
	return true
}

func connectBackup(backupServer string) net.Conn{
	fmt.Println("i am here")

	bconn, err := net.Dial("tcp", backupServer)

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return bconn
}

