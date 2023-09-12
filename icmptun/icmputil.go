package icmptun

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"strconv"
	"sync"
)

var HeadFlag = []byte{0, 8, 0, 2}

type ICMPEnv struct {
	RecvTun    chan *MyMsg
	SendTun    chan *MyMsg
	Listen     string
	TargetIp   string
	TargetPort int
	Server     string
	ConnList   sync.Map
	ICMPListen net.PacketConn
}

type ICMPPackage struct {
	data       *icmp.Message
	srcAddr    net.Addr
	targetAddr net.Addr
}

// icmp.echo.body=HeadFlag+MyMsg
type MyMsg struct {
	Id    string
	Ip    string
	Port  int
	Data  []byte
	SrcIp string
}

func ParseMessage(data []byte) (*MyMsg, bool) {
	msg, err := icmp.ParseMessage(1, data)
	msgData := &MyMsg{}
	if err != nil {
		panic(err)
	}
	//解析MyMsg
	//判断是否为隧道通信
	echoBody := msg.Body
	echoBodyRaw, err := echoBody.Marshal(ipv4.ICMPTypeEcho.Protocol())
	head := echoBodyRaw[4:8]
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(HeadFlag); i++ {
		if head[i] != HeadFlag[i] {
			return nil, false
		}
	}

	//长度
	fmt.Println("LEN:" + strconv.Itoa(len(data)))
	//序列化传输
	decodeData := bytes.Buffer{}
	decodeData.Write(echoBodyRaw[8:])
	dec := gob.NewDecoder(&decodeData)
	err = dec.Decode(msgData)
	if err != nil {
		panic(err)
	}
	return msgData, true
}

func RunListern(env *ICMPEnv) {
	for {
		conn, err := icmp.ListenPacket("ip4:icmp", "")
		env.ICMPListen = conn
		if err != nil {
			panic(err)
		}
		rb := make([]byte, 20480)
		_, target, err := conn.ReadFrom(rb)
		if err != nil {
			panic(err)
		}
		fmt.Printf("收到icmp请求，来自：%s\n", target)
		if err != nil {
			panic(err)
		}

		//解析
		msgData, r := ParseMessage(rb)
		msgData.SrcIp = target.String()
		if r {
			env.RecvTun <- msgData
			fmt.Printf("icmp消息解析成功，发送到env.RecvTun\n")
		} else {
			continue
		}
	}
}
