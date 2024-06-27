package protocol

import (
	"bufio"
	"dfs/global"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	// "time"

	"os"
)

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

func UploadFile(file *os.File, clientId string) {

	clients, err := getClients()
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return
	}

	for _, client := range clients {
		if client.ClientId == clientId {
			continue
		}
		conn, err := net.Dial("tcp", client.Ip+":"+client.Port)
		if err != nil {
			fmt.Println("ADSDAS", err.Error())
			continue
		}
		go uploadToClient(file, conn)

	}

}



func uploadToClient( file *os.File, conn net.Conn) {
	file.Seek(0, 0)
	totlBytes := 0
	var chunkNo = 0
	filename := filepath.Base(file.Name())
	lenOfFilename := uint16(len(filename))
	for {

		buf := make([]byte, global.CHUNK_SIZE-global.HEADER_LEN-int(lenOfFilename))

		bytesRead, err := file.Read(buf)
		totlBytes += bytesRead
		// fmt.Println("CONTENT", string(buf), "BYTES READ", bytesRead)
		if err != nil {

			fmt.Println("ERROR OCCURED WHILE READING BYTES", err.Error())
			if err == io.EOF {

			} else {
				fmt.Println("WE FUCKING RETURNING BOYS")
				return
			}
		}

		if err != nil {
			fmt.Println("ERROR SENDING FILR DIA:", err.Error())
			return
		}

		lenOfChunk := uint64(bytesRead)

		message := make([]byte, global.CHUNK_SIZE)

		binary.LittleEndian.PutUint16(message, lenOfFilename)

		binary.LittleEndian.PutUint64(message[2:], lenOfChunk)
		copy(message[2+8:], []byte(filename))
		copy(message[2+8+lenOfFilename:], buf)

		fmt.Println("ChunkNo", chunkNo, "filenamelen", lenOfFilename, "Sending no of bytes", len(message))
		chunkNo++
		conn.Write(message)

		ackBuf := make([]byte, 1)
		conn.Read(ackBuf)

		if bytesRead < len(buf) {
			//eof

			fmt.Println("totlBytes", totlBytes)

			break

		}

	}

	defer conn.Close()

}

func getClients() ([]global.Client, error) {

	resp, err := http.Get(global.MASTER_SERVER_URL + "/getClients")

	if err != nil {
		fmt.Println("ERROR", err.Error())
		return nil, err
	}

	var clients []global.Client

	err = json.NewDecoder(resp.Body).Decode(&clients)
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return nil, err
	}
	return clients, nil

}
