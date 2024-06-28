package global


var MASTER_SERVER_URL = "127.0.0.1:8000"

const (
	GET_FILE=uint8(iota)
	SENDER_HANDSHAKE
	CLIENT_HANDSHAKE
	PULL_FILE
	REQUEST_FILE
	ADD_CLIENT
	ANNOUNCE
	
	

)



type Client struct {


	ClientId string `json:"clientId"`
	Ip       string `json:"ip"`
	Port     string    `json:"port"`
	Directory string `json:"directory"`
}
func (c *Client) GetUrl() string {
    return c.Ip + ":" + c.Port
}

type File struct {
	ID string
	Filename string
	Clients []Client



}
var CHUNK_SIZE = 1024
var HEADER_LEN=2+8

//client
