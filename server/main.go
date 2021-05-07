/*
This file handles clients' requests including:
- put file to the storage
- get file from the storage
- delete file from the storage
- search for string
Notes:
- the system would support concurrent requests and multiple clients (// go routines)
- replicate files (if replica goes down, reject put operation)
- if client put file that already exists in the system, reject operation or overwrite the file
- operation success/failure acknowledgement after replicating
- detect and repair file corruption (// checksum)
*/

package main

import "net"

func main() {
	// listen to request and decide which request call to make
}

func handleConnection(conn net.Conn) {
	
}