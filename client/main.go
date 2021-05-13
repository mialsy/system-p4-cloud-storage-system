/*
This file lets users make requests about different file storage operations including:
- put/save file to the storage
- get file from the storage
- delete file from the storage
- search file
User's inputs should include the following:
- operation (put/ get/ delete/ search)
- file name (optional for search, path to file for put)
*/

package main

import (
	"P4-siri/message"
	"bufio"
	"fmt"
	"io"
	"encoding/gob"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	argv := os.Args
	if len(argv) != 2 {
		fmt.Println("Incorrect number of arguments!")
		fmt.Println("Usage: ./client hostname:portnumber")
		return
	}

	conn, err := net.Dial("tcp", argv[1])
	check(err)
	defer conn.Close()

	for {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Storj>>> ")
		result := scanner.Scan()
		if result == false {
			break
		}
		// Check result to make sure it's a legit request, then parse in relevant information (operation, file name/path) 
		request := scanner.Text()
		if request == "exit" {
			break
		}
		queryList := strings.Split(request, " ")

		if len(queryList) != 2 {
			fmt.Println("Incorrect request format! Example format:")
			fmt.Println("  - put filename")
			fmt.Println("  - get filename")
			fmt.Println("  - delete filename")
			fmt.Println("  - search [string]/[empty space]")
			continue
		}

		operation := queryList[0]
		if !strings.EqualFold(operation, "put") && !strings.EqualFold(operation, "get") && !strings.EqualFold(operation, "delete") && !strings.EqualFold(operation, "search") {
			fmt.Println("Invalid request! Allowable requests: put, get, delete, search")
			continue
		}
		fileName := queryList[1]

		msg := message.New(operation, fileName)

		// Put operation: open file and find file size information, update size to message header and send message header to server, then send file
		if strings.EqualFold(msg.Operation, "put") {
			file, err := os.OpenFile(msg.FileName, os.O_RDONLY, 0666)
			check(err)
			defer file.Close()

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
		} else {
			msg.Send(conn)

			if strings.EqualFold(msg.Operation, "get") {
				if handleGet(conn) {
					fmt.Println("get success")
				}
			}
		}
	} 
}

func handleGet(conn net.Conn) bool{
	buffer := bufio.NewReader(conn)
	decoder := gob.NewDecoder(buffer)
	var msg message.Message
	err := decoder.Decode(&msg)
	if err == nil {
		// able to get
		fileName := msg.FileName
		if msg.FileSize != 0 {
			file, err := os.OpenFile(fileName, os.O_CREATE | os.O_TRUNC | os.O_RDWR, 0666)
			check(err)

			sz, err := io.CopyN(file, buffer, msg.FileSize)

			if err != nil || sz != msg.FileSize {
				fmt.Printf("copy error, size copied\n", sz)
			}
			file.Close()
		} else {
			fmt.Println(msg.FileName)
			return false
		}
		return true
	} 
	fmt.Println("cannot decode: " + err.Error())
	return false
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
