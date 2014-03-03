package main

import "log"
import "github.com/go-amp/amp"
import "runtime"
import "strconv"
import "flag"
import "fmt"
import "time"
import "sync"
import "os"

var NUM_REQUESTS *int
var test_start time.Time
var responses_mutex = &sync.Mutex{}
var responses_back int = 0
var requests_count int = 0
var previous int = 0
var sent_count int = 0
var isClient *bool
var isServer *bool
var isClientHost *string

func KeepAlive() {
    for { 
        runtime.Gosched()
        time.Sleep(1 * time.Second) 
        if *isClient { 
            log.Println("sent",sent_count,"responses_back",responses_back) 
            //if previous != 0 && previous != *NUM_REQUESTS && responses_back == previous { log.Panic("ouch") }
            previous = responses_back
        } else { log.Println("requests",requests_count) }
    }
}

func init() {
    procs := runtime.NumCPU()
    log.Println("setting number procs to",procs)
    runtime.GOMAXPROCS(procs)
}

func SumRespond(self *amp.Command) {
    for {        
        _ := <- self.Responder
        //log.Println(ask)
        //m := *ask.Arguments        
        //a, _ := strconv.Atoi(m["a"])
        //b, _ := strconv.Atoi(m["b"])
        //total := a + b        
        ////log.Println("SumRespond:",a,"+",b,"=",total)
        //answer := *ask.Response
        //answer["total"] = "11" //strconv.Itoa(total)
        ////log.Println("SumRespond sending",ask)
        //responses_mutex.Lock()
        //requests_count++
        //responses_mutex.Unlock()
        //ask.ReplyChannel <- ask       
        //runtime.Gosched()
    }
}

func BuildSumCommand() *amp.Command {
    /*
     * Need a better way to make these
     * */
    arguments := [2]string{ "a", "b" }
    response := [1]string{ "total" }    
    name := "Sum"
    responder := make(chan *amp.Ask, 100)
    sumCommand := &amp.Command{name, responder, arguments[:], response[:]}    
    go SumRespond(sumCommand)
    return sumCommand
}

func server() {
    log.Println("Hello Server!")    
    commands := make(map[string]*amp.Command)
    sum := BuildSumCommand()
    commands[sum.Name] = sum
    prot := amp.Init(&commands)    
    err := prot.ListenTCP(":8000")
    if err != nil { log.Println(err) } else { KeepAlive() }
}

func RemoteSum(a int, b int, c *amp.Client, command *amp.Command) (string, error) {
    callbox := amp.ResourceCallBox()    
    m := *callbox.Arguments    
    m["a"] = strconv.Itoa(a)
    m["b"] = strconv.Itoa(b)    
    reply := make(chan *amp.CallBox) 
    go RemoteTrap(reply)   
    callbox.Callback = reply
    callbox.Command = command    
    //log.Println("CallRemote",callbox.Command.Name,*callbox.Arguments)
    tag, err := c.CallRemote(callbox)         
    return tag, err
}

func RemoteTrap(reply chan *amp.CallBox) {
    //log.Println("remote trapping",reply)    
    answer := <-reply
    
    //m := *answer.Response
    //a, _ := strconv.Atoi(m["total"])
    responses_mutex.Lock()
    responses_back++    
    //log.Println("RemoteTrap",*answer.Response,"for",*answer.Arguments,"responses_back",responses_back)
    if responses_back == *NUM_REQUESTS {
        now := time.Now()
        fmt.Printf("time taken -- %f\n", float32(now.Sub(test_start))/1000000000.0)
    }
    responses_mutex.Unlock()
    amp.RecycleCallBox(answer)
    //log.Println("done recycling callbox")
}

func client() {
    log.Println("Hello Client!")    
    commands := make(map[string]*amp.Command)
    sum := BuildSumCommand()
    commands[sum.Name] = sum
    prot := amp.Init(&commands)
    c, err := prot.ConnectTCP(*isClientHost)
    if err != nil { log.Println(err) } else {  
        go send_requests(c, sum)
        //log.Println("responses_back",responses_back)
        KeepAlive() 
    }    
}

func send_requests(c *amp.Client, sum *amp.Command) {
    test_start = time.Now()       
    log.Println("sending",*NUM_REQUESTS,"requests")
    for i := 1; i <= *NUM_REQUESTS; i++ {
        //log.Println("client iteration -",i)
        _, err := RemoteSum(i, 0, c, sum)
        if err != nil { log.Println(err) 
        } else { sent_count++ }
        runtime.Gosched()
        
    }
    log.Println("done sending")
}



func main() {
    isServer = flag.Bool("server", false, "use as a server")
    isClient = flag.Bool("client", false, "use as a client")
    isClientHost = flag.String("host","127.0.0.1:8000","host address")
    NUM_REQUESTS = flag.Int("num",100000,"number of requests to do")
    //log.Println("isServer",isServer)
    flag.Parse()    
    if *isServer {
        server()
    } else if *isClient {
        f, err := os.Create("/tmp/logfile")
        log.Println("err",err,f)
        //log.SetOutput(f)
        client()
    } else { flag.Usage() }            
}
