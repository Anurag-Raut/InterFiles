package tracker

import (
	"bufio"
	"crypto/sha512"
	"fmt"
	"interfiles/global"
	"interfiles/verifier"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	dirPath := filepath.Join(currentDir, "tracker_files")
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return
	}

	filePath := filepath.Join(dirPath, fmt.Sprintf("%s_tracker.txt", fileId))
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return
	}

	date := time.Now().Format("02-01-2006 150405")
	fileInfo, err := ogFile.Stat()
	fmt.Println(fileInfo.Size(), "FILE SIZEEE ")
	if err != nil {
		fmt.Println("Error getting file size value", err.Error())
		return
	}
	totalChunks := (fileInfo.Size() / int64(global.CHUNK_SIZE)) + 1
	file.WriteString(fileId)
	file.WriteString("\n")
	file.WriteString(fmt.Sprintf("%s::%s::%d::%d::%s", filename, date, fileInfo.Size(), totalChunks, clientId))
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

		file.WriteString(strconv.Itoa(chunkNo) + "::" + strconv.Itoa(len(message)) + "::" + hashString)
		file.WriteString("\n")

		hasher.Reset()
		chunkNo++

		if lenOfChunk < uint64(global.CHUNK_SIZE) {

			global.SuccessPrint.Printf("Tracker file %s created in directory : %s \n", fileId+"_tracker.txt", directory)
			break
		}

	}

}

func GetMetadata(trackerFile *os.File) (*global.TrackerFileMetadata, error) {
	trackerFile.Seek(0, 0)
	trackerScanner := bufio.NewScanner(trackerFile)

	for i := 0; i < 1; i++ {
		if !trackerScanner.Scan() {
			if trackerScanner.Err() != nil {
				return nil, trackerScanner.Err()
			}
			return nil, fmt.Errorf("file has fewer than %d lines", i)
		} else {
			trackerScanner.Text()
		}
	}

	if !trackerScanner.Scan() {
		if trackerScanner.Err() != nil {
			return nil, trackerScanner.Err()
		}
		return nil, fmt.Errorf("file has fewer than %d lines", 2)
	} else {
		line := trackerScanner.Text()
		args := strings.Split(line, "::")
		var metadata global.TrackerFileMetadata
		metadata.Filename = args[0]
		metadata.Date = args[1]
		fmt.Println(args[2], "ARGS SIZE")
		size, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, err
		}
		metadata.Size = size
		fmt.Println(args[3], "ARGS total chunks")

		totalChunks, err := strconv.Atoi(args[3])
		if err != nil {
			return nil, err
		}
		metadata.TotalChunks = totalChunks
		metadata.ClientId = args[4]
		return &metadata, nil

	}

}
