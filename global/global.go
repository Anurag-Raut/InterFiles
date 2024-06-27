package global


var MASTER_SERVER_URL = "http://127.0.0.1:8000"
type Client struct {
	ClientId string `json:"clientId"`
	Ip       string `json:"ip"`
	Port     string    `json:"port"`
}

var CHUNK_SIZE = 1024
var HEADER_LEN=2+8