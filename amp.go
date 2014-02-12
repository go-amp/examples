package main

import "go/cmn"
import "go/amp"
//import "runtime"
import "strconv"
import "flag"
import "fmt"
//import "time"

func SumRespond(self *amp.Command) {
    for {        
        ask := <- self.Responder
        //cmn.Log(ask)
        m := *ask.Arguments        
        a, _ := strconv.Atoi(m["a"])
        b, _ := strconv.Atoi(m["b"])
        total := a + b        
        cmn.Log("SumRespond:",a,"+",b,"=",total)
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
    cmn.Log("Hello Server!")    
    commands := make(map[string]*amp.Command)
    sum := BuildSumCommand()
    commands[sum.Name] = sum
    prot := amp.Init(&commands)    
    err := prot.ListenTCP(":8000")
    if err != nil { cmn.Log(err) } else { cmn.KeepAlive() }
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
    cmn.Log("RemoteTrap",*answer.Response)
}

func client() {
    cmn.Log("Hello Client!")    
    commands := make(map[string]*amp.Command)
    sum := BuildSumCommand()
    commands[sum.Name] = sum
    prot := amp.Init(&commands)
    c, err := prot.ConnectTCP("127.0.0.1:8000")
    if err != nil { cmn.Log(err) } else {         
        for i := 0; i < 100; i++ {
            tag, err := RemoteSum(i, i*5, c, sum)
            if err != nil { break }
            cmn.Log(fmt.Sprintf("%s: ",tag),"CallRemote",i,i*5)
            //time.Sleep(300 * time.Millisecond)            
            
        }
        cmn.KeepAlive() 
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
