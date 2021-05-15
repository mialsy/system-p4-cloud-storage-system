# **Project 4: Cloud Storage**

Authors: Anh Nguyen, Chuxi Wang

## **Project Description**

In this project, we are building a cloud storage system, `Storj`, similar to Dropbox or Google Drive. Users can save their files to our remote servers and retrieve as well as delete files. We maintain users' files on 2 servers to ensure against file corruption as well as system degrade. IP addresses of our 2 servers are:

```
192.168.122.201:9998
192.168.122.203:9998
```

## **Program Overview**
### Component of the program
The program has 2 components:

- Storage server: 2 are supported to handle requests from multiple clients, replicate files to the other server, detect and handle file corruption. To start the servers, run both of the following commands:

```console
./server <server address 1>:<port1> <server address 2>:<port2>
./server <server address 2>:<port2> <server address 1>:<port1>
```

- Client: sends requests to any of the 2 supported storage servers by supplying the storage server's hostname and port number as command line options at the start, whichever connection works, to start client, run the following command:

```console
./client <server address>:<port>
```

### Server-Client Model
As servers mirror each other, we design the server as each other's client. The relationship between server and client is shown as below.

<img src="https://github.com/usf-cs521-sp21/P4-siri/blob/main/img/cllient_server_model.gif" alt="server client model">

To ensure the storage safe, for put and delete request we always put and delete on both servers during one request. Once a server is down, these two types of request will be rejected. If one server fails to put/delete a file, the attempt would be considered failure. 

<img src="https://github.com/usf-cs521-sp21/P4-siri/blob/main/img/cllient_server_model2.gif" alt="put and delete workflow">

Different from put and delete requests, search and get can still be performed while the backup server is down. For search request, if the file is corrupted, an attempt recovery would be performed. A get request would be sent to the backup server to retrieve the file. Get request only fails when both copies are broken.

<img src="https://github.com/usf-cs521-sp21/P4-siri/blob/main/img/cllient_server_model3.gif" alt="search and get workflow">

To ensure resiliency of meta data after servers get offline, we keep a checkFile.txt to store file storage info as well as hash value of the file. The file info would be written to checkFile every time a successful put/delete is performed.


## **Program Options**
The program supports the following requests:

- Storage: clients can put any type of files to `Storj`
    - to store file, runs:
        ```console
        Storj>>> put <file full path>
        ```
- Retrieval: clients can get files from `Storj` as well as search and list the files in the system
    - to list all files, runs:
        ```console
        Storj>>> search 
        ```
    - to search file, runs:
        ```console
        Storj>>> search <query>
        ```
    - to get file, runs:
        ```console
        Storj>>> get <file name>
        ```
- Delete: client can delete files from `Storj`
    - to list all files, runs:
        ```console
        Storj>>> delete <file name>
        ```

Note that put and delete will not be supported while one server is down. And while search query allows partial matching, get and delete need to specify the complete file name (including file extension).

## **Program Output** 

- Put Operation: client sends a put request to save a file to the storage. If the file is saved on both of the servers, the client will get a message indicating so. If the file already exists, the client will be notified and asked to remove the file if he/she wants to proceed with the operation. If the backup server is down, the client's request will be rejected.

<img src="https://github.com/usf-cs521-sp21/P4-siri/blob/main/img/put.gif" alt="put output">

- Get Operation: client sends a get request to retrieve a file from the storage. If the file exists, the client will receive the file in their current working directory.

<img src="https://github.com/usf-cs521-sp21/P4-siri/blob/main/img/get.gif" alt="get output">

- Delete Operation: client sends a request to delete a file from the storage. If the file is deleted from both of the servers, the client will get a confirmation. If the backup server is down, client's request will be rejected.

<img src="https://github.com/usf-cs521-sp21/P4-siri/blob/main/img/delete.gif" alt="delete output">

- Search Operation: client can search to see the list of files in the storage or search by string to see the list of files whose names match the string.

<img src="https://github.com/usf-cs521-sp21/P4-siri/blob/main/img/search.gif" alt="search output">

## **Included Files**

Following is the list of files included:

- **client/main.go**: processes users' requests to connect with a certain server, perform file storage operations (put, get, delete, search), and receive acknowledgement from the connected server.
- **server/main.go**: performs users' request on the server and the replica server.
- **message/message.go**: includes definition of a struct to store file information and functions to send the struct over the network among clients and the 2 supported servers.
- **utils/utils.go**: includes utils for the TCP storage system, include funtions to send and receiving messages or message and files combo over the connection,
and also error checking functions.
- **clean.sh**: will remove the storj and checkFile, which reset the server.
