package cli

import (
	"bufio"
	"dfs/protocol"
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
		case "start":
			fmt.Println("starting server ")
			protocol.StartServer()
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
