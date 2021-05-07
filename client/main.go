/*
This file lets users make requests about different file storage operations including:
- put/save file to the storage
- get file from the storage
- delete file from the storage
- search file
User's inputs should include the following:
- operation (put/ get/ delete/ search)
- host name: port number
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
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Storj>>> ")
		result := scanner.Scan()
		if result == false {
			break
		}
		//Check result to make sure it's a legit request, then parse in relevant information (operation, host name, port number, file name/path 
			// -- probably use enum here for different kind of messages to store relevant information?)
		message := scanner.Text()
		queryList := strings.Split(message, " ")
		if len(queryList) != 3 {
			fmt.Println("Incorrect request format! Example format:")
			fmt.Println("  - put hostname:portnumber filename")
			fmt.Println("  - get hostname:portnumber filename")
			fmt.Println("  - delete hostname:portnumber filename")
			fmt.Println("  - search hostname:portnumber [string]/[empty space]")
		}
		operation := queryList[0]
		if !strings.EqualFold(operation, "put") && !strings.EqualFold(operation, "get") && !strings.EqualFold(operation, "delete") && !strings.EqualFold(operation, "search") {
			fmt.Println("Invalid request! Allowable requests: put, get, delete, search")
		}
		serverInfo := queryList[1]
		if !strings.Contains(serverInfo, ":") {
			fmt.Println("Missing port number")
		}
		fileInfo := queryList[2]

		conn, err := net.Dial("tcp", serverInfo)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}
		defer conn.Close()
	} 
}