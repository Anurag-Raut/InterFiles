package tracker

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nrednav/cuid2"
)


func CreateTracker(ogFile os.File,noOfBytes string,clientId string){
	id:=cuid2.Generate()
	filename:=filepath.Base(ogFile.Name())
	file,err:=os.Create(fmt.Sprintf("%s_tracker.txt",id))
	if err!=nil {
		fmt.Println("Error creatinf file",err.Error())
		return
	}

	date:=time.Now().String()


	file.WriteString(id)
	file.WriteString(fmt.Sprintf("%s:%s:%s:%s",filename,date,noOfBytes,clientId))
	file.WriteString()

	












}

//uploaderId
//nameid
// filename
