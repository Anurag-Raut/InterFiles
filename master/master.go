package master

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"interfiles/global"
	"interfiles/protocol"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/nrednav/cuid2"
)

type MasterService interface {
	Start()
	startAcceptingConn()
	handleConnection(conn net.Conn)
	addClient(reader *bufio.Reader)
	announceFile(reader *bufio.Reader)
}

type Master struct {
	ID          string
	IP          string
	Port        int
	listener    net.Listener
	clientStore map[string]*global.Client
	fileStore   map[string]*global.File
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	writer := bufio.NewWriter(w)
	writer.WriteString("hello")
	writer.Flush()

}

var replicationFactor = 3

func (master *Master) Start() {

	basePort := 8000
	maxRetries := 12
	var listener net.Listener
	var err error

	for i := 0; i < maxRetries; i++ {
		port := basePort + i
		listener, err = net.Listen("tcp4", fmt.Sprintf(":%d", port))
		if err == nil {
			// Successfully bound to a port
			fmt.Printf("Master Server is listening on port %d\n", port)
			break
		}
		// fmt.Printf("Failed to bind to port %d: %s. Trying next port...\n", port, err)
		time.Sleep(time.Second)
	}
	if err != nil {
		fmt.Println("Error staring server", err.Error())
		return
	}

	addr := listener.Addr().(*net.TCPAddr)

	master.listener = listener
	master.Port = addr.Port
	master.IP = "127.0.0.1"
	master.ID = cuid2.Generate()
	master.clientStore = make(map[string]*global.Client)
	master.fileStore = make(map[string]*global.File)

	go master.startAcceptingConn()

}

func (master *Master) startAcceptingConn() {
	for {
		conn, err := master.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go master.handleConnection(conn)
	}
}

func (master *Master) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	var requestType uint8
	err := binary.Read(reader, binary.LittleEndian, &requestType)
	// fmt.Println(requestType, "REQ_TYPE")
	if err != nil {
		// Handle error
		fmt.Println("Error reading request type:", err)
		return
	}

	// reader.Read(requestType)

	switch requestType {
	case global.ADD_CLIENT:
		master.addClient(reader)

	case global.ANNOUNCE:
		// fmt.Println("BROTHER WE GOOD")
		master.announceFile(reader)

	case global.GET_SENDERS_FOR_FILE:
		master.GetSendersForFile(reader, conn)
	case global.ADD_SENDER_TO_FILE_STORE:
		master.addFile(reader)

	}

}

func (master *Master) addFile(reader *bufio.Reader) {
	// fmt.Println("yooooo boyyyy weeee added the file")
	body, err := io.ReadAll(reader)

	if err != nil {
		fmt.Println("error reading body in addFile")
		return
	}
	args := strings.Split(string(body), ":")
	fileId := args[0]
	clientId := args[1]

	if file, exists := master.fileStore[fileId]; !exists {
		// fmt.Println("DAMNN")
		master.fileStore[fileId] = &global.File{
			ID:      clientId,
			Clients: []global.Client{*master.clientStore[clientId]},
		}
	} else {
		file.Clients = append(file.Clients, *master.clientStore[clientId])
	}

}

func (master *Master) addClient(reader *bufio.Reader) {
	// fmt.Println("LESS FUCKING GOOOO")
	body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("error reading body in addClient")
		return
	}
	clientData := string(body)
	args := strings.Split(clientData, ":")
	// fmt.Println(args[2], "PORTTOOOO")
	newClient := global.Client{
		ClientId:  args[0],
		Ip:        args[1],
		Port:      args[2],
		Directory: args[3],
	}

	master.clientStore[newClient.ClientId] = &newClient

}

func (master *Master) announceFile(reader *bufio.Reader) {
	fmt.Println("YOOOO")
	//finding relevant clients
	body, err := io.ReadAll(reader)

	if err != nil {
		fmt.Println("error reading body in addClient")
		return
	}
	//adding to file store
	// fmt.Println(string(body), "LESS GOOO")
	args := strings.Split(string(body), ":")
	senderClientId := args[0]
	fileId := args[1]
	file := global.File{
		ID:      fileId,
		Clients: []global.Client{*master.clientStore[senderClientId]},
	}

	for receiverClientId, client := range master.clientStore {
		if len(file.Clients) >= replicationFactor {
			fmt.Println("LEN FILE CLIENT", len(file.Clients))
			break
		}
		if receiverClientId == senderClientId {
			continue
		}
		fmt.Println("clientttt", *client, client.GetUrl())
		protocol.RequestToPullFile(*client, file)

	}

}

func (master *Master) GetSendersForFile(reader *bufio.Reader, conn net.Conn) {
	// fmt.Println("WE RAE GOING TO SEND SENDERS LIST")
	fileidLen := uint16(0)
	binary.Read(reader, binary.BigEndian, &fileidLen)
	fileIdBuf := make([]byte, fileidLen)
	_, err := reader.Read(fileIdBuf)

	if err != nil {
		fmt.Println("ERROR wjile reading fileid in master", err)
		return
	}

	fileId := string(fileIdBuf)
	var result string
	if file, exists := master.fileStore[fileId]; exists && file != nil {
		noOfSenders := len(file.Clients)

		for index, client := range master.fileStore[fileId].Clients {

			result += client.ClientId + ":" + client.Ip + ":" + client.Port
			if index != noOfSenders-1 {
				result += ":"
			}
		}

	} else {
		fmt.Println("couldnt access the file store")
		return
	}

	conn.Write([]byte(result))
	// fmt.Println("CLIENTS STRING", result)
	conn.Close()
}

func InitalizeMaster() {
	var master MasterService = &Master{}

	master.Start()

	// return master
}

//todo
// announcement polling
//30 mins distribute
