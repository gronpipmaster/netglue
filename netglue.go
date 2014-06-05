//Layer for net/rpc, server and client includes
package netglue

import (
	"log"
	"net"
	"net/rpc"
	"time"
)

//Enable verbose mode
var Verbose bool

//Buffer queue send requests
var BufferSize = 100

var (
	OutAddr string = ":8888"
	InAddr  string = ":8889"
	Network string = "tcp"
)

type msg struct {
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
}

type NetGlue struct {
	queue  chan *msg
	output chan interface{}
}

//Create new NetGlue server, listen to InAddr and run sender for OutAddr
func NewNetGlue(rcvr interface{}, output chan interface{}) (connect *NetGlue, err error) {
	connect = &NetGlue{make(chan *msg, BufferSize), output}
	//construct listen
	rpc.Register(rcvr)
	listen, err := net.Listen(Network, InAddr)
	if err != nil {
		return
	}
	if Verbose {
		log.Println("Listen success", Network, InAddr)
	}
	//run listener
	go connect.runListener(listen)

	client, err := rpc.Dial(Network, OutAddr)
	if err != nil {
		if Verbose {
			log.Println("Client connect err", err)
			log.Println("Sleep 5 second and try again.")
		}
		time.Sleep(5 * time.Second)
		client, err = rpc.Dial(Network, OutAddr)
		if err != nil {
			return nil, err
		}
	}
	if Verbose {
		log.Println("Connect success", Network, OutAddr)
	}
	//run sender
	go connect.runSender(client)

	return
}

func (p *NetGlue) Send(serviceMethod string, args interface{}, reply interface{}) {
	go func() {
		p.queue <- &msg{serviceMethod, args, reply}
	}()
}

func (p *NetGlue) runListener(listen net.Listener) {
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			if Verbose {
				log.Println("Accept output - ", err)
			}
			//TODO panic/fatal, or error for client this library see rpc.Accept http://golang.org/src/pkg/net/rpc/server.go?s=5911:5935#L595
			break
		} else {
			go rpc.ServeConn(conn)
		}
	}
}

func (p *NetGlue) runSender(client *rpc.Client) {
	for {
		select {
		//waiting msg from self queue
		case msg := <-p.queue:
			//connect to OutAddr
			client, err := rpc.Dial(Network, OutAddr)
			if err == nil {
				//send msg
				err := client.Call(msg.ServiceMethod, msg.Args, msg.Reply)
				if err != nil && Verbose {
					log.Println(err)
				}
				//send output chanel reply
				go func() {
					p.output <- msg.Reply
					client.Close()
				}()
			} else {
				//replay send
				if Verbose {
					log.Println("Dialing output - ", err)
					log.Println("Sleep 1 second and try again.")
				}
				time.Sleep(1 * time.Second)
				p.queue <- msg
			}
		}
	}
}
