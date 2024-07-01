package client

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"interfiles/global"
	"interfiles/protocol"
	"interfiles/tracker"
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
	Start() error
	AddToServer() error
	AnnounceFile(filepath string) error
	startAcceptingConn()
	handleConnection(conn net.Conn)
	handleFileChunk(file *os.File, chunk []byte)
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

func (client *Client) Start() error {
	basePort := 8080
	maxRetries := 12
	var listener net.Listener
	var err error
	var i int = 0
	for i = 0; i < maxRetries; i++ {
		port := basePort + i
		listener, err = net.Listen("tcp4", fmt.Sprintf(":%d", port))
		if err == nil {
			// Successfully bound to a port
			break
		}
		fmt.Printf("Failed to bind to port %d: %s. Trying next port...\n", port, err)
		time.Sleep(time.Second)
	}
	if err != nil {

		return fmt.Errorf("Error staring server %v", err.Error())
	}

	addr := listener.Addr().(*net.TCPAddr)

	client.listener = listener
	client.Port = strconv.Itoa(addr.Port)
	client.IP = addr.IP.String()
	client.ID = cuid2.Generate()
	err = client.AddToServer()
	if err != nil {
		return err
	}
	client.ReqFileChan = make(chan string)
	if client.Directory == "" {
		currentDir, err := os.Getwd()
		if err != nil {

			return fmt.Errorf("failed to get current directory: %v", err)
		}

		client.Directory = filepath.Join(currentDir, fmt.Sprintf("client%d", i)) + "/"
		// fmt.Println(client.Directory, "CLINET DIRECTORYYYY")
	}
	// fmt.Println(string(addr.IP), "ADDR")
	err = os.MkdirAll(client.Directory, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %v ", err)

	}

	go client.startReqFileLoop()
	go client.startAcceptingConn()
	fmt.Printf("Server is listening on %s:%s\n", client.IP, client.Port)
	return nil
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
	// fmt.Println("STARTING REQ LOOP")
	for msg := range client.ReqFileChan {

		args := strings.Split(string(msg), ":")
		senderIP := args[0]
		senderPort := args[1]
		fileId := args[2]
		senderConn, err := net.Dial("tcp4", senderIP+":"+senderPort)
		if err != nil {
			// fmt.Println("REQ FILE ERROR", err.Error())
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
	// fmt.Println(requestType, "REQ_t")
	switch requestType {
	case global.GET_FILE:
		client.getFile(reader, conn)
	case global.PULL_FILE:
		client.pullFile(reader, conn)
	case global.REQUEST_TO_PULL_FILE:
		client.RequestToPullFile(reader, conn)
	case global.DOWNLOAD_FILE:
		protocol.SendFile(reader, conn, global.Client{
			ClientId:  client.ID,
			Ip:        client.IP,
			Port:      client.Port,
			Directory: client.Directory,
		})

	case global.SENDER_HANDSHAKE:
		client.senderHandShake(reader, conn)

	case global.RECEIVER_HANDSHAKE:
		client.receiverHandShake(reader, conn)

	}

}

func (client *Client) senderHandShake(reader *bufio.Reader, conn net.Conn) {

	binary.Write(conn, binary.BigEndian, uint8(1))

}
func (client *Client) receiverHandShake(reader *bufio.Reader, conn net.Conn) {

	binary.Write(conn, binary.BigEndian, uint8(1))

}

func (client *Client) getFile(reader *bufio.Reader, conn net.Conn) {
	// fmt.Println("WE SENDIN FILE BOYZZZZ")
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
		return
	}

	fileId := string(fileIdBuf)
	totalBytes := 0
	file, err := os.OpenFile(client.Directory+fileId, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("ERORR", err)
		return
	}
	for {
		chunk := make([]byte, global.CHUNK_SIZE)
		noOfBytes := uint16(0)
		binary.Read(conn, binary.BigEndian, &noOfBytes)

		_, err := reader.Read(chunk)

		if err != nil {

			fmt.Println("ERROR OCCURED WHILE READING BYTES", err.Error())
			if err == io.EOF {
				break
			} else {
				fmt.Println("WE FUCKING RETURNING BOYS")
				// return
			}
		}
		chunk = chunk[:noOfBytes]
		totalBytes += len(chunk)
		// fmt.Println("chunkNo", chunkNo, "number of bytes", n)
		chunkNo++

		client.handleFileChunk(file, chunk)
		// totalLen += n
		ackBuf := []byte{1}
		conn.Write(ackBuf)
		if err == io.EOF {
			// fmt.Println("Error reading from connection:", err)
			// fmt.Println("totalLen", totalLen)
			// if err == io.EOF {
			// 	break // End of the message
			// }
			break
		}
	}
	// fmt.Println("WE COMPLETED DOWNLOADING NOW WE UPDATE ON MASTER")
	masterConn, err := net.Dial("tcp4", global.MASTER_SERVER_URL)
	if err != nil {
		fmt.Println("ERROR OCCURENT WHIL COMNECTING TO MASTER SERVER")
		return
	}

	binary.Write(masterConn, binary.BigEndian, global.ADD_SENDER_TO_FILE_STORE)
	res := fileId + ":" + client.ID
	// fmt.Println(res, "REAASSSS")
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
	// fmt.Println(fileIdLen, "LEEENENENENE")
	var fileId string
	fileidBuf := make([]byte, fileIdLen)

	// time.Sleep(time.Second * 5)
	err = binary.Read(reader, binary.BigEndian, fileidBuf)
	if err != nil {
		fmt.Println("Binary Read", err.Error())
		return
	}
	fileId = string(fileidBuf)
	file, err := os.OpenFile(client.Directory+fileId, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("PULL FILE ERROR", err.Error())
		return
	}
	// fmt.Println("SENDING::::")
	protocol.UploadToClient(file, conn)

	defer file.Close()

}

func (clinet *Client) RequestToPullFile(reader *bufio.Reader, conn net.Conn) {
	//

	// Please send the file
	// fmt.Println("RECEIVED A REQ TO PULL FiLE")
	body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("REQ FILE ERROR", err.Error())
		return
	}
	go func() {
		clinet.ReqFileChan <- string(body)
	}()

	conn.Close()

}

