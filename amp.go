package main

import "log"
import "github.com/go-amp/amp"
import "runtime"
import "strconv"
import "flag"
//import "fmt"
import "time"

func KeepAlive() {
    for { 
        runtime.Gosched()
        time.Sleep(1 * time.Second) 
    }
}

func init() {
    procs := runtime.NumCPU()
    log.Println("setting number procs to",procs)
    runtime.GOMAXPROCS(procs)
}

func SumRespond(self *amp.Command) {
    for {        
        ask := <- self.Responder
        //log.Println(ask)
        m := *ask.Arguments        
        a, _ := strconv.Atoi(m["a"])
        b, _ := strconv.Atoi(m["b"])
        total := a + b        
        log.Println("SumRespond:",a,"+",b,"=",total)
        answer := *ask.Response
        answer["total"] = strconv.Itoa(total)
        log.Println("SumRespond sending",ask)
        ask.ReplyChannel <- ask       
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
    err := prot.ListenTCP("127.0.0.1:8000")
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
    log.Println("CallRemote",callbox.Command.Name,*callbox.Arguments)
    tag, err := c.CallRemote(callbox)         
    return tag, err
}

func RemoteTrap(reply chan *amp.CallBox) {
    //log.Println("remote trapping",reply)    
    answer := <-reply
    log.Println("RemoteTrap",*answer.Response,"for",*answer.Arguments)
    amp.RecycleCallBox(answer)
    //log.Println("done recycling callbox")
}

func client() {
    log.Println("Hello Client!")    
    commands := make(map[string]*amp.Command)
    sum := BuildSumCommand()
    commands[sum.Name] = sum
    prot := amp.Init(&commands)
    c, err := prot.ConnectTCP("127.0.0.1:8000")
    if err != nil { log.Println(err) } else {         
        for i := 1; i < 1000000; i++ {
            log.Println("client iteration -",i)
            _, err := RemoteSum(i, 1, c, sum)
            if err != nil { log.Println(err); break }                   
        }
        KeepAlive() 
    }
    
}

func main() {
    isServer := flag.Bool("server", false, "use as a server")
    isClient := flag.Bool("client", false, "use as a client")
    flag.Parse()    
    if *isServer {
        server()
    } else if *isClient {
        client()
    } else { flag.Usage() }            
}
