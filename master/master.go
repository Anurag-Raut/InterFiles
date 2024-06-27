package master

import (
	"bufio"
	"dfs/global"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

func getClients(w http.ResponseWriter, r *http.Request) {
	file, err := os.OpenFile(CLIENT_FILE_PATH, os.O_RDONLY,0)

	if err != nil {
		fmt.Println("ERROR",err.Error())
		return
	}
	scanner := bufio.NewScanner(file)
	

	var clients []global.Client
	for i := 0; i < 3 && scanner.Scan(); i++ {
		line := scanner.Text()
		args:=strings.Split(line,":")
		if err != nil {
			fmt.Printf("Invalid port number: %s\n", args[2])
			continue
		}

		newClinet:=global.Client{
			ClientId: args[0],
			Ip: args[1],
			Port: args[2],

		}

		fmt.Println("client id :",newClinet.ClientId,"clientIP :",newClinet.Ip,"client port :",newClinet.Port)

		clients = append(clients, newClinet)
	}

	jsonData, err := json.Marshal(clients)
	if err != nil {
		http.Error(w, "Error marshaling JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)


	

	



	

	

}	



func StartMaster() {

	http.HandleFunc("/", getRoot)
	http.HandleFunc("/addClient", addClient)
	http.HandleFunc("/getClients",getClients)

	err := http.ListenAndServe(":8000", nil)

	if err != nil {
		fmt.Println("AN error occured")
		return
	}


}
