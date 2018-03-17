package internal

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

const (
	Sender_request string = "agent data"
	ActiveCheck    string = "active checks"
)

type SenderRequest struct {
	Request string `json:"request"`
	Data    []Item `json:"data"`
	Clock   int64  `json:"clock"`
}
type Item struct {
	Host  string      `json:"host"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Clock int64       `json:"clock"`
}

func NewSenderRequest(data []Item, clock int64) SenderRequest {
	return SenderRequest{
		Request: Sender_request,
		Data:    data,
		Clock:   clock,
	}
}

func NewItem(host, key string, value interface{}, clock int64) Item {
	return Item{
		host,
		key,
		value,
		clock,
	}
}

type SenderResp struct {
	Response string `json:"response"`
	Info     string `json:"info"`
}

func (data SenderResp) String() string {
	return "[response: " + data.Response + ", info: " + data.Info
}

//    "response":"success",
//  "info":"Processed 2 Failed 0 Total 2 Seconds spent 0.002070"

func ReqEncoding(data interface{}) ([]byte, error) {
	buffer := bytes.Buffer{}

	_, err := buffer.Write(ZabbixHeader)
	if err != nil {
		return []byte{}, err
	}

	body, err := json.Marshal(data)
	if err != nil {
		return []byte{}, err
	}

	length := len(body)
	dataLen := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataLen, uint64(length))

	_, err = buffer.Write(dataLen)
	if err != nil {
		return []byte{}, err
	}

	_, err = buffer.Write(body)
	if err != nil {
		return []byte{}, err
	}

	return buffer.Bytes(), nil
}

func DecodeSenderResp(data []byte) (SenderResp, error) {
	var sendResp SenderResp
	err := json.Unmarshal(data, &sendResp)
	return sendResp, err
}

/*
Agent Sends:
<HEADER><DATALEN>{
   "request":"agent data",
   "data":[
       {
           "host":"<hostname>",
           "key":"log[\/home\/zabbix\/logs\/zabbix_agentd.log]",
           "value":" 13039:20090907:184546.759 zabbix_agentd started. ZABBIX 1.6.6 (revision {7836}).",
           "lastlogsize":80,
           "clock":1252926015
       },
       {
           "host":"<hostname>",
           "key":"agent.version",
           "value":"1.6.6",
           "clock":1252926015
       }
   ],
   "clock":1252926016
}


Server Response:

<HEADER><DATALEN>{
    "response":"success",
    "info":"Processed 2 Failed 0 Total 2 Seconds spent 0.002070"
}

*/

/*
Get list of items

Agent Request:

<HEADER><DATALEN>{
   "request":"active checks",
   "host":"<hostname>"
}

Server Response：

{
    "response":"success",
    "data":[
    {
        "key":"log[\/home\/zabbix\/logs\/zabbix_agentd.log]",
        "delay":"30",
        "lastlogsize":"0"
    },
    {
        "key":"agent.version",
        "delay":"600"
    }
    ]
}

*/

type ItemsQuery struct {
	Request string `json:"request"`
	Host    string `json:"host"`
}

type QueryResp struct {
	Response string     `json:"response"`
	Data     []ItemMeta `json:"data"`
}

type ItemMeta struct {
	Key   string `json:"key"`
	Delay string `json:"deelay"`
}

func NewItemsQuery(hostname string) ItemsQuery {
	return ItemsQuery{
		Request: ActiveCheck,
		Host:    hostname,
	}
}
