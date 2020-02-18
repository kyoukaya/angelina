package angelina

import (
	"flag"
	"net/http"
	"path"

	"github.com/kyoukaya/rhine/log"
	"github.com/kyoukaya/rhine/proxy"
	"github.com/kyoukaya/rhine/utils"
)

var (
	staticDir = flag.String("ange-static", "", "path to static files to serve on the root URL. Serving disabled if empty string.")
	host      = flag.String("ange-host", ":8000", "host on which ange is served")
)

type Hub struct {
	log.Logger

	users   map[string][]*Client
	modules map[string]*proxy.RhineModule

	// Inbound messages from modules when they are initialized.
	modAttach chan *proxy.RhineModule
	// Inbound messages from modules indicating a shutdown.
	modDetach chan *proxy.RhineModule

	// Hub maintains the set of active clients and broadcasts messages to the
	// clients.
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	messages chan *message
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
}

func start(logger log.Logger) {
	if !flag.Parsed() {
		flag.Parse()
	}
	ange := &Hub{
		Logger:     logger,
		modAttach:  make(chan *proxy.RhineModule),
		modDetach:  make(chan *proxy.RhineModule),
		modules:    make(map[string]*proxy.RhineModule),
		users:      make(map[string][]*Client),
		messages:   make(chan *message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
	go ange.run()
	mux := http.NewServeMux()
	if *staticDir != "" {
		dir := *staticDir
		if !path.IsAbs(dir) {
			dir = utils.BinDir + dir
		}
		mux.Handle("/", http.FileServer(http.Dir(dir)))
	}
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ange.ServeWs(w, r)
	})
	proxy.RegisterInitFunc(modName, ange.modInitFunc)
	go func() {
		err := http.ListenAndServe(*host, mux)
		if err != nil {
			ange.Warnln("ListenAndServe: ", err)
			panic(err)
		}
	}()
	ange.Printf("Angelina listening on %s", *host)
}

func init() {
	proxy.OnStart(start)
}
