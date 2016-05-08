
package main

import "flag"
import "fmt"

func main() {
	portArg := flag.Int("t", 6222, "port to contact/listen on")
	serverModeArg := flag.Bool("s", false, "start in server mode")
	flag.Parse()

	fmt.Println("port:", *portArg)
	fmt.Println("server mode:", *serverModeArg)
	fmt.Println("server:", flag.Args())
}
