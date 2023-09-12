package icmptun

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"strconv"
)

func RunServer(env *ICMPEnv) {
	fmt.Println("start Lister ICMP")
	go RunListern(env)
	TcpServer(env)
}

// 根据icmpdata中的信息，转发成tcp流量
func TcpServer(env *ICMPEnv) {
	//从管道中发送icmp请求到客户端
	go func() {
		for {
			msg := <-env.SendTun
			fmt.Println("发送返回消息到客户端")
			sendICMPResp(msg, env)
		}
	}()
	//发送，异步接受tcp请求,
	for {
		msg := <-env.RecvTun
		fmt.Println("发送tcp请求")
		fmt.Println(msg.Id)
		fmt.Printf("target:%s:%d\n", msg.Ip, msg.Port)
		tcpConn, err := net.Dial("tcp", msg.Ip+":"+strconv.Itoa(msg.Port))
		if err != nil {
			panic(err)
		}
		_, err = tcpConn.Write(msg.Data)
		if err != nil {
			panic(err)
		}
		fmt.Println("TCP请求发送成功")
		go recvTCP(tcpConn, env, msg)
	}
}
func sendICMPResp(mymsg *MyMsg, env *ICMPEnv) {
	//defer conn.Close()
	echo := &icmp.Echo{
		ID:   1,
		Seq:  1,
		Data: []byte{0, 8, 0, 2},
	}

	b := bytes.Buffer{}
	enc := gob.NewEncoder(&b)
	enc.Encode(mymsg)
	echo.Data = append(echo.Data, b.Bytes()...)
	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEchoReply,
		Code: 0,
		Body: echo,
	}
	data, err := msg.Marshal(nil)
	if err != nil {
		panic(err)
	}
	target, err := net.ResolveIPAddr("ip4", mymsg.SrcIp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("发送返回icmp消息，目标%s,消息长度%d\n", target, len(data))
	env.ICMPListen.WriteTo(data, target)
}
func recvTCP(conn net.Conn, env *ICMPEnv, msg *MyMsg) {
	buf := make([]byte, 20480)
	_, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("收到tcp响应,id:%s，发动到env.SendTun\n", msg.Id)
	env.SendTun <- &MyMsg{
		Id:    msg.Id,
		Ip:    msg.Ip,
		Port:  msg.Port,
		Data:  buf,
		SrcIp: msg.SrcIp,
	}
}
