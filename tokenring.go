package main

import (
    "fmt"
    "os"
    "flag"
    "net"
    "strconv"
    "./Msg"
    "time"
)

/* A Simple function to verify error */
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}

func main() {
    nodesCntPtr := flag.Int("n", 0, "nodes cnt")
    withholdingPtr := flag.Int("t", 0, "withholding time")
    flag.Parse()

    doneCh := make(chan int)
    for i := 1; i <= *nodesCntPtr; i++ {
        go start(i, *nodesCntPtr, *withholdingPtr, doneCh)
    }
    for i := 1; i <= *nodesCntPtr; i++ {
        <-doneCh
    }
}

const msgPort = 30000
const servicePort = 40000
var emptyMsg = Msg.Msg{Type: "", Dst: 0, Data: "", Src: 0, Ack: false}
func start(nodeID, nodesCnt, withholding int, doneCh chan int) {
    // Here makes message conneciton
    myMsgPort := ":" + strconv.Itoa(msgPort + nodeID)

    msgAddr, err := net.ResolveUDPAddr("udp", myMsgPort)
    CheckError(err)

    msgConn, err := net.ListenUDP("udp", msgAddr)
    CheckError(err)
    defer msgConn.Close()
    // And start listening
    messageCh := make(chan Msg.Msg)
    go listen(nodeID, messageCh, msgConn)

    // here makes sevice connection
    myServicePort := ":" + strconv.Itoa(servicePort + nodeID)

    serviceAddr, err := net.ResolveUDPAddr("udp", myServicePort)
    CheckError(err)

    serviceConn, err := net.ListenUDP("udp", serviceAddr)
    CheckError(err)
    defer serviceConn.Close()
    // And start listening
    serviceCh := make(chan Msg.Msg)
    go listen(nodeID, serviceCh, serviceConn)

    // Here make connection with right neighbour to send him messages
    rightID := 1
    if nodeID == nodesCnt {
        rightID = 1
    } else {
        rightID = nodeID+1
    }
    rihgtNeighbourPort := strconv.Itoa(msgPort + rightID)

    rightAddr,err := net.ResolveUDPAddr("udp","127.0.0.1:"+rihgtNeighbourPort)
    CheckError(err)

    localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    CheckError(err)

    connWithRight, err := net.DialUDP("udp", localAddr, rightAddr)
    CheckError(err)
    defer connWithRight.Close()

    leftID := nodesCnt
    if nodeID == 1 {
        leftID = nodesCnt
    } else {
        leftID = nodeID-1
    }
    timer := time.NewTimer(time.Second * 0)
    wantedMessage := emptyMsg
    gulp := false
    for {
        select {
        case msg := <-messageCh:
            switch{
            case msg.Empty() && wantedMessage.Empty() && !msg.Ack:
                // node 2: recieved token from node 1, sending token to node
                fmt.Println("node", nodeID, ": received token from node",
                    leftID, ", sending token to node", rightID)
                sendMsg(msg, connWithRight, withholding, gulp)

            case msg.Empty() && !wantedMessage.Empty() && !msg.Ack:
                fmt.Println("node", nodeID, ": received token from node",
                    leftID, ", sending token to node", rightID)
                sendMsg(wantedMessage, connWithRight, withholding, gulp)

            case !msg.Empty() && msg.Dst != nodeID && !msg.Ack:
                fmt.Println("node", nodeID, ": received token from node",
                    leftID, "with data from node", msg.Src, "data=('",
                    msg.Data, "')", ", sending token to node", rightID)
                sendMsg(msg, connWithRight, withholding, gulp)

            case !msg.Empty() && msg.Dst == nodeID && !msg.Ack:
                fmt.Println("node", nodeID, ": received token from node",
                    leftID, "with data from node", msg.Src, "data=('",
                    msg.Data, "')", ", sending token to node", rightID)
                ackMsg := Msg.Msg{Type: "", Dst: msg.Src, Data: "", Src: nodeID, Ack: true}
                sendMsg(ackMsg, connWithRight, withholding, gulp)

            case msg.Ack && msg.Dst != nodeID:
                fmt.Println("node", nodeID, ": received token from node",
                    leftID, "with delivery confirmation form node", msg.Src,
                    ", sending token to node", rightID)
                sendMsg(msg, connWithRight, withholding, gulp)

            case msg.Ack && msg.Dst == nodeID:
                fmt.Println("node", nodeID, ": received token from node",
                    leftID, "with delivery confirmation form node", msg.Src,
                    ", sending token to node", rightID)
                wantedMessage = emptyMsg
                sendMsg(emptyMsg, connWithRight, withholding, gulp)
            }
            timer = time.NewTimer(time.Second * time.Duration(nodesCnt) + 1)
            gulp = false

        case service := <-serviceCh:
            fmt.Println("node", nodeID, ": recieved service message:",
                service.ToJson())
            switch service.Type {
            case "send":
                wantedMessage = Msg.Msg{Type: "", Dst: service.Dst,
                    Data: service.Data, Src: nodeID, Ack: false}
            case "drop":
                gulp = true
            }

        case <-timer.C:
            if nodeID == nodesCnt {
                fmt.Println("Drop detected!")
                sendMsg(emptyMsg, connWithRight, 0, false)
            }
            timer = time.NewTimer(time.Second * time.Duration(nodesCnt) + 1)
        }
    }
}

func listen(nodeID int, ch chan Msg.Msg, conn *net.UDPConn) {
    buf := make([]byte, 1024)
    for {
        n,_,err := conn.ReadFromUDP(buf)

        if err != nil {
            fmt.Println("Error: ",err)
        }

        ch <- Msg.New(buf[0:n])
    }
}

func sendMsg(msg Msg.Msg, conn *net.UDPConn, withholding int, gulp bool) {
    if (gulp) {
        return
    }
    time.Sleep(time.Second * time.Duration(withholding))
    jsonMsg := msg.ToJson()
    buf := []byte(jsonMsg)
    _,err := conn.Write(buf)
    if err != nil {
        fmt.Println(jsonMsg, err)
    }
}
