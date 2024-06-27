package client

import (
	"bufio"
	"bytes"
	"dfs/global"
	"dfs/protocol"
	"encoding/binary"
	"fmt"
	"io"

	// "io"
	"net"
	"net/http"
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
	handleFileChunk(filename string ,chunk []byte)
}

type Client struct {
	ID        string
	IP        string
	Port      int
	listener  net.Listener
	Directory string
}

func (client *Client) Start() {
	basePort := 8080
	maxRetries := 12
	var listener net.Listener
	var err error

	for i := 0; i < maxRetries; i++ {
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
	client.Port = addr.Port
	client.IP = "127.0.0.1"
	client.ID = cuid2.Generate()
	client.Directory = "/home/anurag/projects/dfs/dum/"
	client.AddToServer()

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

func (client *Client) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	totalLen := 0
	fmt.Println("BROTHER CHILLLLL ",conn.RemoteAddr().Network())
	var chunkNo=0
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
		
		filenameLen:=int(binary.LittleEndian.Uint16(chunk))
		fmt.Println("chunkNo",chunkNo,"number of bytes", n,"filenameLen",filenameLen)
		chunkNo++
		filename:=string(chunk[global.HEADER_LEN:global.HEADER_LEN+filenameLen])
		data:=chunk[2+8+filenameLen:]
	
	
		// fmt.Println("RECIVED FILENAME",filename,"CHUNK",chunk)
	
		client.handleFileChunk(filename,data)
		totalLen += n
		ackBuf := []byte{1}
		conn.Write(ackBuf)
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			fmt.Println("totalLen",totalLen)
			// if err == io.EOF {
			// 	break // End of the message
			// }
			break
		}
		


	}

	

	

}
func (client *Client) handleFileChunk(filename string,chunk []byte) {

    file, err := os.OpenFile(client.Directory+filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
         fmt.Println("failed to create file", err.Error())
		 return
    }
	defer file.Close()
	_,err=file.Write(chunk)
	if err != nil {
		fmt.Println("failed to write chunk to file:", err.Error())

		return
	}





}
func (client *Client) AddToServer() {

	fmt.Println("writing")
	content := fmt.Sprintf("%s:%s:%d:%s \n", client.ID, client.IP, client.Port, client.Directory)

	req, err := http.NewRequest("POST", global.MASTER_SERVER_URL+"/addClient", bytes.NewBufferString(content))

	if err != nil {
		fmt.Println("ERROR WHILE SENDING A REQUEST", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/text")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("client added succesfully")
	} else {
		fmt.Printf("Failed to send data, status code: %d\n", resp.StatusCode)
	}

}
func (client *Client) UploadFile(filepath string) {
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0)

	if err != nil {
		fmt.Println("ERROR OPENING FILE", err.Error())

	}

	protocol.UploadFile(file,client.ID)
}

func InitalizeClient() ClientService {
	var client ClientService = &Client{}

	client.Start()

	return client
}
