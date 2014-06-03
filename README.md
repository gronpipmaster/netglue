# NetGlue [![GoDoc](https://godoc.org/github.com/gronpipmaster/netglue?status.png)](http://godoc.org/github.com/gronpipmaster/netglue)

Layer for net/rpc server and client includes

    Server Pinger -> call remote Ponger.Pong()
    Server Ponger -> call remote Pinger.Ping()

If you want to, of course recursion, Server Pinger -> call remote Ponger.Pong() -> call remote Pinger.Ping() -> call remote Ponger.Pong()

## Getting Started

```bash
go get github.com/gronpipmaster/netglue
```

Server Pinger
```go
package main

import (
    "log"
    "models" //some sharing model
    "github.com/gronpipmaster/netglue"
    "time"
)

type Pinger struct {
}

func (m *Pinger) Ping(args *models.Args, reply *models.Reply) error {
    reply.Foo = "PongFoo"
    reply.Bar = "PongBar"
    reply.Time = args.Time
    return nil
}

func main() {
    netglue.OutAddr = ":8888"
    netglue.InAddr = ":8889"
    netglue.Verbose = true
    result := make(chan interface{})
    foo, err := netglue.NewNetGlue(&Pinger{}, result)
    if err != nil {
        log.Fatalln(err)
    }
    args := models.Args{"Foo", "Bar", time.Now()}
    reply := new(models.Reply)
    //remote method call and send my locale args
    foo.Send("Ponger.Pong", args, reply)
    for {
        select {
        case msg := <-result:
            log.Printf("Received : %+v", msg)
            replyMsg := msg.(*models.Reply)
            if replyMsg != nil {
                args = models.Args{replyMsg.Foo, replyMsg.Bar, replyMsg.Time}
                foo.Send("Ponger.Pong", args, reply)
            }
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```

Server Ponger
```go
package main

import (
    "log"
    "models"  //some sharing model
    "github.com/gronpipmaster/netglue"
    "time"
)

type Ponger struct {
}

func (m *Ponger) Pong(args *models.Args, reply *models.Reply) error {
    reply.Foo = "PingFoo"
    reply.Bar = "PingBar"
    reply.Time = args.Time
    return nil
}

func main() {
    netglue.OutAddr = ":8889"
    netglue.InAddr = ":8888"
    netglue.Verbose = true
    result := make(chan interface{})
    foo, err := netglue.NewNetGlue(&Ponger{}, result)
    if err != nil {
        log.Fatalln(err)
    }
    args := models.Args{"Foo", "Bar", time.Now()}
    reply := new(models.Reply)
    //remote method call and send my locale args
    foo.Send("Pinger.Ping", args, reply)
    for {
        select {
        case msg := <-result:
            log.Printf("Received : %+v", msg)
            replyMsg := msg.(*models.Reply)
            if replyMsg != nil {
                args = models.Args{replyMsg.Foo, replyMsg.Bar, replyMsg.Time}
                foo.Send("Pinger.Ping", args, reply)
            }
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```

And share data
```go
//Sharing data
package models

import (
    "time"
)

type Args struct {
    Foo  string
    Bar  string
    Time time.Time
}

type Reply struct {
    Foo  string
    Bar  string
    Time time.Time
}

```