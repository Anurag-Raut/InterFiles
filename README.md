# InterFiles
Welcome to the InterFiles project! This project implements a peer-to-peer file sharing system with a master server for coordination and client nodes that communicate directly with each other. The system is built using Go and raw TCP for communication.

## Table of Contents\
- [Introduction](#introduction)  
- [Features](#features)  
- [Architecture](#architecture)  
- [Installation](#installation)  
- [Usage](#usage)  
- [Contact](#contact)  
## Introduction
This project aims to provide a distributed file system similar to BitTorrent. It features a master server that handles peer discovery and coordination, while client nodes communicate in a peer-to-peer manner to share files.

## Features
**Master Server**: Manages peer discovery and coordination.  
**Client Nodes**: Communicate directly with each other to share files.  
**Peer-to-Peer Communication**: Efficient file sharing using raw TCP.  
**Custom Protocol**: Developed in Go for robust client-master communication.  
## Architecture
Master Server
The master server is responsible for:

Maintaining a list of active clients.
Facilitating peer discovery.
Coordinating file distribution.
Client Nodes
Client nodes:

Register with the master server.
Discover peers through the master server.
Establish direct connections with other peers to exchange files.
## Installation
### Prerequisites
Go (version 1.16 or higher)  
Git  
### Clone the Repository
**bash**  
```
git clone https://github.com/yourusername/distributed-file-system.git
 && 
cd distributed-file-system
```

**bash**  
```
go mod download
```
## Usage
**Start Master Server**  

- Open a new terminal
```
go run main.go
```
- Start Master
```
master
```


**Start Multiple Client Node**  
- Open a new terminal
```
go run main.go
```
- Start Client
```
client
```
 
**Upload a file** 
```
upload -p /path/to/file
```
This would create a tracker file

 
**Download a file** 
```
download -p /path/to/tracker-file
```

**Stats**
```
stat -p /path/to/tracker-file
```



## Contact
For any questions or feedback, please contact anurag.raut.86@gmail.com.