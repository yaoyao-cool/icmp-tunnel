package main

import (
	"flag"
	"fmt"
	"icmptun/icmptun"
	"sync"
)

var wg sync.WaitGroup

var t string
var server string
var target string
var targetPort int
var listen string

func init() {
	flag.StringVar(&t, "t", "c", "'-t c' or '-t s'")
	flag.StringVar(&server, "sip", "", "if user '-t c',must give a ip,'-ip 192.168.1.1'")
	flag.StringVar(&target, "tip", "", "-tg 1.1.1.1")
	flag.IntVar(&targetPort, "tport", 0, "-tport 80")
	flag.StringVar(&listen, "l", "", "-l 8080")
	flag.Parse()
}

func main() {
	fmt.Println(t)
	if t != "s" && t != "c" {
		fmt.Println("param error")
		return
	}
	env := icmptun.ICMPEnv{
		RecvTun:    make(chan *icmptun.MyMsg, 100),
		SendTun:    make(chan *icmptun.MyMsg, 100),
		Listen:     listen,
		TargetIp:   target,
		TargetPort: targetPort,
		Server:     server,
	}
	if t == "c" {
		if target == "" || listen == "" || targetPort == 0 || server == "" {
			fmt.Println("param error")
			return
		}
		wg.Add(1)
		icmptun.RunClient(&env)
	} else if t == "s" {
		wg.Add(1)
		icmptun.RunServer(&env)
	}
	wg.Wait()
}
