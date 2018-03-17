package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

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

	timestamp := time.Now().Unix()
	item := internal.NewItem(hostname, key, value, timestamp)
	senderReq := internal.NewSenderRequest([]internal.Item{item}, timestamp)
	data, err := senderReq.Encoding()
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
	err1 := json.Unmarshal(resp_data[0:], &sendResp)
	log.Println(string(resp_data[0:]))
        log.Println(err1)

	if err1 != nil {
		log.Println("decode response body to struct failed:", err1)
		return
	}

	//fmt.Printf("%s\n", sender_resp.String())
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
	/*
	   -i --input-file input-file   Load values from input file. Specify - for
	                              standard input. Each line of file contains
	                              whitespace delimited: <host> <key> <value>.
	                              Specify - in <host> to use hostname from
	                              configuration file or --host argument

	   -T --with-timestamps       Each line of file contains whitespace delimited:
	                              <host> <key> <timestamp> <value>. This can be used
	                              with --input-file option. Timestamp should be
	                              specified in Unix timestamp format

	   -r --real-time             Send metrics one by one as soon as they are
	                              received. This can be used when reading from
	                              standard input
	*/
}
