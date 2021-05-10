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
	"bufio"
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
		//Check result to make sure it's a legit request, then parse in relevant information (operation, file name/path) 
		message := scanner.Text()
		if message == "exit" {
			break
		}
		queryList := strings.Split(message, " ")

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
		fileInfo := queryList[1]

		fmt.Println(fileInfo)

		msgBytes := make([]byte, 128)
		copy(msgBytes, message)
		conn.Write(msgBytes)
	} 
}

func check(err error) {
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
}
