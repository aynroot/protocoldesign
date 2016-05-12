package main

import (
	"fmt"
	"net"
	"os"
	//"flag"
	"strconv"
	"time"
	pft "./pft/"
)

func main() {

        //portArg := flag.Int("t", 6222, "port to contact/listen on")
        //serverModeArg := flag.Bool("s", false, "start in server mode")
        //fileArg := flag.String("f", "", "file to be downloaded")
        //flag.Parse()
        //
        //if *serverModeArg && *fileArg != "" {
        //fmt.Println("can only download file in client mode")
        //return
        //}
        //
        //resource := "file-list"
        //if *fileArg != "" {
        //resource = "file:" + *fileArg
        //}
        //
        //fmt.Println("port:", *portArg)
        //fmt.Println("server mode:", *serverModeArg)
        //fmt.Println("server:", flag.Args())
        //
        //
        //if *serverModeArg {
        //Server(*portArg)
        //} else {
        //if len(flag.Args()) != 1 {
        //    fmt.Println("need to supply exactly one target server in client mode")
        //    return
        //}
        //Client(*portArg, flag.Args()[0], resource)
        //}

    // REQ
    _,req := pft.EncodeReq("file:/tmp/uno.txt")
    fmt.Println("Sent REQ:", req)

    _, received_req, pt := pft.DecodeReq(req)
    fmt.Println("Received REQ:", received_req, pt)
    //

    // PUSH
    _,push := pft.EncodePush("file:/tmp/dos.txt")
    fmt.Println("Sent PUSH:", push)

    _, received_push, pt := pft.DecodePush(push)
    fmt.Println("Received PUSH:", received_push, pt)
    //

    //NACK
    _,nack := pft.EncodeNack()
    fmt.Println("Sent NACK:", nack)

    _, received_nack := pft.DecodeNack(nack)
    fmt.Println("Received REQ:", received_nack)
    //

    //PUSH-ACK
    _,push_ack := pft.EncodePushAck()
    fmt.Println("Sent PUSH-ACK:", push_ack)

    _, received_push_ack := pft.DecodePushAck(push_ack)
    fmt.Println("Received PUSH-ACK:", received_push_ack)

    //RST
    _,rst := pft.EncodeRst()
    fmt.Println("Sent RST:", rst)

    _, received_rst := pft.DecodeRst(rst)
    fmt.Println("Received RST:", received_rst)

    //GET
    _,get := pft.EncodeGet([4]byte{20,23,129,20})
    fmt.Println("Sent GET:", get)

    _, received_get, pt := pft.DecodeGet(get)
    fmt.Println("Received GET:", received_get, pt)

    //DATA
    _,data := pft.EncodeData([4]byte{2},[491]byte{100,100})
    fmt.Println("Sent DATA:", data)

    _, received_data, block, pt := pft.DecodeData(data)
    fmt.Println("Received DATA:", received_data, block,  pt)

    //REQ-ACK
    _,req_ack := pft.EncodeReqAck([8]byte{3,4,5,6})
    fmt.Println("Sent REQ-ACK:", req_ack)

    _, received_req_ack, resource_size, pt := pft.DecodeReqAck(req_ack)
    fmt.Println("Received REQ-ACK:", received_req_ack, resource_size,  pt)


}


func CheckError(err error) {
    if err != nil {
        fmt.Println("Error: ", err)
        os.Exit(0)
    }
}


func Server(port int) {
    addr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(port))
    CheckError(err)

    fmt.Println("listening on", addr)

    /* Now listen at selected port */
    conn, err := net.ListenUDP("udp", addr)
    CheckError(err)
    defer conn.Close()

    buf := make([]byte, 512)
    for {
        size, sender, err := conn.ReadFromUDP(buf)
        CheckError(err)
        fmt.Println("Received ", string(buf[0:size]), " from ", sender)
    }
    fmt.Println("closing server")
}


func Client(port int, server string, resource string) {
    server_addr, err := net.ResolveUDPAddr("udp", server + ":" + strconv.Itoa(port))
    CheckError(err)

    local_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    CheckError(err)

    conn, err := net.DialUDP("udp", local_addr, server_addr)
    CheckError(err)
    defer conn.Close()

    i := 1
    for {
        conn.Write([]byte(strconv.Itoa(i)))
        i += 1
        time.Sleep(time.Second)
    }

}

