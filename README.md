# **Project 4: Cloud Storage**

Authors: Anh Nguyen, Chuxi Wang

## **Project Description**

In this project, we are building a cloud storage system, `Storj`, similar to Dropbox or Google Drive. Users can save their files to our remote servers and retrieve as well as delete files. We maintain users' files on 2 servers to ensure against file corruption as well as system degrade. IP addresses of our 2 servers are:

```
192.168.122.201:9998
192.168.122.203:9998
```

## **Program Overview**

The program has 2 components:

- Client: sends requests to any of the 2 supported storage servers by supplying the storage server's hostname and port number as command line options at the start, whichever connection works

```
./client 192.168.122.201:9998   or
./client 192.168.122.203:9998
```

- Storage server: 2 are supported to handle requests from multiple clients, replicate files to the other server, detect and handle file corruption. To start the servers, run both of the following commands:

```
./server 192.168.122.201:9998 192.168.122.203:9998
./server 192.168.122.203:9998 192.168.122.201:9998
```

The program supports the following requests:

- Storage: clients can put any type of files to `Storj`
- Retrieval: clients can get files from `Storj` as well as search and list the files in the system

```
TODO: Insert flowchart (Chuxi)
```

## **Program Output** (TODO)

- Put Operation: client sends a put request to save a file to the storage. If the file is saved on both of the servers, the client will get a message indicating so. If the file already exists, the client will be notified and asked to remove the file if he/she wants to proceed with the operation. If the backup server is down, the client's request will be rejected.

```

```

- Get Operation: client sends a get request to retrieve a file from the storage. If the file exists, the client will receive the file in their current working directory.

```

```

- Delete Operation: client sends a request to delete a file from the storage. If the file is deleted from both of the servers, the client will get a confirmation. If the backup server is down, client's request will be rejected.

```

```

- Search Operation: client can search to see the list of files in the storage or search by string to see the list of files whose names match the string.

```

```

## **Included Files**

Following is the list of files included:

- **client/main.go**: processes users' requests to connect with a certain server, perform file storage operations (put, get, delete, search), and receive acknowledgement from the connected server
- **server/main.go**: performs users' request on the server and the replica server
- **message/message.go**: includes definition of a struct to store file information and functions to send and recieve the struct over the network among clients and the 2 supported servers
