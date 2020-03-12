package server

import (
	"net/http"
	"path"

	"github.com/gorilla/websocket"
	"github.com/kyoukaya/rhine/log"
	"github.com/kyoukaya/rhine/proxy"
	"github.com/kyoukaya/rhine/utils"
	"github.com/kyoukaya/rhine/utils/gamedata"
)

type Ange struct {
	log.Logger
	gamedata  *gamedata.GameData
	staticDir string
	host      string
	upgrader  websocket.Upgrader

	// Maps a user ID to a slice of attached clients
	attachedClients map[string][]*Client
	// Maps a user ID to their RhineModule
	modules map[string]*angeModule
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from modules when they are initialized.
	modAttach chan *angeModule
	// Inbound messages from modules indicating a shutdown.
	modDetach chan *angeModule
	// Inbound messages from the clients.
	messages chan *messageT
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
}

const (
	wsReadBufSiz  = 512
	wsWriteBufSiz = 1024
)

func New(staticDir, host string, unsafeOrigin bool) *Ange {
	var checkFunc func(r *http.Request) bool
	if unsafeOrigin {
		checkFunc = func(r *http.Request) bool {
			return true
		}
	}
	ange := &Ange{
		staticDir: staticDir,
		host:      host,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  wsReadBufSiz,
			WriteBufferSize: wsWriteBufSiz,
			CheckOrigin:     checkFunc,
		},
		attachedClients: make(map[string][]*Client),
		modules:         make(map[string]*angeModule),
		clients:         make(map[*Client]bool),
		modAttach:       make(chan *angeModule),
		modDetach:       make(chan *angeModule),
		messages:        make(chan *messageT),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
	}
	return ange
}

func (ange *Ange) Run(logger log.Logger) {
	gd, err := gamedata.New("GL", logger)
	if err != nil {
		panic(err)
	}
	ange.gamedata = gd
	ange.Logger = logger
	go ange.runHub()
	mux := http.NewServeMux()
	if ange.staticDir != "" {
		dir := ange.staticDir
		if !path.IsAbs(dir) {
			dir = utils.BinDir + dir
		}
		mux.Handle("/", http.FileServer(http.Dir(dir)))
	}
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ange.ServeWs(w, r)
	})
	mux.Handle("/ange/static/",
		http.StripPrefix("/ange/static/", http.FileServer(http.Dir(utils.BinDir+"data"))))
	proxy.RegisterInitFunc(modName, ange.modInitFunc)
	go func() {
		ange.Printf("[Ange] listening on %s", ange.host)
		err := http.ListenAndServe(ange.host, mux)
		if err != nil {
			ange.Warnln("[Ange] ListenAndServe: ", err)
			panic(err)
		}
	}()
}
