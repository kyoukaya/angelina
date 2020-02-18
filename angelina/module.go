package angelina

import (
	"strconv"

	"github.com/kyoukaya/rhine/proxy"
)

const modName = "Angelina"

func getModIdentifier(mod *proxy.RhineModule) string {
	return mod.Region + "_" + strconv.Itoa(mod.UID)
}

func (hub *Hub) modInitFunc(mod *proxy.RhineModule) {
	hub.modAttach <- mod
	mod.OnShutdown(func(bool) {
		hub.modDetach <- mod
	})
}
