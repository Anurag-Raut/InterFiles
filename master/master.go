package master

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
)

var CLIENT_FILE_PATH="/home/anurag/projects/dfs/clients.txt"

func getRoot(w http.ResponseWriter, r *http.Request) {
	writer := bufio.NewWriter(w)
	writer.WriteString("hello")
	writer.Flush()


}

func addClient(w http.ResponseWriter, r *http.Request) {
	
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		clientData := string(body)
		fmt.Println(clientData,"client data")
		file, err := os.OpenFile(CLIENT_FILE_PATH, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

		if err != nil {
			if os.IsExist(err) {
				fmt.Println("File already exists..")
			} else {
	
				fmt.Println("Error opening file:", err)
				return
			}
		}

		file.WriteString(clientData)

		defer file.Close()

}



func StartMaster() {

	http.HandleFunc("/", getRoot)
	http.HandleFunc("/addClient", addClient)

	err := http.ListenAndServe(":8000", nil)

	if err != nil {
		fmt.Println("AN error occured")
		return
	}


}
