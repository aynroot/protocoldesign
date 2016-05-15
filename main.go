package main

import (
	"fmt"
	"flag"
	"os"
)

const UDP_BUFFER_SIZE = 512

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func main() {
	portArg := flag.Int("t", 6222, "port to contact/listen on")
	serverModeArg := flag.Bool("s", false, "start in server mode")
	fileArg := flag.String("f", "", "file to be downloaded")
	flag.Parse()

	if *serverModeArg && *fileArg != "" {
		fmt.Println("can only download file in client mode")
		return
	}

	resource := "file-list"
	if *fileArg != "" {
		resource = "file:" + *fileArg
	}

	fmt.Println("port:", *portArg)
	fmt.Println("server mode:", *serverModeArg)
	fmt.Println("server:", flag.Args())

	if *serverModeArg {
		Server(*portArg)
	} else {
		if len(flag.Args()) != 1 {
			fmt.Println("need to supply exactly one target server in client mode")
			return
		}
		Client(*portArg, flag.Args()[0], resource)
	}
}
