/*
This file lets users make requests about different file storage operations including:
- put/save file to the storage
- get file from the storage
- delete file from the storage
- search file
User's inputs should include the following:
- operation (put/ get/ delete/ search)
- file name (or file path to file for put)
*/

package main

import (
	"P4-siri/message"
	"P4-siri/utils"
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

	// set up logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	conn, err := net.Dial("tcp", argv[1])
	utils.Check(err)
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

		// preprocess exit and search without keyword
		if strings.EqualFold(request, "exit") {
			break
		} else if strings.EqualFold(request, "search") {
			msg := message.New(request, "")
			msg.Send(conn)

			feedBack, err := utils.GetMsg(conn)
			if err != nil {
				log.Println(err.Error())
			} 
			fmt.Println(feedBack)
			continue
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

		if strings.EqualFold(msg.Operation, "put") {
			err := utils.SendMsgAndFile(msg, conn)
			if err != nil {
				// if send file fail, not receiving
				log.Println(err.Error())
				continue
			}
		} else {
			msg.Send(conn)
		}

		if strings.EqualFold(msg.Operation, "get") {
			err := utils.GetMsgAndFile(".", conn)
			if err != nil {
				log.Println(err.Error())
			} else {
				fmt.Println("success")
			}
		} else {
			feedBack, err := utils.GetMsg(conn)
			if err != nil {
				log.Println(err.Error())
			} else {
				fmt.Println(feedBack)
			}
		} 
	}
}
