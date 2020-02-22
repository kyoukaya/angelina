package angelina

import "github.com/kyoukaya/rhine/proxy"

type clientHook struct {
	kind   string // 'gamestate' or 'packet'
	target string
	event  bool
	hook   proxy.Hooker
}

func (ch *clientHook) Unhook() {
	ch.hook.Unhook()
}

const gameStateHook = "gamestate"
const packetHook = "packet"
