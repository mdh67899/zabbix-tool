package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/mdh67899/zabbix-tool/internal"
)

const (
	version string = "0.0.1"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	flagSet := flags()
	flagSet.Parse(os.Args[1:])

	if flagSet.Lookup("V").Value.String() == "true" {
		log.Println("version is:", version)
		return
	}

	if flagSet.Lookup("h").Value.String() == "true" {
		flagSet.Usage()
		return
	}

	hostname := flagSet.Lookup("s").Value.String()
	if hostname == "" {
		flagSet.Usage()
		return
	}

	key := flagSet.Lookup("k").Value.String()
	if key == "" {
		flagSet.Usage()
		return
	}

	value := flagSet.Lookup("o").Value.String()
	if value == "" {
		flagSet.Usage()
		return
	}

	item := internal.NewItem(hostname, key, value)
	senderReq := internal.NewSenderRequest([]internal.Item{item})
	data, err := internal.ReqEncoding(senderReq)
	if err != nil {
		log.Println("encoing sender request failed:", err)
		return
	}

	destination := flagSet.Lookup("z").Value.String() + ":" + flagSet.Lookup("p").Value.String()
	conn, err := internal.NewConn(destination)
	if err != nil {
		log.Println("create connection to zabbix server failed:", err)
		return
	}

	defer conn.Close()

	err = conn.Write(data)
	if err != nil {
		log.Println("write data to tcp connection failed:", err)
		return
	}

	resp_data, err := conn.ReadResp()
	if err != nil {
		log.Println("read response from zabbix agent failed:", err)
		return
	}

	var sendResp internal.SenderResp
	err = json.Unmarshal(bytes.Trim(resp_data, "\x00"), &sendResp)
	if err != nil {
		log.Println("decode response body to struct failed:", err)
		return
	}

	fmt.Printf("%s\n", sendResp.String())
}

func flags() *flag.FlagSet {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	flagSet.String("z", "127.0.0.1", "Specify host name or IP address of a host.")
	flagSet.Int("p", 10051, "Specify port number of trapper process of Zabbix server or proxy.. Default is 10051.")
	flagSet.String("s", "localhost", "Specify host name the item belongs to(as registered in Zabbix frontend). Host IP address and DNS name will not work.")
	flagSet.String("k", "", "Specify item key")
	flagSet.String("o", "", "Specify item value")
	flagSet.Bool("h", false, "Display this help and exit.")
	flagSet.Bool("V", false, "Output version information and exit.")

	return flagSet
}
