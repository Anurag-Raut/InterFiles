package protocol

import (
	"bufio"
	"dfs/global"
	"encoding/binary"
	"fmt"
	"io"
	"net"
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

func AnnounceFile(fileId,filename, clientId string) {

	conn,err:=net.Dial("tcp",global.MASTER_SERVER_URL)
	if err != nil {
		fmt.Println("Failed to connect to the server:", err)
		return
	}
	binary.Write(conn,binary.BigEndian,global.ANNOUNCE)
	file_body:=clientId+":"+fileId+":"+filename
	conn.Write([]byte(file_body))
	conn.Close()



	

}



func UploadToClient( file *os.File, conn net.Conn) {
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

func HandShake(receiver , sender global.Client) error {
	//sender handshake
	conn,err :=net.Dial("tcp",sender.GetUrl())
	if err !=nil {
		fmt.Println("Error occured ",err.Error())
		return err
	}
	conn.Write([]byte{byte(global.SENDER_HANDSHAKE)})
	var status int8
	err = binary.Read(conn, binary.BigEndian, &status)
	if err != nil {
		fmt.Printf("Error reading status: %v", err)
		return nil
	}

	if status==0{
		return fmt.Errorf("sender handshake failed")
	}

	conn.Close()

	conn,err= net.Dial("tcp",receiver.GetUrl())
	if err !=nil {
		fmt.Println("Error occured ",err.Error())
		return err
	}

	conn.Write([]byte{byte(global.CLIENT_HANDSHAKE)})

	
	err = binary.Read(conn, binary.BigEndian, &status)
	if err != nil {
		fmt.Printf("Error reading status: %v", err)
		return nil
	}

	if status==0{
		return fmt.Errorf("receiver handshake failed")
	}

	conn.Close()






	return nil
	

}

// func DownloadFile(file global.File,sender global.Client){
// 	// receiver <- sender
// 	// make connection to sneder and ask for file

// 	conn,err:=net.Dial("tcp",sender.GetUrl())

// 	binary.Write(conn,binary.LittleEndian,global.SENDER_PULLFILE)

	






// }

func RequestFile(receiver global.Client,file global.File){
	fmt.Println("REQUESTING FIlE")
	for _,sender:=range file.Clients {
		fmt.Println("THIS DUDE IS POTENTIAL sender")
		// protocol.HandShake(receiver,sender)
		
		conn,err:=net.Dial("tcp",receiver.GetUrl())
		if err !=nil {
			fmt.Println("ERROR"  ,err.Error())
			return

		}

		binary.Write(conn,binary.BigEndian,global.REQUEST_FILE)
		fmt.Println(sender.GetUrl()+":"+file.ID,sender.Ip,"YOO asd")
		conn.Write([]byte(sender.GetUrl()+":"+file.ID))


		conn.Close()
		

		


		


	}

}




