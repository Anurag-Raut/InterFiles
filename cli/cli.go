package cli

import (
	"bufio"
	"dfs/client"
	"dfs/master"
	"flag"
	"fmt"
	"os"
	"strings"
)

func StartCli() {

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}
		switch input {
		case "hello":
			fmt.Println("Hello there!")
		case "start client":
			fmt.Println("starting server ")
			clientObj := client.InitalizeClient()
			StartClientCli(clientObj)

		case "send file":

		case "start master":
			fmt.Println("starting master server ")
			master.StartMaster()

		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  hello - Get a greeting")
			fmt.Println("  help  - Show this help message")
			fmt.Println("  exit  - Exit the program")
		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}

	}

}

func StartClientCli(c client.ClientService) {
	reader := bufio.NewReader(os.Stdin)
	for {

		fmt.Print("CLIENT > ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		args := strings.Fields(input)

		command := args[0]

		if len(command) == 0 {
			return
		}
	

		if command == "exit" {
			break
		}

		switch command {
		case "uploadFile":
			filepath:=handleFileOperation(command,args[1:])
			if isValidFilePath(filepath){
				c.UploadFile(filepath)
					
			} else{
				fmt.Println("invalid path")
			}

		}

	}

}

func handleFileOperation(operation string, args []string) string {
	// Define flags
	flagSet := flag.NewFlagSet(operation, flag.ContinueOnError)
	filePath := flagSet.String("f", "", "File path for "+operation)

	err := flagSet.Parse(args)
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		return ""
	}

	if *filePath == "" {
		fmt.Println("Please specify a file path using the -f flag")
		return ""
	}

	return *filePath
}


func isValidFilePath(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Error: File does not exist.")
		} else if os.IsPermission(err) {
			fmt.Println("Error: Permission denied to access the file.")
		} else {
			fmt.Println("Error:", err)
		}
		return false
	}
	return true
}


