package protocol

import (
	"bufio"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"

	// "encoding/hex"
	"fmt"
	"interfiles/global"
	"interfiles/tracker"

	"interfiles/verifier"

	// "interfiles/verifier"

	// "interfiles/verifier"
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
		return

	}

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

func AnnounceFile(fileId, filename, clientId string) error {

	conn, err := net.Dial("tcp", global.MASTER_SERVER_URL)
	if err != nil {
		return fmt.Errorf("failed to connect to the server: %v", err)
	}
	binary.Write(conn, binary.BigEndian, global.ANNOUNCE)
	file_body := clientId + "::" + fileId + "::" + filename
	conn.Write([]byte(file_body))
	defer conn.Close()

	global.SuccessPrint.Println("Your file  is announced to master , you can check the status using \"stat -p path/to/tracker-file \" ")
	return nil
}

func UploadToClient(file *os.File, conn net.Conn) {
	reader := bufio.NewReader(conn)
	file.Seek(0, 0)
	totlBytes := 0
	var chunkNo = 0
	filename := filepath.Base(file.Name())
	lenOfFilename := uint16(len(filename))
	binary.Write(conn, binary.BigEndian, lenOfFilename)

	conn.Write([]byte(filename))
	for {

		message := make([]byte, global.CHUNK_SIZE)

		bytesRead, fileErr := file.Read(message)
		totlBytes += bytesRead
		// fmt.Println("CONTENT", string(buf), "BYTES READ", bytesRead)
		if fileErr != nil {

			// fmt.Println("ERROR OCCURED WHILE READING BYTES", err.Error())
			if fileErr == io.EOF {
				break
			} else {
				fmt.Println("WE FUCKING RETURNING BOYS")
				return
			}
		}

		// fmt.Println("ChunkNo", chunkNo, "filenamelen", lenOfFilename, "Sending no of bytes", uint32(bytesRead))
		chunkNo++
		var a uint32 = uint32(bytesRead)
		binary.Write(conn, binary.BigEndian, a)

		conn.Write(message)

		ackBuf := make([]byte, 1)
		_, err := reader.Read(ackBuf)
		if err != nil {
			fmt.Println("ERR", err)
			continue
		}
		// fmt.Println("ack len", n)

		if fileErr == io.EOF {
			//eof

			break

		}

	}

	defer conn.Close()

}

