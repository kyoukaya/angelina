package server

import "github.com/kyoukaya/rhine/proxy"

type clientHook struct {
	kind   string // 'gamestate' or 'packet'
	target string
	event  bool
	hook   proxy.Hooker
}

const gameStateHook = "gamestate"
const packetHook = "packet"

func (ch *clientHook) Unhook() {
	ch.hook.Unhook()
}
