package icmptun

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
)

func RunClient(env *ICMPEnv) {
	go RunTcpServer(env)
}

func RunTcpServer(env *ICMPEnv) {
	tcpServer, err := net.Listen("tcp4", ":"+env.Listen)
	if err != nil {
		panic(err)
	}
	//从管道中接受消息，作为icmp发送出去
	conn, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		panic(err)
	}
	env.ICMPListen = conn
	go func() {
		for {
			data := <-env.SendTun
			sendICMP(*data, env)
		}
	}()
	//接受ICMP响应
	go func() {
		for {
			recvICMP(env)
		}
	}()
	//返回tcp连接
	go func() {
		for {
			mymsg := <-env.RecvTun
			//resp := make([]byte, 40960)
			conn, result := env.ConnList.LoadAndDelete(mymsg.Id)
			if result {
				fmt.Printf("收到tcp返回消息，消息id%d\n", mymsg.Id)
				_, err := conn.(net.Conn).Write(mymsg.Data)
				if err != nil {
					panic(err)
				}
				conn.(net.Conn).Close()
			}
		}

	}()
	for {
		//监听本地端口，将流量转发为icmp协议,每个连接单独起一个进程处理
		buf := [1024]byte{}
		conn, err := tcpServer.Accept()
		defer conn.Close()
		//接受本地tcp请求，发送给sendICMP进程
		go func() {
			if err != nil {
				panic(err)
			}
			conn.Read(buf[:])
			connUUID, err := uuid.NewUUID()
			if err != nil {
				panic(err)
			}
			//通过uuid将icmp请求和tcp请求关联起来Z
			env.ConnList.Store(connUUID.String(), conn)
			env.SendTun <- &MyMsg{
				Id:   connUUID.String(),
				Ip:   env.TargetIp,
				Port: env.TargetPort,
				Data: buf[:],
			}
			fmt.Printf("收到tcp请求,消息id%s\n", connUUID)
		}()
	}

}
func sendICMP(mymsg MyMsg, env *ICMPEnv) {
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
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: echo,
	}
	data, err := msg.Marshal(nil)
	if err != nil {
		panic(err)
	}
	target, err := net.ResolveIPAddr("ip4", env.Server)
	if err != nil {
		panic(err)
	}
	//长度断点
	fmt.Printf("发送icmp消息，长度%d\n", len(data))
	env.ICMPListen.WriteTo(data, target)
}

func recvICMP(env *ICMPEnv) {
	buf := make([]byte, 40960)
	env.ICMPListen.ReadFrom(buf)
	fmt.Printf("收到icmp请求\n")
	data, result := ParseMessage(buf)
	if result {
		env.RecvTun <- data
		fmt.Printf("解析成功，消息id：%s\n", data.Id)
	}
}
