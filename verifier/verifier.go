package verifier

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"dfs/global"
	"encoding/hex"
	"fmt"
	"hash"
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

func VerifyFile(file, trackerFile *os.File) (bool, error) {

	file.Seek(0, 0)
	hasher := sha512.New()
	chunkNo := 0
	trackerScanner := bufio.NewScanner(trackerFile)

	for i := 0; i < 2; i++ {
		if !trackerScanner.Scan() {
			if trackerScanner.Err() != nil {
				return false, trackerScanner.Err()
			}
			return false, fmt.Errorf("file has fewer than %d lines", i)
		}
	}

	for {
		buf := make([]byte, global.CHUNK_SIZE)
		bytesReaded, err := trackerFile.Read(buf)
		if err != nil {
			fmt.Println("ERROR", err.Error())
			return false, err
		}

		if !trackerScanner.Scan() {
			return false, fmt.Errorf("error ")

		}

		line := trackerScanner.Text()
		trackerLineData := strings.Split(line, ":")
		trackerChunkno := trackerLineData[0]
		// chunkSize:=trackerLineData[1]
		hashHex := trackerLineData[2]
		fmt.Println(trackerChunkno, "trackerChunknumber")

		hashBytes, err := hex.DecodeString(hashHex)
		if err != nil {
			return false, err

		}
		hasher.Write(buf)
		hash := hasher.Sum(nil)
		trackerChunknoInt, err := strconv.Atoi(trackerChunkno)
		if err != nil {

			return false, fmt.Errorf("error converting tracker chunk number to integer: %v", err)
		}

		if trackerChunknoInt != chunkNo {
			return false, fmt.Errorf("chunk mismatch: %v", err)

		}

		if !bytes.Equal(hash, hashBytes) {
			return false, fmt.Errorf("error while verifying chunk %s ", trackerChunkno)
		}
		hasher.Reset()

		if bytesReaded < global.CHUNK_SIZE {

			break
		}

		chunkNo++

	}

	return true, nil

}

func VerifyChunk(chunk []byte, chunkno uint64, trackerFile *os.File, hasher hash.Hash) error {
	trackerFile.Seek(0, 0)
	trackerScanner := bufio.NewScanner(trackerFile)
	fmt.Println("chunko", chunkno)
	for i := 0; i < 2; i++ {
		if !trackerScanner.Scan() {
			if trackerScanner.Err() != nil {
				return trackerScanner.Err()
			}
			return fmt.Errorf("file has fewer than %d lines", i)
		} else {
			fmt.Println("skipping Line", trackerScanner.Text())
		}

	}

	for i := uint64(0); i < chunkno; i++ {
		if !trackerScanner.Scan() {
			if trackerScanner.Err() != nil {
				return trackerScanner.Err()
			}
			return fmt.Errorf("file has fewer than %d lines", i)
		} else {
			fmt.Println("skipping Line", trackerScanner.Text())
		}

	}
	if !trackerScanner.Scan() {
		if trackerScanner.Err() != nil {
			return trackerScanner.Err()
		}
		return fmt.Errorf("file has fewer  lines")
	} else {
		fmt.Println("current", trackerScanner.Text())
	}
	trackerChunkData := strings.Split((trackerScanner.Text()), ":")
	trackerchunkNo := trackerChunkData[0]
	chunkHashHex := trackerChunkData[2]
	fmt.Println(chunkHashHex, "chunkHashHex")

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
		fmt.Println("CALCHASH",calcHash,"HASHBYTES",hashBytes)
		return fmt.Errorf("chunks not match ing ")

	}

	return nil

}