func HandShake(receiver, sender global.Client, fileId string) error {
	//sender handshake
	conn, err := net.Dial("tcp", sender.GetUrl())
	if err != nil {
		fmt.Println("Error occured ", err.Error())
		return err
	}
	conn.Write([]byte{byte(global.SENDER_HANDSHAKE)})
	binary.Write(conn, binary.BigEndian, uint16(len(fileId)))
	conn.Write([]byte(fileId))

	var status int8
	err = binary.Read(conn, binary.BigEndian, &status)
	if err != nil {
		fmt.Printf("Error reading status: %v", err)
		return nil
	}

	if status == 0 {
		return fmt.Errorf("sender handshake failed")
	}

	conn.Close()

	conn, err = net.Dial("tcp", receiver.GetUrl())
	if err != nil {
		fmt.Println("Error occured ", err.Error())
		return err
	}

	conn.Write([]byte{byte(global.RECEIVER_HANDSHAKE)})

	err = binary.Read(conn, binary.BigEndian, &status)
	if err != nil {
		fmt.Printf("Error reading status: %v", err)
		return nil
	}

	if status == 0 {
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

func RequestToPullFile(receiver global.Client, file global.File) {
	// fmt.Println("REQUESTING FIlE")
	//finding potential senders to get the file
	for _, sender := range file.Clients {
		// fmt.Println("THIS DUDE IS POTENTIAL sender", sender.GetUrl())
		// protocol.HandShake(receiver,sender)

		conn, err := net.Dial("tcp", receiver.GetUrl())
		if err != nil {
			fmt.Println("ERROR", err.Error())
			return

		}
		err = HandShake(receiver, sender, file.ID)
		if err != nil {
			fmt.Println("HANDSHAKE FAILED TRY ANOTHER ONE")
			continue

		}

		binary.Write(conn, binary.BigEndian, global.REQUEST_TO_PULL_FILE)
		// fmt.Println(sender.GetUrl()+":"+file.ID, sender.Ip, "YOO asd")
		conn.Write([]byte(sender.Ip + "::" + sender.Port + "::" + file.ID))

		conn.Close()
		break

	}

}

func GetSendersFromMaster(fileId string) ([]global.Client, error) {

	conn, err := net.Dial("tcp", global.MASTER_SERVER_URL)
	if err != nil {
		fmt.Println("Error getting clients", err.Error())
		return nil, err

	}

	binary.Write(conn, binary.BigEndian, global.GET_SENDERS_FOR_FILE)

	binary.Write(conn, binary.BigEndian, uint16(len(fileId)))

	conn.Write([]byte(fileId))
	reader := bufio.NewReader(conn)

	body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("ERROR while reading data from ", err.Error())
		return nil, err
	}
	var clients []global.Client
	// fmt.Println(string(body), "GET SENDERS FROM MASTER")
	data := strings.Split(string(body), "::")
	if len(data) < 3 {
		return nil, fmt.Errorf("CONTAINS LESS ELEMENTS ")
	} else {

		for i := 0; i < len(data); i += 3 {
			clientId := data[i]
			clientIp := data[i+1]
			clientPort := data[i+2]
			newClient := &global.Client{
				ClientId: clientId,
				Ip:       clientIp,
				Port:     clientPort,
			}
			clients = append(clients, *newClient)

		}
	}
	conn.Close()
	return clients, nil

}

func DownloadFile(receiver global.Client, fileId string, clients []global.Client, trackerFile *os.File) {
	file, err := os.OpenFile(receiver.Directory+"newww"+fileId, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("Error while opening the write file", err.Error())
		return
	}
	var chunksWanted []string = nil
	metadata, err := tracker.GetMetadata(trackerFile)
	if err != nil {
		global.ErrorPrint.Println("ERROR WHILE PARSING METADATA:", err.Error())
	}
	for _, client := range clients {
		HandShake(receiver, client, fileId)
		chunksWanted, err = downloadFileFromClient(client, file, chunksWanted, trackerFile, fileId, metadata)
		if len(chunksWanted) > 0 {
			continue
		} else if err == nil {
			global.SuccessPrint.Println("DOWNLOADED FILE FROM ", client.GetUrl())
			break
		} else if err != nil {
			continue
		} else if chunksWanted == nil {
			continue
		} else if len(chunksWanted) == 0 {
			break
		}

	}

	defer file.Close()

}

func downloadFileFromClient(sender global.Client, file *os.File, chunksWanted []string, trackerFile *os.File, fileId string, metadata *global.TrackerFileMetadata) ([]string, error) {
	//receiving side
	fmt.Println("TRYING TO DOWNLOAD FROM A SENDER", sender.GetUrl())
	conn, err := net.Dial("tcp", sender.GetUrl())
	if err != nil {
		return chunksWanted, err
	}

	binary.Write(conn, binary.BigEndian, global.DOWNLOAD_FILE)
	fileIdLen := uint16(len(fileId))
	binary.Write(conn, binary.BigEndian, fileIdLen)
	conn.Write([]byte(fileId))
	if chunksWanted == nil {
		//data flag
		binary.Write(conn, binary.BigEndian, uint8(0))

	} else {
		binary.Write(conn, binary.BigEndian, uint8(1))
		chunksFormated := strings.Join(chunksWanted, "::")
		fmt.Println("CHUNS WANTED ", chunksFormated)
		len := uint16(len(chunksFormated))
		binary.Write(conn, binary.BigEndian, len)
		conn.Write([]byte(chunksFormated))
	}

	reader := bufio.NewReader(conn)

	for {
		buf := make([]byte, global.CHUNK_SIZE+8)
		var noOfBytes uint32
		binary.Read(reader, binary.BigEndian, &noOfBytes)
		bytesRead, buferr := io.ReadFull(reader, buf)

		if buferr != nil {
			if buferr == io.EOF {
			} else {

				fmt.Println("Error while reading a chunk", err.Error())
				return chunksWanted, buferr
			}
		}

		if noOfBytes == 0 {
			// fmt.Println("ABEEEEE")
			break
		}
		// fmt.Println(noOfBytes, "AADSDSD")
		chunkNo := binary.BigEndian.Uint64(buf)
		fmt.Printf("\rDownloading chunk %d out of %d ", chunkNo+1, metadata.TotalChunks)
		// fmt.Println("CHUBKNO", chunkNo)

		data := buf[8 : 8+noOfBytes]
		// fmt.Println("STRING DATA RECEIVED : ", string(buf))
		err = writeChunkToFile(chunkNo, data, file)
		// chunksWanted=removeElement(chunksWanted,strconv.Itoa(int(chunkNo)))
		if err != nil {
			fmt.Println("ALRIGHT BOYZ WE DONE ", err.Error())
			break

		}
		conn.Write([]byte{1})

		if buferr == io.EOF {
			fmt.Println("WE DONE BOYSZZ", bytesRead)
			break
		}

	}

	_, newChunksWanted, err := verifier.VerifyFile(file, trackerFile)
	// fmt.Println("NEW CHUNKS WANTED", newChunksWanted)
	if len(newChunksWanted) > 0 {
		// file.Truncate(0)
		// file.Seek(0, 0)
		return newChunksWanted, nil
	}
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return nil, err
	}
	return []string{}, nil

}
func replaceChunk(originalFile *os.File, chunkOffset, chunkSize int64, newChunk []byte) error {

	originalFile.Seek(0, 0)
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "tempfile")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	// Copy the part before the chunk
	_, err = io.CopyN(tempFile, originalFile, chunkOffset)
	if err != nil {

		return err
	}

	// fmt.Println("befo",before)

	// Write the new chunk to the temporary file
	_, err = tempFile.Write(newChunk)
	if err != nil {

		return err
	}
	// fmt.Println(chunkbyte,"chunkBytes")

	// Skip the chunk in the original file
	_, err = originalFile.Seek(chunkOffset+chunkSize, io.SeekStart)
	if err != nil {

		return err
	}
	// fmt.Println("SEEKED ",seek)

	// Copy the rest of the original file to the temporary file
	_, err = io.Copy(tempFile, originalFile)
	if err != nil {

		return err
	}
	// fmt.Println("after ",after)

	// Close the files to ensure data integrity

	// Replace the original file with the temporary file
	originalFile.Truncate(0)
	originalFile.Seek(0, io.SeekStart)
	tempFile.Seek(0, io.SeekStart)
	if _, err := io.Copy(originalFile, tempFile); err != nil {

		return err
	}

	return nil

}

