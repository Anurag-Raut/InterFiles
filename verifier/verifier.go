package verifier

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"interfiles/global"
	"io"
	"os"
	"strconv"
	"strings"
	// "text/scanner"
)

func HashChunk(hasher hash.Hash, content []byte) string {

	hasher.Write(content)
	hash := hasher.Sum(nil)
	hashString := hex.EncodeToString(hash)
	hasher.Reset()
	return hashString
}

func VerifyFile(file, trackerFile *os.File) (bool,[]string, error) {
	file.Seek(0, 0)
	trackerFile.Seek(0,0)
	hasher := sha512.New()
	chunkNo := 0
	trackerScanner := bufio.NewScanner(trackerFile)
	var chunksWanted []string
	for i := 0; i < 2; i++ {
		if !trackerScanner.Scan() {
			if trackerScanner.Err() != nil {
				return false ,nil, trackerScanner.Err()
			}
			return false,nil, fmt.Errorf("file has fewer than %d lines", i)
		}else {
			trackerScanner.Text()
		}
	}

	for {
		buf := make([]byte, global.CHUNK_SIZE)
		bytesReaded, bufError := file.Read(buf)
		if bufError != nil  && bufError!= io.EOF{
			fmt.Println("ERROR", bufError.Error())
			return false,nil, bufError
		}


		if !trackerScanner.Scan() {
			return false,nil, fmt.Errorf("error ")

		}

		line := trackerScanner.Text()
		trackerLineData := strings.Split(line, ":")
		trackerChunkno := trackerLineData[0]
		// chunkSize:=trackerLineData[1]
		hashHex := trackerLineData[2]
		// fmt.Println(trackerChunkno, "trackerChunknumber")
		buf = buf[:bytesReaded]
		hashBytes, err := hex.DecodeString(hashHex)
		if err != nil {
			return false,nil, err

		}
		hasher.Reset()
		hasher.Write(buf)
		hash := hasher.Sum(nil)
		trackerChunknoInt, err := strconv.Atoi(trackerChunkno)
		if err != nil {

			return false,nil, fmt.Errorf("error converting tracker chunk number to integer: %v", err)
		}

		if trackerChunknoInt != chunkNo {
			
			return false,nil, fmt.Errorf("order mismatch trackerChunkNo : %d , actual chunk : %d",trackerChunknoInt,chunkNo)

		}
		chunkNo++


		if !bytes.Equal(hash, hashBytes) {
			fmt.Println("NOT MATCHING",chunkNo-1,hashHex)
			chunksWanted=append(chunksWanted, trackerChunkno)
			continue
		}
		hasher.Reset()

		if bytesReaded < int(global.CHUNK_SIZE)  || bufError==io.EOF{

			break
		}


	}

	return true,chunksWanted, nil

}

func VerifyChunk(chunk []byte, chunkno uint64, trackerFile *os.File, hasher hash.Hash) error {
	// fmt.Println("VERIFYING CHUNK ", chunkno)
	trackerFile.Seek(0, 0)
	trackerScanner := bufio.NewScanner(trackerFile)
	// fmt.Println("chunko", chunkno)
	for i := 0; i < 2; i++ {
		if !trackerScanner.Scan() {
			if trackerScanner.Err() != nil {
				return trackerScanner.Err()
			}
			return fmt.Errorf("file has fewer than %d lines", i)
		} else {
			// fmt.Println("skipping Line", trackerScanner.Text())
			trackerScanner.Text()
		}

	}

	for i := uint64(0); i < chunkno; i++ {
		if !trackerScanner.Scan() {
			if trackerScanner.Err() != nil {
				return trackerScanner.Err()
			}
			return fmt.Errorf("file has fewer than %d lines", i)
		} else {
			trackerScanner.Text()
			// fmt.Println("skipping Line", trackerScanner.Text())
		}
	}

	if !trackerScanner.Scan() {
		if trackerScanner.Err() != nil {
			return trackerScanner.Err()
		}
		return fmt.Errorf("file has fewer  lines")
	} else {
		trackerScanner.Text()
	}
	trackerChunkData := strings.Split(trackerScanner.Text(), ":")
	trackerchunkNo := trackerChunkData[0]
	chunkHashHex := trackerChunkData[2]
	trackerChunkNoInt, err := strconv.Atoi(trackerchunkNo)
	if err != nil {
		fmt.Println(trackerChunkNoInt, "trackerChunkInt")
		return err
	}
	if chunkno != uint64(trackerChunkNoInt) {

		return fmt.Errorf("chunks number different ")

	}
	hashBytes, err := hex.DecodeString(chunkHashHex)
	if err != nil {
		return err

	}
	hasher.Reset()
	hasher.Write(chunk)
	calcHash := hasher.Sum(nil)

	hasher.Reset()

	if !bytes.Equal(hashBytes, calcHash) {
		fmt.Println("tracker chunk hex", chunkHashHex, "calc hex", hex.EncodeToString(calcHash))
		return fmt.Errorf("chunks not match ing ")

	}

	return nil

}
