package cli

import (
	"bufio"
	"flag"
	"fmt"
	"interfiles/client"
	"interfiles/global"
	"interfiles/master"
	"os"
	"strings"
)


func StartCli() {

	fmt.Printf(
		"Welcome to %v! These are the available commands: \n",
		"InterFiles",
	)
	fmt.Println("help    - Show available commands")
	fmt.Println("exit    - Exit ")
	fmt.Println("client - Starts Client")
	fmt.Println("master - Starts Master")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		args := strings.Fields(input)

		command := args[0]

		switch command {
		case "hello":
			fmt.Println("Hello there!")
		case "client":
			// fmt.Println("starting server ")
			directory := handleDirectoryOperation(command, args[1:])
			clientObj, err := client.InitalizeClient(directory)
			if err != nil {
				global.ErrorPrint.Print(err)
				fmt.Println("")
				return
			}
			StartClientCli(clientObj)

		case "master":
			fmt.Println("starting master server ")
			master.InitalizeMaster()
			// fmt.Println(master)



		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  hello - Get a greeting")
			fmt.Println("client  - start client server")
			fmt.Println("master  - start master server")

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
		if len(input) == 0 {
			continue
		}
		args := strings.Fields(input)

		command := args[0]

		if len(command) == 0 {
			continue
		}

		if command == "exit" {
			break
		}

		switch command {
		case "upload":
			filepath := handleFileOperation(command, args[1:])
			if isValidFilePath(filepath) {
				err := c.AnnounceFile(filepath)
				if err != nil {
					global.ErrorPrint.Println(err)
					return
				}

			}

		case "download":
			filepath := handleFileOperation(command, args[1:])
			if isValidFilePath(filepath) {
				c.DownloadFile(filepath)

			}

		case "stat":
			filepath := handleFileOperation(command, args[1:])
			if isValidFilePath(filepath) {
				c.GetStats(filepath)

			}
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("upload -p path/to/file -  Upload a file")
			fmt.Println("download -p path/to/tracker-file -  Downlaod a file")
			fmt.Println("download -p path/to/tracker-file  - Get stats of file")

			fmt.Println("  help  - Show this help message")
			fmt.Println("  exit  - Exit the program")

		}

	}

}

func handleFileOperation(operation string, args []string) string {
	// Define flags
	flagSet := flag.NewFlagSet(operation, flag.ContinueOnError)
	filePath := flagSet.String("p", "", "File path for "+operation)

	err := flagSet.Parse(args)
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		return ""
	}

	if *filePath == "" {
		fmt.Println("Please specify a file path using the -p flag")
		return ""
	}

	return *filePath
}

func handleDirectoryOperation(operation string, args []string) string {
	// Define flags
	flagSet := flag.NewFlagSet(operation, flag.ContinueOnError)
	directoryPath := flagSet.String("d", "", "File path for "+operation)

	err := flagSet.Parse(args)
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		return ""
	}

	return *directoryPath
}

func isValidFilePath(path string) bool {
	if path == "" {
		return false
	}
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
