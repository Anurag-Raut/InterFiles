package protocol

import (
	"bufio"
	"crypto/sha512"
	"dfs/global"
	"dfs/verifier"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"strings"

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
	defer conn.Close()

	fmt.Println("DONE WITH ANNOUNCING FILE")



	

}



func UploadToClient( file *os.File, conn net.Conn) {
	file.Seek(0, 0)
	totlBytes := 0
	var chunkNo = 0
	filename := filepath.Base(file.Name())
	lenOfFilename := uint16(len(filename))
	binary.Write(conn,binary.BigEndian,lenOfFilename)

	conn.Write([]byte(filename))

	for {

		buf := make([]byte, global.CHUNK_SIZE)

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


		message := make([]byte, global.CHUNK_SIZE)


		copy(message, buf)

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




func GetSendersFromMaster(fileId string) ([]global.Client,error) {

	conn,err:=net.Dial("tcp",global.MASTER_SERVER_URL)
	if err != nil {
		fmt.Println("Error getting clients",err.Error())
		return nil,err

	}

	binary.Write(conn,binary.BigEndian,global.GET_SENDERS_FOR_FILE)
	
	binary.Write(conn,binary.BigEndian,uint16(len(fileId)))

	conn.Write([]byte(fileId))
	reader:=bufio.NewReader(conn)

	body,err:=io.ReadAll(reader)
	if err != nil {
		fmt.Println("ERROR while reading data from ",err.Error())
		return nil,err
	}
	var clients []global.Client
	fmt.Println(string(body),"GET SENDERS FROM MASTER")
	data:=strings.Split(string(body), ":")
	if len(data) <3 {
		fmt.Println("CONTAINS LESS ELEMENTS ")
	}else{

		for i := 0; i < len(data); i += 3 {
			clientId:=data[i]
			clientIp:=data[i+1]
		clientPort:=data[i+2]
		newClient:=&global.Client{
			ClientId: clientId,
			Ip: clientIp,
			Port: clientPort,
		}
		clients = append(clients, *newClient)
		
	}
}
	conn.Close()
	return clients,nil




}

func DownloadFile(receiver global.Client,fileId string , clients []global.Client,trackerFile *os.File) {
	file,err:=os.OpenFile(receiver.Directory+fileId,os.O_WRONLY|os.O_APPEND,0644)
	if err != nil {
		fmt.Println("Error while opening the write file",err.Error())
		return
	}
	var chunksWanted []string = nil
	fmt.Println("Starting downloading file ")

	for _,client:= range clients	{


		downloadFileFromClient(client,file,chunksWanted,trackerFile,fileId)


	}




}

func downloadFileFromClient(sender global.Client ,file *os.File,chunksWanted []string,trackerFile *os.File,fileId string)([]string,error) {
	//receiving side
	hasher:=sha512.New()
	conn,err:=net.Dial("tcp",sender.GetUrl())
		if err != nil {
			fmt.Println("Erorr while connecting to a sender ")

			return chunksWanted,err
		}



		binary.Write(conn,binary.BigEndian,global.DOWNLOAD_FILE)
		fileIdLen:=uint16(len(fileId))
		binary.Write(conn,binary.BigEndian,fileIdLen)
		conn.Write([]byte(fileId))
		if chunksWanted == nil {
			//data flag
			binary.Write(conn, binary.BigEndian, uint8(0))


		}else {
			binary.Write(conn, binary.BigEndian, uint8(1))
			chunksFormated:=strings.Join(chunksWanted, ":")
			conn.Write([]byte(chunksFormated))
			conn.Write([]byte{'\x04'})
		}

		var newChunksWanted []string

		reader:=bufio.NewReader(conn)

		for {
			buf:=make([]byte,global.CHUNK_SIZE+8)

			bytesRead,err:=reader.Read(buf)
			if err != nil {
				fmt.Println("Error while reading a chunk",err.Error())
				return chunksWanted,err
			}
			

			chunkNo :=  binary.BigEndian.Uint64(buf)


			data:=buf[8:]


			err=writeChunkToFile(chunkNo,data,file,trackerFile,hasher)
			if err != nil {
				fmt.Println("ERROR while verifying chunk",err)
				newChunksWanted=append(newChunksWanted, fmt.Sprintf("%d",chunkNo))

			}



			



			






			if bytesRead< global.CHUNK_SIZE {
				
				break;
			}



			


		}




		if len(newChunksWanted)>0 {
			return newChunksWanted,nil
		}
		return nil,nil

}


func writeChunkToFile(chunkno uint64,chunk []byte,file *os.File,trackerFile *os.File,hasher hash.Hash) error {

	err:=verifier.VerifyChunk(chunk,chunkno,trackerFile,hasher)
	if err != nil {
		return err
	}
	file.Seek(int64(chunkno)*int64(global.CHUNK_SIZE),0)
	file.Write(chunk)

	return nil

}


func SendFile(reader *bufio.Reader, conn net.Conn ,sender global.Client){
	var flag uint8
	var fileIdlen int16
	binary.Read(reader,binary.BigEndian,&fileIdlen)
	fileIdbuf:=make([]byte,fileIdlen)
	fileIdBytesRead,err:=reader.Read(fileIdbuf) 
	if err != nil {
		fmt.Println("error readind fileIdbuf")
		return
	}
	
	fileId:=string(fileIdBytesRead)
	fmt.Println(fileId,"FILEIDDD IT ISSS ")
	binary.Read(reader,binary.BigEndian,&flag)

	filetoSend,err:=os.Open(sender.Directory+fileId)
	if err !=nil {
		fmt.Println("error while opening file to send",err.Error())
		return 
	}
	if flag==1 {

		// receiver is gonna send which chunks they want

		bytes,err:=reader.ReadBytes('\x04')
		if err != nil {
			fmt.Println("errorrorr reding chunss")
		}
		chunks:=strings.Split(string(bytes), ":")

		for _,chunkNo := range chunks {
			chunkNumber,err:=strconv.Atoi(chunkNo)
			if err!=nil {
				fmt.Println("Error converting chunk string to chunk number")
				continue

			}

			buf := make([]byte, global.CHUNK_SIZE+8)
			binary.BigEndian.PutUint64(buf,uint64(chunkNumber))
			_, err = filetoSend.ReadAt(buf[8:],int64(global.CHUNK_SIZE*chunkNumber))
			
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
	
	
			message := make([]byte, global.CHUNK_SIZE)
	
	
			copy(message, buf)
	
			fmt.Println("ChunkNo", chunkNo, "Sending no of bytes", len(message))
			conn.Write(message)
	
			ackBuf := make([]byte, 1)
			conn.Read(ackBuf)
	
			
	
		}




	}else{
		chunkNo:=uint64(0);
		
		for {
			
			buf := make([]byte, global.CHUNK_SIZE+8)
			binary.BigEndian.PutUint64(buf,uint64(chunkNo))
			bytesRead, err := filetoSend.Read(buf[8:])
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
	
	
			message := make([]byte, global.CHUNK_SIZE)
	
	
			copy(message, buf)
	
			fmt.Println("ChunkNo", chunkNo, "Sending no of bytes", len(message))
			chunkNo++
			conn.Write(message)
	
			ackBuf := make([]byte, 1)
			conn.Read(ackBuf)
	
			if bytesRead < len(buf) {
				//eof
	
	
				break
	
			}
	
		}
	

	}

	conn.Close()




}