func (client *Client) handleFileChunk(file *os.File, chunk []byte) {

	_, err := file.Write(chunk)
	if err != nil {
		return
	}

}
func (client *Client) AddToServer() error {

	// fmt.Println("writing", global.MASTER_SERVER_URL)
	content := fmt.Sprintf("%s:%s:%s:%s \n", client.ID, client.IP, client.Port, client.Directory)

	conn, err := net.Dial("tcp4", global.MASTER_SERVER_URL)

	if err != nil {
		return fmt.Errorf("ERROR WHILE SENDING A REQUEST", err.Error())

	}

	binary.Write(conn, binary.BigEndian, global.ADD_CLIENT)

	conn.Write([]byte(content))

	err = conn.Close()

	if err != nil {
		return fmt.Errorf("error WHILE CLosing A REQUEST", err.Error())

	}

	return nil

}
func (client *Client) AnnounceFile(filePath string) error {
	fileId := cuid2.Generate()
	file, err := os.Open(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))
	fileId += ext
	if err != nil {
		return fmt.Errorf("error opening file %v", err.Error())

	}
	destPath := client.Directory + fileId

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %v", err.Error())

	}

	_, err = io.Copy(destFile, file)
	if err != nil {

		return fmt.Errorf("error copying file: %v", err.Error())
	}
	tracker.CreateTrackerFile(destFile, client.ID, fileId, client.Directory)

	err = protocol.AnnounceFile(fileId, file.Name(), client.ID)
	if err != nil {
		return err
	}
	defer destFile.Close()
	defer file.Close()

	return nil
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

func InitalizeClient(directory string) (ClientService, error) {
	var client ClientService = &Client{
		Directory: directory,
	}

	err := client.Start()

	return client, err
}
