/*
This file handles clients' requests including:
- put file to the storage
- get file from the storage
- delete file from the storage
- search for string
Notes:
- the system would support concurrent requests and multiple clients
- replicate files (if replica goes down, reject put operation, but other options are ok)
- if client puts file that already exists in the system, reject operation
- operation success/failure acknowledgement after replicating
- detect and repair file corruption
*/

package main

import (
	"P4-siri/message"
	"P4-siri/utils"
	"bufio"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const storj = "./storj"           // Folder stores files from clients
const checkFile = "checkFile.txt" // File stores files' checksum

func main() {
	argv := os.Args
	if len(argv) != 3 {
		log.Println("Incorrect number of arguments!")
		log.Println("Usage: ./server hostname:portnumber backup_hostname:portnumber")
		return
	}

	// set up logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fileHash := make(map[string]string) // Map to store a file checksum
	// Read checkFile to a map
	if _, err := os.Stat(checkFile); err == nil {
		file, err := os.OpenFile(checkFile, os.O_RDONLY, 0666)
		utils.Check(err)
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
	utils.Check(err)

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
	defer conn.Close()

	for {
		// Receive message from client
		buffer := bufio.NewReader(conn)
		decoder := gob.NewDecoder(buffer)
		var msg message.Message
		decoder.Decode(&msg)

		if strings.EqualFold(msg.Operation, "put") {
			handlePut(msg, conn, buffer, backupServer, fileHash)
		} else if strings.EqualFold(msg.Operation, "get") {
			handleGet(msg, conn, backupServer, fileHash)
		} else if strings.EqualFold(msg.Operation, "search") {
			handleSearch(msg, conn, fileHash)
		} else if strings.EqualFold(msg.Operation, "delete") {
			handleDelete(msg, conn, backupServer, fileHash)
		} 
	}
}

/*
Function to handle put operation: store received files in "storj" folder, if file already exist or backup server is down, then reject the request
@param msg: message sent from client side
@param conn: the connection 
@param buffer: the buffer
@param backupServer: hostname and port number of the backup server
@param fileHash: map to store file name and its checksum value
*/
func handlePut(msg message.Message, conn net.Conn, buffer *bufio.Reader, backupServer string, fileHash map[string]string) {

	if _, err := os.Stat(storj); os.IsNotExist(err) {
		os.Mkdir(storj, 0755)
	}
	filePath := strings.Split(msg.FileName, "/")
	fileName := storj + "/" + filePath[len(filePath)-1]
	val, present := fileHash[fileName]

	if present && !strings.EqualFold(val, "deleted") {
		msg.FileName = "File already exists. Please delete the file to proceed with the operation."
		msg.FileSize = -1
		msg.Send(conn)
		return
	}

	// Store the file
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	utils.Check(err)
	defer file.Close()

	if _, err := io.CopyN(file, buffer, msg.FileSize); err != nil {
		log.Fatalln(err.Error())
		msg.FileSize = -1
		msg.FileName = err.Error()
		msg.Send(conn)
		return
	}

	// Send message to backup server
	if msg.CopyRemain > 0 {
		bconn := connectBackup(backupServer)
		if bconn == nil {
			msg.FileName = "Sorry, can't perform request at this moment"
			os.Remove(fileName)
			msg.Send(conn)
			return
		}
		defer bconn.Close()

		msg.CopyRemain -= 1
		msg.FileName = fileName

		err := utils.SendMsgAndFile(&msg, bconn)
		if err != nil {
			log.Println(err.Error())
			return
		}

		feedBack, err := utils.GetMsg(bconn)
		if err != nil {
			// error saving on backup server
			log.Println(err.Error())
			msg.FileSize = -1
			msg.FileName = err.Error()
			msg.Send(conn)
			return
		} else {
			log.Println(feedBack)
		}
	}

	// Hash the file
	value := hash(fileName)
	fileHash[fileName] = value
	fileStored, err := os.OpenFile(checkFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	utils.Check(err)
	defer fileStored.Close()
	line := fileName + " " + value + "\n"
	fileStored.WriteString(line)
	msg.FileName = "File stored"
	msg.Send(conn)

}

/*
Function to handle get operation request: find the file in "storj" folder and check its checksum. If the file is corrupted, contact backup server
to get the correct file then send the file to client.
@param msg: message sent from client side
@param conn: the connection
@param backupServer: hostname and port number of the backup server
@param fileHash: map to store file name and its checksum value
*/
func handleGet(msg message.Message, conn net.Conn, backupServer string, fileHash map[string]string) {
	filePath := strings.Split(msg.FileName, "/")
	fileName := storj + "/" + filePath[len(filePath)-1]

	expectedVal, isInMap := fileHash[fileName]

	if !isInMap || expectedVal == "deleted" {
		log.Println("no such file")
		// not in the map, not able to do get
		msg.FileName = "no such file"
		msg.FileSize = -1
		msg.Send(conn)
		return
	}

	actualVal := hash(fileName)

	if expectedVal != actualVal && msg.CopyRemain == 0 {
		// file broken
		msg.FileName = "file broken"
		msg.FileSize = -1
		msg.Send(conn)
		return
	} else if expectedVal != actualVal && msg.CopyRemain > 0 {
		// send to the request to the other server
		bconn := connectBackup(backupServer)

		if bconn == nil {
			log.Println("backup off")
			msg.FileName = "original file broken, backup off"
			msg.FileSize = -1
			msg.Send(conn)
			return
		}
		defer bconn.Close()

		// send request to backup server
		log.Println("send message")
		msg.CopyRemain -= 1

		msg.Send(bconn)

		err := utils.GetMsgAndFile(storj, bconn)

		// fail to get from backup as well
		if err != nil {
			log.Println(err.Error())
			msg.FileSize = -1
			msg.FileName = err.Error()
			msg.Send(conn)
			return
		} 
	}

	msg.FileName = fileName
	utils.SendMsgAndFile(&msg, conn)
}

/*
Function to handle search operation request: send client the list of files that contain the string in the search. If the string is empty, then list all of the files.
@param msg: message sent from client side
@param conn: the connection
@param fileHash: map to store file name and its checksum value
*/
func handleSearch(msg message.Message, conn net.Conn, fileHash map[string]string) {
	query := msg.FileName

	queryRes := make([]byte, 0)

	for fileName := range fileHash {
		if fileHash[fileName] == "deleted" {
			continue
		}
		strs := strings.Split(fileName, "/")
		fileName = strs[len(strs) - 1]
		if strings.Index(fileName, query) != -1 {
			if len(queryRes) > 0 {
				queryRes = append(queryRes, ", "...)
			}
			queryRes = append(queryRes, strs[len(strs)-1]...)
		}
	}
	if len(queryRes) > 0 {
		msg.FileName = "Query result: " + string(queryRes)
	} else {
		msg.FileName = "No matching result found."
	}
	msg.Send(conn)
}

/*
Function to handle delete operation request: find the file in "storj" folder and delete files on both servers. If one of the servers is down, then reject the operation.
@param msg: message sent from client side
@param conn: the connection
@param backupServer: hostname and port number of the backup server
@param fileHash: map to store file name and its checksum value
*/
func handleDelete(msg message.Message, conn net.Conn, backupServer string, fileHash map[string]string) {
	fileName := storj + "/" + msg.FileName
	val, present := fileHash[fileName]
	if !present || strings.EqualFold(val, "deleted") {
		msg.FileName = "File doesn't exist"
		msg.Send(conn)
	} else {
		if msg.CopyRemain > 0 {
			bconn := connectBackup(backupServer)
			if bconn == nil {
				msg.FileName = "Sorry, can't perform request at this moment"
				msg.Send(conn)
				return
			}
			defer bconn.Close()
			msg.CopyRemain -= 1
			msg.Send(bconn)
		}

		err := os.Remove(fileName)
		utils.Check(err)
		fileHash[fileName] = "deleted"
		file, err := os.OpenFile(checkFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		utils.Check(err)
		defer file.Close()
		line := fileName + " " + "deleted" + "\n"
		file.WriteString(line)
		msg.FileName = "File removed"
		msg.Send(conn)
	}
}

/*
Function to connect with backup server
@param backupServer: hostname and port number of the backup server
@return the connection
*/
func connectBackup(backupServer string) net.Conn {
	bconn, err := net.Dial("tcp", backupServer)

	if err != nil {
		log.Println(err.Error())
		return nil
	}
	return bconn
}

/*
Function to find the checksum of a file
@param fileName: name of the file
@return the checksum value of the file
*/
func hash(fileName string) string {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	utils.Check(err)
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Fatalln(err.Error())
		return ""
	}
	value := hex.EncodeToString(hasher.Sum(nil))
	return value
}
