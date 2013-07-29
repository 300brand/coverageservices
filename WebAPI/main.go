package main

import (
	"git.300brand.com/coverage/config"
	"git.300brand.com/coverage/skytypes"
	"github.com/gorilla/handlers"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/service"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type Service struct{}
type RPCManager struct{}
type RPCSearch struct{}

const ServiceName = "WebAPI"

var (
	_       service.ServiceDelegate = &Service{}
	Manager *client.ServiceClient
	Search  *client.ServiceClient

	jsonrpc  = rpc.NewServer()
	cmdOnce  = &skytypes.ClockCommand{Command: "once"}
	cmdStart = &skytypes.ClockCommand{Command: "start", Tick: time.Minute}
	cmdStop  = &skytypes.ClockCommand{Command: "stop"}
)

func init() {
	jsonrpc.RegisterCodec(json.NewCodec(), "application/json")
	jsonrpc.RegisterService(new(RPCManager), "Manager")
	jsonrpc.RegisterService(new(RPCSearch), "Search")
}

// Funcs required for ServiceDelegate

func (s *Service) MethodCalled(m string)                        {}
func (s *Service) MethodCompleted(m string, d int64, err error) {}
func (s *Service) Registered(service *service.Service)          {}
func (s *Service) Started(service *service.Service)             {}
func (s *Service) Stopped(service *service.Service)             {}
func (s *Service) Unregistered(service *service.Service)        {}

// Service funcs

func (m *RPCManager) AddOneFeed(r *http.Request, in *skytypes.NullType, out *skytypes.NullType) (err error) {
	return Manager.SendOnce(nil, "FeedAdder", cmdOnce, skytypes.Null)
}

func (m *RPCManager) StartQueue(r *http.Request, in *skytypes.NullType, out *skytypes.NullType) (err error) {
	return Manager.SendOnce(nil, "FeedAdder", cmdStart, skytypes.Null)
}

func (m *RPCManager) StopQueue(r *http.Request, in *skytypes.NullType, out *skytypes.NullType) (err error) {
	return Manager.SendOnce(nil, "FeedAdder", cmdStop, skytypes.Null)
}

func (m *RPCManager) ProcessNextFeed(r *http.Request, in *skytypes.NullType, out *skytypes.NullType) (err error) {
	return Manager.SendOnce(nil, "FeedProcessor", cmdOnce, skytypes.Null)
}

func (m *RPCManager) StartFeeds(r *http.Request, in *skytypes.NullType, out *skytypes.NullType) (err error) {
	return Manager.SendOnce(nil, "FeedProcessor", cmdStart, skytypes.Null)
}

func (m *RPCManager) StopFeeds(r *http.Request, in *skytypes.NullType, out *skytypes.NullType) (err error) {
	return Manager.SendOnce(nil, "FeedProcessor", cmdStop, skytypes.Null)
}

func (m *RPCSearch) Search(r *http.Request, in *skytypes.SearchQuery, out *skytypes.SearchQueryResponse) (err error) {
	return Search.SendOnce(nil, "Search", in, out)
}

// Main

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(ServiceName + " ")

	// Skynet Client
	cc, _ := skynet.GetClientConfig()
	c := client.NewClient(cc)

	Manager = c.GetService("Manager", "", "", "")
	Search = c.GetService("Search", "", "", "")

	// RPC
	listener, err := net.Listen("tcp", config.RPCServer.Address)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go func(l net.Listener) {
		log.Printf("Listening on %s", l.Addr())
		http.Handle("/rpc", handlers.LoggingHandler(os.Stdout, jsonrpc))
		log.Fatal(http.Serve(l, nil))
	}(listener)

	// Skynet Service
	sc, _ := skynet.GetServiceConfig()
	sc.Name = ServiceName
	sc.Region = "Management"
	sc.Version = "1"
	s := service.CreateService(&Service{}, sc)
	defer s.Shutdown()

	s.Start(true).Wait()
}
