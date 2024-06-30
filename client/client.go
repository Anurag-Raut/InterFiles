package client

import (
	"bufio"
	"dfs/global"
	"dfs/protocol"
	"dfs/tracker"
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	// "io"
	"net"
	"os"
	"time"

	"github.com/nrednav/cuid2"
)

type ClientService interface {
	Start()
	AddToServer()
	UploadFile(filepath string)
	startAcceptingConn()
	handleConnection(conn net.Conn)
	handleFileChunk(filename string, chunk []byte)
	startReqFileLoop()
	DownloadFile(trackerFilePath string)
}

type Client struct {
	ID          string
	IP          string
	Port        string
	listener    net.Listener
	Directory   string
	ReqFileChan chan string
}

const (
	GET_FILE = iota
)

func (client *Client) Start() {
	basePort := 8080
	maxRetries := 12
	var listener net.Listener
	var err error
	var i int = 0
	for i = 0; i < maxRetries; i++ {
		port := basePort + i
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			// Successfully bound to a port
			fmt.Printf("Server is listening on port %d\n", port)
			break
		}
		fmt.Printf("Failed to bind to port %d: %s. Trying next port...\n", port, err)
		time.Sleep(time.Second)
	}
	if err != nil {
		fmt.Println("Error staring server", err.Error())
		return
	}

	addr := listener.Addr().(*net.TCPAddr)

	client.listener = listener
	client.Port = strconv.Itoa(addr.Port)
	client.IP = "127.0.0.1"
	client.ID = cuid2.Generate()
	client.AddToServer()
	client.ReqFileChan = make(chan string)
	if client.Directory == "" {
		client.Directory = "/home/anurag/projects/dfs/client" + fmt.Sprintf("%d", i) + "/"
	}

	err = os.MkdirAll(client.Directory, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	go client.startReqFileLoop()
	go client.startAcceptingConn()

}
func (client *Client) startAcceptingConn() {
	for {
		conn, err := client.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go client.handleConnection(conn)
	}
}

func (client *Client) startReqFileLoop() {
	fmt.Println("STARTING REQ LOOP")
	for msg := range client.ReqFileChan {

		args := strings.Split(string(msg), ":")
		senderIP := args[0]
		senderPort := args[1]
		fileId := args[2]
		fmt.Println("WANT TO GET THIS FILE", fileId)

		senderConn, err := net.Dial("tcp", senderIP+":"+senderPort)
		if err != nil {
			fmt.Println("REQ FILE ERROR", err.Error())
			return
		}
		binary.Write(senderConn, binary.BigEndian, global.PULL_FILE)
		binary.Write(senderConn, binary.BigEndian, uint16(len(fileId)))
		senderConn.Write([]byte(fileId))
		// senderConn.Write([]byte{'\x00'})

		reader := bufio.NewReader(senderConn)

		client.getFile(reader, senderConn)

	}
}

func (client *Client) handleConnection(conn net.Conn) {

	var requestType uint8
	reader := bufio.NewReader(conn)

	binary.Read(reader, binary.LittleEndian, &requestType)
	fmt.Println(requestType, "REQ_t")
	switch requestType {
	case GET_FILE:
		client.getFile(reader, conn)
	case global.PULL_FILE:
		client.pullFile(reader, conn)
	case global.REQUEST_FILE:
		client.RequestFile(reader, conn)
	case global.DOWNLOAD_FILE:
		protocol.SendFile(reader, conn, global.Client{
			ClientId:  client.ID,
			Ip:        client.IP,
			Port:      client.Port,
			Directory: client.Directory,
		})
	}

}

func (client *Client) getFile(reader *bufio.Reader, conn net.Conn) {

	totalLen := 0
	var chunkNo = 0
	var fileIdLen uint16
	err := binary.Read(reader, binary.BigEndian, &fileIdLen)
	if err != nil {
		fmt.Println("Error in reading file name len", err.Error())
		return
	}
	fileIdBuf := make([]byte, fileIdLen)
	_, err = reader.Read(fileIdBuf)
	if err != nil {
		fmt.Println("Error in reading file name", err.Error())

		return
	}

	fileId := string(fileIdBuf)

	for {
		chunk := make([]byte, global.CHUNK_SIZE)
		n, err := reader.Read(chunk)

		if err != nil {

			fmt.Println("ERROR OCCURED WHILE READING BYTES", err.Error())
			if err == io.EOF {
				fmt.Println("WE FUCKING EOF")

			} else {
				fmt.Println("WE FUCKING RETURNING BOYS")
				// return
			}
		}

		fmt.Println("chunkNo", chunkNo, "number of bytes", n)
		chunkNo++
		data := chunk

		// fmt.Println("RECIVED FILENAME",filename,"CHUNK",chunk)

		client.handleFileChunk(fileId, data)
		totalLen += n
		ackBuf := []byte{1}
		conn.Write(ackBuf)
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			fmt.Println("totalLen", totalLen)
			// if err == io.EOF {
			// 	break // End of the message
			// }
			break
		}
		fmt.Println("WE NOT EXITING BOYS IG")

	}
	fmt.Println("WE COMPLETED DOWNLOADING NOW WE UPDATE ON MASTER")
	masterConn, err := net.Dial("tcp", global.MASTER_SERVER_URL)
	if err != nil {
		fmt.Println("ERROR OCCURENT WHIL COMNECTING TO MASTER SERVER")
		return
	}

	binary.Write(masterConn, binary.BigEndian, global.ADD_SENDER_TO_FILE_STORE)
	res := fileId + ":" + client.ID
	fmt.Println(res, "REAASSSS")
	masterConn.Write([]byte(res))
	masterConn.Close()

}

