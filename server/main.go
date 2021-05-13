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
		buffer := bufio.NewReader(conn)
		decoder := gob.NewDecoder(buffer)
		// decoder := gob.NewDecoder(conn)
		var msg message.Message
		err := decoder.Decode(&msg)
		if err == nil {
			log.Println(err.Error())
		}

		if strings.EqualFold(msg.Operation, "put") {
			handlePut(msg, buffer, backupServer, fileHash)
		} else if strings.EqualFold(msg.Operation, "get") {
			handleGet(msg, conn, backupServer, fileHash)	
		} else if strings.EqualFold(msg.Operation, "search") {
			handleSearch(msg, conn, fileHash)
		} else if strings.EqualFold(msg.Operation, "delete") {
			handleDelete(msg, backupServer, fileHash)
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
func handlePut(msg message.Message, buffer *bufio.Reader, backupServer string, fileHash map[string]string) bool {

	// defer conn.Close()

	fmt.Println("handling put")
	filePath := strings.Split(msg.FileName, "/")
	if _, err := os.Stat(storj); os.IsNotExist(err) {
		os.Mkdir(storj, 0755)
	}
	fileName := storj + "/" + filePath[len(filePath) - 1]

	fmt.Println(fileName)
	val, present := fileHash[fileName]
	fmt.Println(fileHash)
	if present && !strings.EqualFold(val, "deleted") {
		fmt.Println("File already exists. Please delete the file to proceed with the operation.")
	} else {
		// Store the file
		fmt.Println("storing file")
		file, err := os.OpenFile(fileName, os.O_CREATE | os.O_TRUNC | os.O_RDWR, 0666)
		check(err)
		defer file.Close()

		if _, err := io.CopyN(file, buffer, msg.FileSize); err != nil {
			log.Fatalln(err.Error())
			return false
		}

		// Send same message to backup server
		fmt.Println(msg.CopyRemain)
		if msg.CopyRemain > 0 {
			bconn := connectBackup(backupServer)
			if bconn == nil {
				fmt.Println("backup off")
				return false
			}
			defer bconn.Close()
			fmt.Println("send message")
			msg.CopyRemain -= 1
			// msg.Send(bconn)

			fileCopy, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
			check(err)
			defer fileCopy.Close()

			bbuffer := bufio.NewWriter(bconn)
			encoder := gob.NewEncoder(bbuffer)
			fmt.Println(msg)
			encoder.Encode(msg)
			sz, err := io.Copy(bbuffer, fileCopy)
			fmt.Println(sz)
			if err != nil {
				fmt.Println(err.Error())
			}
			bbuffer.Flush()

			// if _, err := io.Copy(bconn, fileCopy); err != nil {
			// 	log.Fatalln(err.Error())
			// 	return false
			// }
			// time.Sleep(8 * time.Second)

		}

		// Hash the file
		fileCheck, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
		check(err)
		defer fileCheck.Close()
		hasher := sha256.New()
		if _, err := io.Copy(hasher, fileCheck); err != nil {
			log.Fatalln(err.Error())
			return false
		}
		value := hex.EncodeToString(hasher.Sum(nil))
		fileHash[fileName] = value
		fileStored, err := os.OpenFile(checkFile, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
		check(err)
		defer fileStored.Close()
		line := fileName + " " + value + "\n"
		fileStored.WriteString(line)
		fmt.Println("File stored in Storj")
	}

	return true
}

func handleGet(msg message.Message, conn net.Conn, backupServer string, fileHash map[string]string)  {
	filePath := strings.Split(msg.FileName, "/")
	fileName := storj + "/" + filePath[len(filePath) - 1]
	fmt.Println(msg.FileName)

	expectedVal, isInMap := fileHash[fileName]
	log.Println(fileHash)

	if !isInMap {
		log.Println("no such file")
		// not in the map, not able to do get
		msg.FileName = "no such file"
		msg.Send(conn)
		return
	}

	fileCheck, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	check(err)

	hasher := sha256.New()
	if _, err := io.Copy(hasher, fileCheck); err != nil {
		msg.FileName = "hash error"
		msg.Send(conn)
		return
	}
	fileCheck.Close()

	actualVal := hex.EncodeToString(hasher.Sum(nil))

	if expectedVal != actualVal && msg.CopyRemain == 0 {
		// file broken
		msg.FileName = "file broken"
		msg.Send(conn)
		return
	} else if expectedVal != actualVal && msg.CopyRemain > 0 {
		// send to the request to the other server
		bconn := connectBackup(backupServer)

		if bconn == nil {
			fmt.Println("backup off")
			msg.FileName = "original file broken, backup off"
			msg.Send(conn)
			return 
		}
		defer bconn.Close()

		// send request to backup server
		fmt.Println("send message")
		msg.CopyRemain -= 1

		msg.Send(bconn)

		readerBuffer := bufio.NewReader(bconn)
		bdecoder := gob.NewDecoder(readerBuffer)
		var msgBackup message.Message
		err := bdecoder.Decode(&msgBackup)

		if err == nil {
			// able to get
			if msgBackup.FileSize != 0 {
				fmt.Println("get from back up " + fileName)
				file, err := os.OpenFile(fileName, os.O_CREATE | os.O_TRUNC | os.O_RDWR, 0666)
				check(err)
				file.Truncate(0)

				sz, err := io.CopyN(file, readerBuffer, msgBackup.FileSize)

				if err != nil || sz != msgBackup.FileSize {
					log.Printf("copy error, size copied%d\n", sz)
				}
				file.Close()
				file1, _ := os.OpenFile(fileName, os.O_RDONLY, 0666)

				// if reaches here, should have the correct copy
				stat1, _ := file1.Stat()
				fmt.Println(stat1.Size())
			} else {
				// error copying
				fmt.Println(msgBackup.FileName)
				msgBackup.Send(conn)
				return
			}
		}


	} 
	
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)

	// if reaches here, should have the correct copy
	stat, err := file.Stat()
	check(err)
	size := stat.Size()
	msg.FileSize = size
	buffer := bufio.NewWriter(conn)
	encoder := gob.NewEncoder(buffer)
	encoder.Encode(msg)
	sz, err := io.Copy(buffer, file)
	fmt.Println(sz)
	if err != nil {
		fmt.Println(err.Error())
	}
	buffer.Flush()
	file.Close()
}

func handleSearch(msg message.Message, conn net.Conn, fileHash map[string]string) {
	query := msg.FileName

	queryRes := make([]byte, 0)

	for fileName := range fileHash {		
		if strings.Index(fileName, query) != -1 {
			if len(queryRes) > 0 {
				queryRes = append(queryRes, ", "...)
			}
			strs := strings.Split(fileName, "/")
			queryRes = append(queryRes, strs[len(strs) - 1]...)
		}
	} 
	msg.FileName = string(queryRes)
	msg.Send(conn)
}

func handleDelete(msg message.Message, backupServer string, fileHash map[string]string) {
	fileName := storj + "/" + msg.FileName
	val, present := fileHash[fileName]
	if !present || strings.EqualFold(val, "deleted") {
		fmt.Println("File doesn't exist")
	} else {
		if msg.CopyRemain > 0 {
			bconn := connectBackup(backupServer)
			if bconn == nil {
				fmt.Println("backup off")
				return
			}
			defer bconn.Close()
			msg.CopyRemain -= 1
			msg.Send(bconn)
		}

		err := os.Remove(fileName)
		check(err)
		fileHash[fileName] = "deleted"
		file, err := os.OpenFile(checkFile, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
		check(err)
		defer file.Close()
		line := fileName + " " + "deleted" + "\n"
		file.WriteString(line)
		fmt.Print("File removed")
	}
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
