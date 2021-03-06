package WebAPI

import (
	"github.com/300brand/coverage"
	"github.com/300brand/coverage/social"
	"github.com/300brand/coverageservices/service"
	"github.com/300brand/coverageservices/types"
	"github.com/300brand/disgo"
	"github.com/300brand/go-toml-config"
	"github.com/300brand/logger"
	"github.com/gorilla/handlers"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net"
	"net/http"
	"time"
)

type Service struct {
	client *disgo.Client
}

type logWriter struct{}

type RPCArticle struct{ s *Service }
type RPCFeed struct{ s *Service }
type RPCManager struct{ s *Service }
type RPCPublication struct{ s *Service }
type RPCSearch struct{ s *Service }
type RPCSocial struct{ s *Service }

var (
	_ service.Service = new(Service)

	cfgHttpListen = config.String("WebAPI.httplisten", ":8080")

	jsonrpc  = rpc.NewServer()
	cmdOnce  = &types.ClockCommand{Command: "once"}
	cmdStart = &types.ClockCommand{Command: "start", Tick: time.Minute}
	cmdStop  = &types.ClockCommand{Command: "stop"}
)

func init() {
	service.Register("WebAPI", new(Service))
	jsonrpc.RegisterCodec(json.NewCodec(), "application/json")
}

// Logwriter functions
func (l logWriter) Write(b []byte) (n int, err error) {
	logger.Debug.Printf("WebAPI: %s", b)
	return len(b), nil
}

// Funcs required for Service

func (s *Service) Start(client *disgo.Client) (err error) {
	s.client = client

	jsonrpc.RegisterService(&RPCArticle{s}, "Article")
	jsonrpc.RegisterService(&RPCFeed{s}, "Feed")
	jsonrpc.RegisterService(&RPCManager{s}, "Manager")
	jsonrpc.RegisterService(&RPCPublication{s}, "Publication")
	jsonrpc.RegisterService(&RPCSearch{s}, "Search")
	jsonrpc.RegisterService(&RPCSocial{s}, "Social")

	s.StartRPC()
	return
}

// Service funcs

func (s *Service) StartRPC() (err error) {
	listener, err := net.Listen("tcp", *cfgHttpListen)
	if err != nil {
		logger.Error.Fatal(err)
		return
	}

	go func(l net.Listener) {
		defer l.Close()
		logger.Debug.Printf("WebAPI: HTTP RPC Listening on %s", l.Addr())
		w := new(logWriter)
		http.Handle("/rpc", handlers.LoggingHandler(w, jsonrpc))
		http.HandleFunc("/exportSearch/", s.HandleExport)
		logger.Error.Fatal(http.Serve(l, nil))
	}(listener)

	return
}

// Bunk func to get service registered
func (s *Service) NOOP(in, out *disgo.NullType) (err error) { return }

// RPC Funcs

func (m *RPCArticle) Get(r *http.Request, in *types.ObjectId, out *coverage.Article) (err error) {
	return m.s.client.Call("StorageReader.Article", in, out)
}

func (m *RPCArticle) Reprocess(r *http.Request, in *types.ObjectId, out *disgo.NullType) (err error) {
	a := new(coverage.Article)
	if err = m.Get(r, in, a); err != nil {
		return
	}
	return m.s.client.Call("Article.Process", a, new(disgo.NullType))
}

func (m *RPCFeed) Add(r *http.Request, in *types.NewFeed, out *coverage.Feed) (err error) {
	return m.s.client.Call("Feed.Add", in, out)
}

func (m *RPCFeed) Process(r *http.Request, in *types.ObjectId, out *disgo.NullType) (err error) {
	go m.s.client.Call("Feed.Process", in, out)
	return nil
}

func (m *RPCFeed) Remove(r *http.Request, in *types.ObjectId, out *disgo.NullType) (err error) {
	return m.s.client.Call("Feed.Remove", in, out)
}

func (m *RPCManager) ProcessNextFeed(r *http.Request, in *disgo.NullType, out *disgo.NullType) (err error) {
	return m.s.client.Call("Manager.FeedProcessor", cmdOnce, disgo.Null)
}

func (m *RPCManager) StartFeeds(r *http.Request, in *disgo.NullType, out *disgo.NullType) (err error) {
	return m.s.client.Call("Manager.FeedProcessor", cmdStart, disgo.Null)
}

func (m *RPCManager) Stats(r *http.Request, in *disgo.NullType, out *types.Stats) (err error) {
	return m.s.client.Call("StorageReader.Stats", in, &out.Database)
}

func (m *RPCManager) StopFeeds(r *http.Request, in *disgo.NullType, out *disgo.NullType) (err error) {
	return m.s.client.Call("Manager.FeedProcessor", cmdStop, disgo.Null)
}

func (m *RPCPublication) Add(r *http.Request, in *types.Pub, out *coverage.Publication) (err error) {
	return m.s.client.Call("Publication.Add", in, out)
}

func (m *RPCPublication) Get(r *http.Request, in *types.ObjectId, out *coverage.Publication) (err error) {
	return m.s.client.Call("StorageReader.Publication", in, out)
}

func (m *RPCPublication) GetAll(r *http.Request, in *types.MultiQuery, out *types.MultiPubs) (err error) {
	return m.s.client.Call("StorageReader.Publications", in, out)
}

func (m *RPCPublication) Set(r *http.Request, in *types.Set, out *disgo.NullType) (err error) {
	return m.s.client.Call("StorageWriter.UpdatePublication", in, out)
}

func (m *RPCPublication) View(r *http.Request, in *types.ViewPubQuery, out *types.ViewPub) (err error) {
	return m.s.client.Call("Publication.View", in, out)
}

func (m *RPCSearch) Group(r *http.Request, in *types.GroupQuery, out *types.SearchQueryResponse) (err error) {
	return m.s.client.Call("Search.GroupSearch", in, out)
}

func (m *RPCSearch) Search(r *http.Request, in *types.SearchQuery, out *types.SearchQueryResponse) (err error) {
	return m.s.client.Call("Search.Search", in, out)
}

func (m *RPCSocial) Article(r *http.Request, in *types.ObjectId, out *social.Stats) (err error) {
	a := new(coverage.Article)
	if err = m.s.client.Call("StorageReader.Article", in, a); err != nil {
		return err
	}
	return m.s.client.Call("Social.Article", a, out)
}
