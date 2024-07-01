package tracker

import (
	"crypto/sha512"
	"fmt"
	"interfiles/global"
	"interfiles/verifier"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func CreateTrackerFile(ogFile *os.File, clientId string, fileId string, directory string) {
	fmt.Println("Creating you tracker file")
	ogFile.Seek(0, 0)
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("failed to get current directory: %w", err)
		return
	}
	filename := filepath.Base(ogFile.Name())

	file, err := os.Create(filepath.Join(currentDir, fmt.Sprintf("tracker_files/%s_tracker.txt", fileId)))
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

			// fmt.Println("ERROR OCCURED WHILE READING BYTES", err.Error())
			if err != io.EOF {

				fmt.Println("WE FUCKING TRACKER BOYS", err.Error())
				break
			}
		}

		if err != nil {
			fmt.Println("ERROR SENDING TRACKER FILR DIA:", err.Error())
			return
		}
		lenOfChunk := uint64(bytesRead)

		message = message[:bytesRead]

		hashString := verifier.HashChunk(hasher, message)

		file.WriteString(strconv.Itoa(chunkNo) + ":" + strconv.Itoa(len(message)) + ":" + hashString)
		file.WriteString("\n")

		hasher.Reset()
		chunkNo++

		if lenOfChunk < uint64(global.CHUNK_SIZE) {

			fmt.Printf("Tracker file %s created in directory : %s \n", fileId+"_tracker.txt", directory)
			break
		}

	}

}
