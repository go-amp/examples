from twisted.internet.protocol import ClientCreator
from twisted.protocols import amp
from twisted.internet import reactor, defer
from twisted.internet.protocol import Factory
import sys

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
           print 'Did a sum: %d + %d = %d' % (a, b, total)
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
        for i in xrange(100):
            yield
            print 'sending again..'
            p.callRemote(Sum, a=i, b=1).addCallback(self.got_sum)
            #p.transport.loseConnection()
  
    def got_sum(self, result):
        print 'result',result
        #reactor.stop()
    
if __name__== """__main__""":
    c = sys.argv[1]
    if c == 'server':
        server()
    if c == 'client':
        Client()
        reactor.run()
