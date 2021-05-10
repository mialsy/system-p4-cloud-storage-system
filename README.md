# **Project 4: Cloud Storage**

Author: Anh Nguyen, Chuxi Wang

## **Project Description**

In this project, we are building a cloud storage system, `Storj`, similar to Dropbox or Google Drive.

### **Program Overview**

The program has 2 components:

- Client: sends requests to any of the 2 supported storage servers by supplying the storage server's hostname and port as command line options
- Storage server: 2 are supported to handle requests from multiple clients, replicate files to the other server, detect and handle file corruption

The program supports the following requests:

- Storage: clients can put any type of files to `Storj`
- Retrieval: clients can get files from `Storj` as well as search and list the files in the system

### **Program Output**

- Put Operation

- Get Operation

- Delete Operation

- Search Operation

### **Included Files**

Following is the list of files included:

- **client/main.go**:
- **server/main.go**:
- **Makefile**:
