package tracker

import (
	"crypto/sha512"
	"dfs/global"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nrednav/cuid2"
)

func CreateTracker(ogFile os.File, noOfBytes string, clientId string)	 {
	id := cuid2.Generate()
	filename := filepath.Base(ogFile.Name())
	file, err := os.Create(fmt.Sprintf("/home/anurag/projects/dfs/tracker_files/%s_tracker.txt", id))
	if err != nil {
		fmt.Println("Error creatinf file", err.Error())
		return
	}

	date := time.Now().String()
	fileInfo, err := file.Stat()
	if err!=nil{
		fmt.Println("Error getting file size value",err.Error())
		return
	}

	file.WriteString(id)
	file.WriteString(fmt.Sprintf("%s:%s:%d:%s", filename, date,fileInfo.Size()	, clientId))

	//chunks
	file.Seek(0, 0)
	totlBytes := 0
	var chunkNo = 0
	lenOfFilename := uint16(len(filename))
	hasher := sha512.New()

	for {
		buf := make([]byte, global.CHUNK_SIZE-global.HEADER_LEN-int(lenOfFilename))

		bytesRead, err := file.Read(buf)
		totlBytes += bytesRead
		// fmt.Println("CONTENT", string(buf), "BYTES READ", bytesRead)
		if err != nil {

			fmt.Println("ERROR OCCURED WHILE READING BYTES", err.Error())
			if err != io.EOF {

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

		hasher.Write(message)
		hash := hasher.Sum(nil)
		hashString := hex.EncodeToString(hash)

		// fmt.Println("ChunkNo", chunkNo, "filenamelen", lenOfFilename, "Sending no of bytes", len(message))

		file.WriteString(strconv.Itoa(chunkNo) + ":" + strconv.Itoa(len(message)) + ":" + hashString)

		hasher.Reset()
		chunkNo++

	}

}

