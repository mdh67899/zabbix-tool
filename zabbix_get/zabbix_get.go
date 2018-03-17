package main

import (
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

	key := flagSet.Lookup("k").Value.String()
	if key == "" {
		flagSet.Usage()
		return
	}

	destination := flagSet.Lookup("s").Value.String() + ":" + flagSet.Lookup("p").Value.String()

	conn, err := internal.NewConn(destination)
	if err != nil {
		log.Println("create connection failed:", err)
		return
	}

	defer conn.Close()

	err = conn.Write([]byte(key))
	if err != nil {
		log.Println("write data to tcp connection failed:", err)
		return
	}

	data, err := conn.ReadAgentPassiveCheck()
	if err != nil {
		log.Println("read response from zabbix agent failed:", err)
		return
	}

	fmt.Printf("%s\n", string(data[:]))
}

func flags() *flag.FlagSet {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	flagSet.String("s", "127.0.0.1", "Specify host name or IP address of a host.")
	flagSet.Int("p", 10050, "Specify port number of agent running on the host. Default is 10050.")
	//flagSet.String("I", "", "Specify source IP address.")
	flagSet.String("k", "", "Specify key of item to retrieve value for.")
	flagSet.Bool("h", false, "Display this help and exit.")
	flagSet.Bool("V", false, "Output version information and exit.")

	return flagSet
}

func printUsage() {
	log.Println("usage: zabbix_get [-hV] -s <host name or IP> [-p <port>] [-I <IP address>] -k <key>")
	log.Println("Options:")
	log.Println(" -s --host <host name or IP>          Specify host name or IP address of a host")
	log.Println(" -p --port <port number>              Specify port number of agent running on the host. Default is 10050")
	//log.Println(" -I --source-address <IP address>     Specify source IP address")
	log.Println("-k --key <key of metric>             Specify key of item to retrieve value for")
	log.Println("-h --help                            Give this help")
	log.Println(" -V --version                         Display version number")
	log.Println("Example: zabbix_get -s 127.0.0.1 -p 10050 -k 'system.cpu.load[all,avg1]'")
}