func writeChunkToFile(chunkno uint64, chunk []byte, file *os.File) error {
	offset := int64(chunkno) * int64(global.CHUNK_SIZE)
	// fmt.Println("OVRRIDING OFFSET",offset)
	err := replaceChunk(file, offset, int64(global.CHUNK_SIZE), chunk)
	if err != nil {
		return err
	}

	return nil

}

func SendFile(reader *bufio.Reader, conn net.Conn, sender global.Client) {
	var flag uint8
	var fileIdlen uint16
	binary.Read(reader, binary.BigEndian, &fileIdlen)
	fileIdbuf := make([]byte, fileIdlen)
	_, err := reader.Read(fileIdbuf)
	if err != nil {
		return
	}

	fileId := string(fileIdbuf)
	binary.Read(reader, binary.BigEndian, &flag)

	filetoSend, err := os.Open(sender.Directory + fileId)
	filetoSend.Seek(0, 0)
	if err != nil {
		return
	}
	if flag == 1 {

		// receiver is gonna send which chunks they want
		var len uint16
		binary.Read(conn, binary.BigEndian, &len)

		bytes := make([]byte, len)

		io.ReadFull(reader, bytes)

		chunks := strings.Split(string(bytes), "::")
		fmt.Println("CHUNKS STRING ", chunks)
		for _, chunkNo := range chunks {
			chunkNumber, err := strconv.Atoi(chunkNo)
			if err != nil {
				continue

			}

			buf := make([]byte, global.CHUNK_SIZE+8)
			binary.BigEndian.PutUint64(buf, uint64(chunkNumber))
			bytesRead, readErr := filetoSend.ReadAt(buf[8:], int64(chunkNumber)*int64(global.CHUNK_SIZE))

			if readErr != nil {

				// fmt.Println("ERROR OCCURED WHILE READING BYTES", readErr.Error())
				if readErr == io.EOF {

				} else {
					return
				}
			}

			message := buf
			binary.Write(conn, binary.BigEndian, uint32(bytesRead))
			hasher := sha512.New()
			hasher.Write(message[8 : 8+bytesRead])

			fmt.Println("sending chunk", chunkNo, "Sending no of bytes", bytesRead, "HASH", hex.EncodeToString(hasher.Sum(nil)))
			// fmt.Println("ChunkNo", chunkNo, "Sending no of bytes", uint32(bytesRead))
			_, err = conn.Write(message)
			if err != nil {
				break
			}
			ackBuf := make([]byte, 1)
			reader.Read(ackBuf)

			if readErr == io.EOF {
				//eof
				conn.Close()
				break

			}

		}

	} else {
		chunkNo := uint64(0)

		for {

			message := make([]byte, global.CHUNK_SIZE+8)
			binary.BigEndian.PutUint64(message, chunkNo)
			bytesRead, readErr := filetoSend.Read(message[8:])
			// fmt.Println("CONTENT", string(buf), "BYTES READ", bytesRead)
			if readErr != nil {

				// fmt.Println("ERROR OCCURED WHILE READING BYTES", readErr.Error())
				if readErr == io.EOF {

				} else {
					return
				}
			}

			binary.Write(conn, binary.BigEndian, uint32(bytesRead))
			hasher := sha512.New()
			hasher.Write(message[8 : 8+bytesRead])

			// fmt.Println("sending chunk", chunkNo, bytesRead, "HASH", hex.EncodeToString(hasher.Sum(nil)))
			chunkNo++
			_, err := conn.Write(message)
			if err != nil {
				break
			}
			ackBuf := make([]byte, 1)
			reader.Read(ackBuf)

			if readErr == io.EOF {
				//eof
				break

			}

		}

	}

	defer conn.Close()

}
