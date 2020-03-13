package main

import (
	"flag"
	"log"

	"github.com/kyoukaya/angelina/server"
	"github.com/kyoukaya/rhine/proxy"
)

var logPath = flag.String("log-path", "logs/proxy.log", "file to output the log to")
var silent = flag.Bool("silent", false, "don't print anything to stdout")
var filter = flag.Bool("filter", false, "enable the host filter")
var verbose = flag.Bool("v", false, "print Rhine verbose messages")
var verboseGoProxy = flag.Bool("v-goproxy", false, "print verbose goproxy messages")
var host = flag.String("host", ":8080", "host on which the proxy is served")
var disableCertStore = flag.Bool("disable-cert-store", false, "disables the built in certstore, reduces memory usage but increases HTTP latency and CPU usage")
var unsafeOrigin = flag.Bool("unsafe-origin", true, "allow any HTTP request, "+
	"no matter what origin they specify, to upgrade into a ws connection")
var staticDir = flag.String("ange-static", "", "path to static files to serve on the root URL. Serving disabled if empty string.")
var angeHost = flag.String("ange-host", ":8000", "host on which ange is served")
var noUnknownJSON = flag.Bool("no-unk-json", false, "disallows unknown fields when unmarshalling json in the gamestate module")

func main() {
	flag.Parse()
	options := &proxy.Options{
		LogPath:          *logPath,
		LogDisableStdOut: *silent,
		EnableHostFilter: *filter,
		LoggerFlags:      log.Lshortfile | log.Ltime,
		Verbose:          *verbose,
		VerboseGoProxy:   *verboseGoProxy,
		Address:          *host,
		DisableCertStore: *disableCertStore,
		NoUnknownJSON:    *noUnknownJSON,
	}
	rhine := proxy.NewProxy(options)
	ange := server.New(*staticDir, *angeHost, *unsafeOrigin)
	ange.Run(rhine.Logger)
	rhine.Start()
}
