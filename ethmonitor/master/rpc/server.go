package rpc

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Server struct {
	serverPort int

	// events
	newNodeFeed        event.Feed
	newTxFeed          event.Feed
	txStateChangesFeed event.Feed
	newChainHeadFeed   event.Feed
	newChainSideFeed   event.Feed
	txExecutedFeed     event.Feed
	peerAddFeed        event.Feed
	peerRemoveFeed     event.Feed

	traverseTxFeed event.Feed

	scope event.SubscriptionScope
}

func NewServer(port int) *Server {
	return &Server{
		serverPort: port,
	}
}

func (s *Server) Start() {
	// RPC listener
	e := rpc.Register(s)
	if e != nil {
		panic(e)
	}
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", s.serverPort))
	if e != nil {
		log.Error("RPC server listen error", "e", e.Error())
	}
	go http.Serve(l, nil)
	log.Info("RPC server started", "port", s.serverPort)

	//Json RPC listener
	l, e = net.Listen("tcp", fmt.Sprintf(":%d", common.ServerPortJson))
	if e != nil {
		log.Error("Json RPC server listen error", "e", e.Error())
		return
	}
	go http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("in handler", "req", r.Body)
		serverCodec := jsonrpc.NewServerCodec(&HttpConn{in: r.Body, out: w})
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(200)
		err := rpc.ServeRequest(serverCodec)
		if err != nil {
			log.Error("Error while serving JSON request", "e", err)
			http.Error(w, "Error while serving JSON request, details have been logged.", 500)
			return
		}
	}))
	log.Info("JSON RPC server started", "port", common.ServerPortJson)
}

// adapt HTTP connection to ReadWriteCloser
type HttpConn struct {
	in  io.Reader
	out io.Writer
}

func (c *HttpConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HttpConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }
func (c *HttpConn) Close() error                      { return nil }

func (s *Server) StartJson() {
	//// 新建Server
	//server := rpc.NewServer()
	//
	//// 开始监听
	//listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.serverPortJson))
	//if err != nil {
	//	log.Error("server listen error", "err", err.Error())
	//	return
	//}
	//defer listener.Close()
	//
	//log.Info("RPC json server started", "port", s.serverPortJson)
	//
	//// 注册处理器
	//err = server.Register(s)
	//if err != nil {
	//	panic(err)
	//}
	//server.HandleHTTP("/", "/debug")
	//// 等待并处理链接

}

func (s *Server) SubscribeNewNodeEvent(ch chan<- common.NewNodeEvent) event.Subscription {
	return s.scope.Track(s.newNodeFeed.Subscribe(ch))
}

func (s *Server) SubscribeNewTxEvent(ch chan<- common.NewTxEvent) event.Subscription {
	return s.scope.Track(s.newTxFeed.Subscribe(ch))
}

func (s *Server) SubscribeNewChainHeadEvent(ch chan<- common.NewChainHeadEvent) event.Subscription {
	return s.scope.Track(s.newChainHeadFeed.Subscribe(ch))
}

func (s *Server) SubscribeNewChainSideEvent(ch chan<- common.NewChainSideEvent) event.Subscription {
	return s.scope.Track(s.newChainSideFeed.Subscribe(ch))
}

func (s *Server) SubscribePeerAddEvent(ch chan<- common.PeerAddEvent) event.Subscription {
	return s.scope.Track(s.peerAddFeed.Subscribe(ch))
}

func (s *Server) SubscribePeerRemoveEvent(ch chan<- common.PeerRemoveEvent) event.Subscription {
	return s.scope.Track(s.peerRemoveFeed.Subscribe(ch))
}

func (s *Server) SubscribeTraverseTxEvent(ch chan<- common.TraverseTxEvent) event.Subscription {
	return s.scope.Track(s.traverseTxFeed.Subscribe(ch))
}

/*
RPC Methods called by geth nodes
*/
func (s *Server) OnNewTxRPC(msg *NewTxMsg, reply *Reply) (e error) {
	log.Debug("Got new transaction", "role", msg.Role, "tx", common.PrettifyHash(msg.Hash))
	s.newTxFeed.Send(common.NewTxEvent{
		Hash:   msg.Hash,
		Role:   msg.Role,
		Sender: msg.Sender,
		Nonce:  msg.Nonce,
	})
	return
}

func (s *Server) OnNodeStartRPC(msg *NodeStartMsg, reply *Reply) (e error) {
	log.Debug("Node started", "role", msg.Role)
	s.newNodeFeed.Send(common.NewNodeEvent{
		Role:        msg.Role,
		ServerPort:  msg.ServerPort,
		URL:         msg.URL,
		BlockNumber: msg.BlockNumber,
		BlockHash:   msg.BlockHash,
		Td:          msg.Td,
	})
	return
}

func (s *Server) OnNewChainHeadRPC(msg *NewChainHeadMsg, reply *Reply) (e error) {
	log.Debug("Node got new chain head", "role", msg.Role, "number", msg.Number, "td", msg.Td, "txs", len(msg.Txs), "hash", msg.Hash)
	s.newChainHeadFeed.Send(common.NewChainHeadEvent{
		Role:   msg.Role,
		Hash:   msg.Hash,
		Number: msg.Number,
		Td:     msg.Td,
		Txs:    msg.Txs,
	})
	return
}

func (s *Server) OnNewChainSideRPC(msg *NewChainSideMsg, reply *Reply) (e error) {
	log.Debug("Node got new chain side", "role", msg.Role, "number", msg.Number, "td", msg.Td, "txs", len(msg.Txs), "hash", msg.Hash)
	s.newChainSideFeed.Send(common.NewChainSideEvent{
		Role:   msg.Role,
		Hash:   msg.Hash,
		Number: msg.Number,
		Td:     msg.Td,
		Txs:    msg.Txs,
	})
	return
}

func (s *Server) OnPeerAddRPC(msg *PeerAddMsg, reply *Reply) (e error) {
	log.Debug("Node add peer", "role", msg.Role)
	s.peerAddFeed.Send(common.PeerAddEvent{Role: msg.Role})
	return
}

func (s *Server) OnPeerRemoveRPC(msg *PeerRemoveMsg, reply *Reply) (e error) {
	log.Debug("Node remove peer", "role", msg.Role)
	s.peerRemoveFeed.Send(common.PeerRemoveEvent{Role: msg.Role})
	return
}

/*
Json RPC methods called by dArcher
*/
//func (s *Server) TraverseTxRPC(msg *TraverseTxMsg, reply *Reply) (e error) {
//	log.Debug("dArcher tells to start traversing tx lifecycle", "tx", msg.Hash[:8])
//	s.traverseTxFeed.Send(TraverseTxEvent{Hash: msg.Hash})
//	return
//}
//
//func (s *Server) Hello(msg *string, reply *Reply) (e error) {
//	log.Debug("hello " + *msg)
//	reply.Err = fmt.Errorf("hello " + *msg)
//	return
//}