func (client *Client) pullFile(reader *bufio.Reader, conn net.Conn) {
	//sender would send to client
	var fileIdLen uint16
	err := binary.Read(reader, binary.BigEndian, &fileIdLen)
	if err != nil {
		fmt.Println("PULL FILE DAYUM", err.Error())
		return
	}
	fmt.Println(fileIdLen, "LEEENENENENE")
	var fileId string
	fileidBuf := make([]byte, fileIdLen)

	// time.Sleep(time.Second * 5)
	err = binary.Read(reader, binary.BigEndian, fileidBuf)
	if err != nil {
		fmt.Println("Binary Read", err.Error())
		return
	}
	fileId = string(fileidBuf)
	fmt.Println("ZQUU", client.Directory+fileId)

	file, err := os.OpenFile(client.Directory+fileId, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("PULL FILE ERROR", err.Error())
		return
	}
	fmt.Println("SENDING::::")
	protocol.UploadToClient(file, conn)

	defer file.Close()

}

func (clinet *Client) RequestFile(reader *bufio.Reader, conn net.Conn) {
	//

	// bro send file
	body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("REQ FILE ERROR", err.Error())
		return
	}
	go func() {
		fmt.Println("BRO : SEND FILE", string(body))
		clinet.ReqFileChan <- string(body)
		fmt.Println("TASKEDDE")
	}()

	conn.Close()

}

func (client *Client) handleFileChunk(filename string, chunk []byte) {

	file, err := os.OpenFile(client.Directory+filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("failed to create file", err.Error())
		return
	}
	defer file.Close()
	_, err = file.Write(chunk)
	if err != nil {
		fmt.Println("failed to write chunk to file:", err.Error())

		return
	}

}
func (client *Client) AddToServer() {

	fmt.Println("writing", global.MASTER_SERVER_URL)
	content := fmt.Sprintf("%s:%s:%s:%s \n", client.ID, client.IP, client.Port, client.Directory)

	conn, err := net.Dial("tcp", global.MASTER_SERVER_URL)

	if err != nil {
		fmt.Println("ERROR WHILE SENDING A REQUEST", err.Error())
		return
	}

	binary.Write(conn, binary.BigEndian, global.ADD_CLIENT)

	conn.Write([]byte(content))

	err = conn.Close()

	if err != nil {
		fmt.Println("ERROR WHILE CLosing A REQUEST", err.Error())
		return
	}

}
func (client *Client) UploadFile(filePath string) {
	fileId := cuid2.Generate()
	file, err := os.Open(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))
	fileId += ext
	if err != nil {
		fmt.Println("ERROR OPENING FILE", err.Error())

	}
	destPath := client.Directory + fileId

	destFile, err := os.Create(destPath)
	if err != nil {
		fmt.Println("error creating destination file:", err.Error())
		return
	}

	_, err = io.Copy(destFile, file)
	if err != nil {
		fmt.Println("error copying file:", err.Error())
		return
	}

	tracker.CreateTrackerFile(destFile, client.ID, fileId)
	protocol.AnnounceFile(fileId, file.Name(), client.ID)
	defer destFile.Close()
	defer file.Close()
}

func (client *Client) DownloadFile(trackerFilePath string) {

	trackerFile, err := os.Open(trackerFilePath)
	if err != nil {
		fmt.Println("Error opening tracker file")
		return
	}

	scanner := bufio.NewScanner(trackerFile)

	// Read the first line (file ID)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Println("error reading tracker file: %w", err)
			return
		}
		fmt.Println("tracker file is empty")
		return
	}

	fileId := scanner.Text()

	potentialSender, err := protocol.GetSendersFromMaster(fileId)
	if err != nil {
		fmt.Println("ERROR while getting clients data")
		return
	}

	protocol.DownloadFile(global.Client{
		ClientId:  client.ID,
		Ip:        client.IP,
		Port:      client.Port,
		Directory: client.Directory,
	}, fileId, potentialSender, trackerFile)

}

func InitalizeClient(directory string) ClientService {
	var client ClientService = &Client{
		Directory: directory,
	}

	client.Start()

	return client
}
