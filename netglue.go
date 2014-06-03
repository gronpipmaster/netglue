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
func NewNetGlue(rcvr interface{}, output chan interface{}) (*NetGlue, error) {
	p := &NetGlue{make(chan *msg, BufferSize), output}
	//construct listen
	rpc.Register(rcvr)
	listen, err := net.Listen(Network, InAddr)
	if err != nil {
		return nil, err
	}
	if Verbose {
		log.Println("Listen success", Network, InAddr)
	}
	//run listener
	go p.runListener(listen)
	//run sender
	go p.runSender()
	return p, nil
}

func (p *NetGlue) Send(serviceMethod string, args interface{}, reply interface{}) {
	go func() {
		p.queue <- &msg{serviceMethod, args, reply}
	}()
}

func (p *NetGlue) runListener(listen net.Listener) {
	for {
		conn, err := listen.Accept()
		if err != nil {
			if Verbose {
				log.Println(err)
			}
			//TODO panic/fatal, or error for client this library see rpc.Accept http://golang.org/src/pkg/net/rpc/server.go?s=5911:5935#L595
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func (p *NetGlue) runSender() {
	for {
		select {
		//waiting msg from self queue
		case msg := <-p.queue:
			//connect to OutAddr
			client, err := rpc.Dial(Network, OutAddr)
			if err == nil {
				//send async msg
				divCall := client.Go(msg.ServiceMethod, msg.Args, msg.Reply, nil)
				replyMsg := <-divCall.Done // will be equal to divCall
				if replyMsg.Error != nil && Verbose {
					log.Println(replyMsg.Error)
				}
				//send output chanel reply
				go func() {
					p.output <- replyMsg.Reply
				}()
			} else {
				//replay send
				if Verbose {
					log.Println("Dialing output", err)
					log.Println("Sleep 1 second and try again.")
				}
				time.Sleep(1 * time.Second)
				p.queue <- msg
			}
		}
	}
}
