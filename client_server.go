package main

//import "net" 
import "log"
//import "fmt" 
import "time"
import "flag"
import "runtime"
import "github.com/go-amp/amp"

var NUM_REQUESTS *int

var isClient *bool
var isServer *bool
var isClientHost *string
var sent_count int = 0
var requests_count int = 0
var received_back int = 0
var startTime time.Time
const SUM_COMMAND string = "Sum"
var done = false

func KeepAlive() {
    for { 
        runtime.Gosched()
        time.Sleep(1 * time.Second) 
        if *isClient && !done { 
            log.Println("sent",sent_count,"received_back",received_back)             
        } else {  }
    }
}

func server() {
    prot := amp.Init()
    responder := make(chan *amp.AskBox)
    prot.RegisterResponder(SUM_COMMAND, responder)
    go do_sum(responder)
    prot.ListenTCP(":8000")
    KeepAlive()
}

func do_sum(in chan *amp.AskBox) {
    for ask := range in {
        //log.Println(*ask.Args)
        ask.Response["i"] = []byte("Buenos Vida")
        requests_count++
        ask.Reply()
    }
}

func response_trap(in chan *amp.CallBox) { 
    for reply := range in {
        //log.Println(*reply.Response)
        amp.RecycleCallBox(reply)
        received_back++
        if received_back == *NUM_REQUESTS {
            done = true
            endTime := time.Now()
            log.Println("ElapsedTime:", endTime.Sub(startTime))
            close(in)
        }
    }
}

func client() {
    prot := amp.Init()    
    c, err := prot.ConnectTCP(*isClientHost)
    if err != nil { return }    
    go send_requests(c)
    KeepAlive()    
}

func send_requests(c *amp.Client) {
    replies := make(chan *amp.CallBox)
    go response_trap(replies)
    startTime = time.Now()   
    for i := 1; i <= *NUM_REQUESTS; i++ {
        //send := []byte{0,1,97,0,6,54,54,50,55,49,54,0,1,98,0,1,48,0,4,95,97,115,107,0,5,97,49,99,98,99,0,8,95,99,111,109,109,97,110,100,0,3,83,117,109,0,0}
        //log.Println("writing",send)
        box := amp.ResourceCallBox()
        box.Args["i"] = []byte("hi there!")
        box.Callback = replies
        err := c.CallRemote(SUM_COMMAND, box)
        //_, err := c.Conn.Write(send)
        if err != nil { log.Println("err",err); break }
        sent_count++
        runtime.Gosched()
    }
    
}

func main() {
    isServer = flag.Bool("server", false, "use as a server")
    isClient = flag.Bool("client", false, "use as a client")
    isClientHost = flag.String("host","127.0.0.1:8000","host address")
    NUM_REQUESTS = flag.Int("num",100000,"number of requests to do")
    log.Println("hi")
    flag.Parse()    
    if *isServer {
        server()
    } else if *isClient {        
        client()
    } else { flag.Usage() }  
}

