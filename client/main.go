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
			msg.Send(conn)
			
			if _, err := io.Copy(conn, file); err != nil {
				log.Fatalln(err.Error())
				return
			}
		} else {
			msg.Send(conn)
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
