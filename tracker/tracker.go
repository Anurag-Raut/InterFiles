package tracker

import (
	"crypto/sha512"
	"dfs/global"
	"dfs/verifier"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func CreateTrackerFile(ogFile *os.File, clientId string, fileId string) {
	fmt.Println("HELLO FROM TRACKER CREATER")
	ogFile.Seek(0, 0)
	filename := filepath.Base(ogFile.Name())
	file, err := os.Create(fmt.Sprintf("/home/anurag/projects/dfs/tracker_files/%s_tracker.txt", fileId))
	if err != nil {
		fmt.Println("Error creatinf file", err.Error())
		return
	}

	date := time.Now().String()
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file size value", err.Error())
		return
	}

	file.WriteString(fileId)
	file.WriteString("\n")
	file.WriteString(fmt.Sprintf("%s:%s:%d:%s", filename, date, fileInfo.Size(), clientId))
	file.WriteString("\n")

	fmt.Println("once HELLO FROM TRACKER CREATER")

	//chunks
	
	totlBytes := 0
	var chunkNo = 0
	hasher := sha512.New()

	for {
		message := make([]byte, global.CHUNK_SIZE)

		bytesRead, err := ogFile.Read(message)
		totlBytes += bytesRead
		// fmt.Println("CONTENT", string(buf), "BYTES READ", bytesRead)
		if err != nil {

			fmt.Println("ERROR OCCURED WHILE READING BYTES", err.Error())
			if err != io.EOF {

				fmt.Println("WE FUCKING TRACKER BOYS", err.Error())
				break
			}
		}

		if err != nil {
			fmt.Println("ERROR SENDING TRACKER FILR DIA:", err.Error())
			return
		}
		// fmt.Println("once HELLO FROM TRACKER CREATER")

		lenOfChunk := uint64(bytesRead)


		

	
		hashString := verifier.HashChunk(hasher,message)

		fmt.Println("TRACKER SAN ChunkNo", chunkNo, "lenOfChunk", lenOfChunk, "Sending no of bytes", len(message))

		file.WriteString(strconv.Itoa(chunkNo) + ":" + strconv.Itoa(len(message)) + ":" + hashString)
		file.WriteString("\n")

		hasher.Reset()
		chunkNo++

		if lenOfChunk<uint64(global.CHUNK_SIZE){

			break
		}

	}

}



