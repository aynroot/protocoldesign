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
    req := pft.EncodeReq("file:/tmp/uno.txt")
    fmt.Println("Sent REQ:", req)
    fmt.Println("REQ HASH ok:", pft.VerifyPacket(req))
    _, resource := pft.DecodeReq(req)
    fmt.Println("Received REQ:", resource)

    // PUSH
    push := pft.EncodePush("file:/tmp/dos.txt")
    fmt.Println("Sent PUSH:", push)

    //NACK
    nack := pft.EncodeNack()
    fmt.Println("Sent NACK:", nack)
    fmt.Println("NACK ok:", len(nack) == 17 && nack[16] == 3 && pft.VerifyPacket(nack))


    //PUSH-ACK
    push_ack := pft.EncodePushAck()
    fmt.Println("Sent PUSH-ACK:", push_ack)

    //RST
    rst := pft.EncodeRst()
    fmt.Println("Sent RST:", rst)

    //GET
    get := pft.EncodeGet(123456)
    fmt.Println("Sent GET:", get)

    _,index := pft.DecodeGet(get)
    fmt.Println("Received GET:", index)

    //DATA
    data := pft.EncodeData(123456789, []byte{9, 8, 7, 6, 5, 4, 3, 2, 1})
    fmt.Println("Sent DATA:", data)

    fmt.Println("DATA hash ok: ", pft.VerifyPacket(data))

    _, block, data := pft.DecodeData(data)
    fmt.Println("Received DATA:", block,  data)

    //REQ-ACK
    req_ack := pft.EncodeReqAck(40000000001,
        []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16})
    fmt.Println("Sent REQ-ACK:", req_ack)

    err, size, hash := pft.DecodeReqAck(req_ack)
    if err != nil {
        fmt.Println("error in req ack:", err)
    } else {
        fmt.Println("received req ack with size", size, "and hash", hash)
    }
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

