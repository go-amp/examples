from twisted.internet.protocol import ClientCreator
from twisted.protocols import amp
from twisted.internet import reactor, defer
from twisted.internet.protocol import Factory
import sys
import time

num_requests = 100000
received = 0
start = time.time()

def sleep(secs):
    d = defer.Deferred()
    reactor.callLater(secs, d.callback, None)
    return d

class Sum(amp.Command):
       arguments = [('a', amp.Integer()),
                    ('b', amp.Integer())]
       response = [('total', amp.Integer())]
              
class JustSum(amp.AMP):
       def sum(self, a, b):
           total = a + b
           #print 'Did a sum: %d + %d = %d' % (a, b, total)
           return {'total': total}
       Sum.responder(sum)

def server():
    print 'server..'
    factory = Factory()
    factory.protocol = JustSum
    reactor.listenTCP(8000, factory)
    reactor.run()
    

    
class Client():
    def __init__(self):
        print 'client..'
        ClientCreator(reactor, amp.AMP).connectTCP("127.0.0.1", 8000).addCallback(self.connected)

    @defer.inlineCallbacks
    def connected(self, p):
        print 'sending..'
        p.callRemote(Sum, a=333, b=333).addCallback(self.got_sum)
        global num_requests
        print 'doing',num_requests,'requests'
        global start
        start = time.time()
        for i in xrange(num_requests):
            yield
            #print 'sending again..'
            p.callRemote(Sum, a=i, b=0).addCallback(self.got_sum)
            #p.transport.loseConnection()
  
    def got_sum(self, result):
        #print 'result',result
        global received
        received += 1        
        if received == num_requests:
            global start
            print 'done',time.time()-start
        #reactor.stop()
    
if __name__== """__main__""":
    c = sys.argv[1]
    if c == 'server':
        server()
    if c == 'client':
        Client()
        reactor.run()
