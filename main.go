package main

import (
	"dfs/protocol"
	"sync"
	"time"
)

func main(){

	var wg sync.WaitGroup

	wg.Add(1)

	go func ()  {
		defer wg.Done()
		 protocol.StartServer()
	}()
	time.Sleep(5*time.Second)
	protocol.SendMessage("helllooo")
	
	wg.Wait()
}

