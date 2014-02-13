package main

import "log"
import "github.com/go-amp/amp"
import "runtime"
import "strconv"
import "flag"
import "fmt"
import "time"

func KeepAlive() {
    for { 
        runtime.Gosched()
        time.Sleep(1 * time.Second) 
    }
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
        ask.Reply()        
    }
}

func BuildSumCommand() *amp.Command {
    /*
     * Need a better way to make these
     * */
    arguments := [2]string{ "a", "b" }
    response := [1]string{ "total" }    
    name := "Sum"
    responder := make(chan *amp.AskBox)
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

func RemoteSum(a int, b int, c *amp.Connection, command *amp.Command) (string, error) {
    m := make(map[string]string)
    m["a"] = strconv.Itoa(a)
    m["b"] = strconv.Itoa(b)    
    reply := make(chan *amp.AnswerBox) 
    go RemoteTrap(reply)   
    tag, err := c.CallRemote(command, &m, reply)        
    return tag, err
}

func RemoteTrap(reply chan *amp.AnswerBox) {    
    answer := <-reply
    log.Println("RemoteTrap",*answer.Response)
}

func client() {
    log.Println("Hello Client!")    
    commands := make(map[string]*amp.Command)
    sum := BuildSumCommand()
    commands[sum.Name] = sum
    prot := amp.Init(&commands)
    c, err := prot.ConnectTCP("127.0.0.1:8000")
    if err != nil { log.Println(err) } else {         
        for i := 0; i < 100; i++ {
            tag, err := RemoteSum(i, i*5, c, sum)
            if err != nil { break }
            log.Println(fmt.Sprintf("%s: ",tag),"CallRemote",i,i*5)
            //time.Sleep(300 * time.Millisecond)            
            
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
