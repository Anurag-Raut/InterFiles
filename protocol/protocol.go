	package protocol

	import (
		"bufio"
		"encoding/binary"
		"fmt"
		"net"
	)



	func StartServer(){
		fmt.Println("starting server on port 8080")
		listener, err := net.Listen("tcp", ":8080")
		if err != nil {
			fmt.Println("Error staring server",err.Error())
			return 
		}

		defer listener.Close()
		fmt.Println("Listening on port 8080")




		for {
			conn,err:=listener.Accept()
			
			if err != nil {
				fmt.Println("Error accepting connections",err.Error())

				return 
			}


			go handleConnection(conn)


			
			






		}


		



	}

	func handleConnection(conn net.Conn ){
		defer conn.Close()

		reader:= bufio.NewReader(conn)
		var msgLen int32

		err := binary.Read(reader, binary.BigEndian, &msgLen)
		if err != nil {
			fmt.Println("error reading message len:",err.Error())
			return
		}

		buf := make([]byte, msgLen)

		_, err = reader.Read(buf)
		if err != nil {
			fmt.Println("error reading message:",err.Error())

			return 

		}


		fmt.Println(string(buf))




	}


	func SendMessage(message string) {
		// Connect to the server
		conn, err := net.Dial("tcp", ":8080")
		if err != nil {
			fmt.Println("Failed to connect to the server:", err)
			return
		}
		defer conn.Close()
	
		// Create a buffered writer
		writer := bufio.NewWriter(conn)
		defer writer.Flush()
	
		// Write the message length
		msgLen := int32(len(message))
		err = binary.Write(writer, binary.BigEndian, msgLen)
		if err != nil {	
			fmt.Println("Failed to write message length:", err)
			return
		}
	
		// Write the message payload
		_, err = writer.WriteString(message)
		if err != nil {	
			fmt.Println("Failed to write message:", err)
			return
		}
	
		fmt.Println("Message sent successfully")
	}
	
		