package main

import (
	"encoding/json"
	"flag"
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

	req := internal.NewItemsQuery(hostname)
	data, err := internal.ReqEncoding(req)
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

	var Resp internal.QueryResp
	err1 := json.Unmarshal(resp_data, &Resp)
	log.Println(string(resp_data[:]))

	if err1 != nil {
		log.Println("decode response body to struct failed:", err1)
		return
	}

	fmt.Printf("%v\n", Resp)
}

func flags() *flag.FlagSet {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	flagSet.String("z", "localhost", "zabbix server ip.")
	flagSet.Int("p", 10051, "zabbix server port.")
	flagSet.String("s", "", "zabbix host hostname.")
	flagSet.Bool("h", false, "Display this help and exit.")
	flagSet.Bool("V", false, "Output version information and exit.")

	return flagSet
}
