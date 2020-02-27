package angelina

import (
	"flag"
	"net/http"
	"path"

	"github.com/kyoukaya/rhine/log"
	"github.com/kyoukaya/rhine/proxy"
	"github.com/kyoukaya/rhine/utils"
	"github.com/kyoukaya/rhine/utils/gamedata"
)

var (
	staticDir = flag.String("ange-static", "", "path to static files to serve on the root URL. Serving disabled if empty string.")
	host      = flag.String("ange-host", ":8000", "host on which ange is served")
)

type Hub struct {
	log.Logger
	gamedata *gamedata.GameData

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

func start(logger log.Logger) {
	if !flag.Parsed() {
		flag.Parse()
	}
	gd, err := gamedata.New("GL", logger)
	if err != nil {
		panic(err)
	}
	ange := &Hub{
		Logger:          logger,
		attachedClients: make(map[string][]*Client),
		modules:         make(map[string]*angeModule),
		clients:         make(map[*Client]bool),
		modAttach:       make(chan *angeModule),
		modDetach:       make(chan *angeModule),
		messages:        make(chan *messageT),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		gamedata:        gd,
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
	mux.Handle("/ange/static/",
		http.StripPrefix("/ange/static/", http.FileServer(http.Dir(utils.BinDir+"data"))))
	proxy.RegisterInitFunc(modName, ange.modInitFunc)
	go func() {
		err := http.ListenAndServe(*host, mux)
		if err != nil {
			ange.Warnln("[Ange] ListenAndServe: ", err)
			panic(err)
		}
	}()
	ange.Printf("[Ange] listening on %s", *host)
}

func init() {
	proxy.OnStart(start)
}
