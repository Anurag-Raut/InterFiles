package protocol

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)
var CHUNK_SIZE=1024 * 1024

func StartServer() {
	fmt.Println("starting server on port 8080")
	var listener net.Listener

	basePort := 8080
	maxRetries := 12

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

	addClient(listener)

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connections", err.Error())

			return
		}

		go handleConnection(conn)

	}

}

func addClient(listener net.Listener) {
	addr := listener.Addr().(*net.TCPAddr)
	ip := "127.0.0.1"
	port := addr.Port
	directory := "/home/anurag/projects/dfs/dump"
	fmt.Printf("Server is listening on %s:%d\n", ip, addr.Port)
	file, err := os.OpenFile("/home/anurag/projects/dfs/clients.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		if os.IsExist(err) {
			fmt.Println("File already exists..")
		} else {

			fmt.Println("Error opening file:", err)
			return
		}
	}
	fmt.Println("writing")
	content:=fmt.Sprintf("%s:%d:%s \n", ip, port, directory)
	fmt.Println("WRITING ",content)
	_,err=file.WriteString(content)
	if err!= nil {
		fmt.Println("ERROR while writin to client.txt",err.Error())


	}

	defer file.Close()

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var msgLen int32

	err := binary.Read(reader, binary.BigEndian, &msgLen)
	if err != nil {
		fmt.Println("error reading message len:", err.Error())
		return
	}

	buf := make([]byte, msgLen)

	_, err = reader.Read(buf)
	if err != nil {
		fmt.Println("error reading message:", err.Error())

		return

	}

	fmt.Println(string(buf))

}

func SendMessage(message string) {
	// Connect to the server
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println("Failed to connect to the server:", err)
		return
	}
	defer conn.Close()

	// Create a buffered writer
	writer := bufio.NewWriter(conn)
	defer writer.Flush()

	// Write the message length
	msgLen := int32(len(message))
	err = binary.Write(writer, binary.BigEndian, msgLen)
	if err != nil {
		fmt.Println("Failed to write message length:", err)
		return
	}

	// Write the message payload
	_, err = writer.WriteString(message)
	if err != nil {
		fmt.Println("Failed to write message:", err)
		return
	}

	fmt.Println("Message sent successfully")
}

func UploadFile(filename string){

	//check file exists
	file,err:= os.OpenFile(filename,os.O_RDONLY,0)
	if err!=nil {
		fmt.Println("Error opening file . Please try again")

		return

	}
	

	buf:=make([]byte,CHUNK_SIZE)
	for{
		bytesRead, err := file.Read(buf)

		if err!=nil {
			fmt.Println("ERROR OCCURED WHILE READING BYTES")
			return
		}

		if bytesRead<CHUNK_SIZE{
			//eof

		}


	}


	

}
