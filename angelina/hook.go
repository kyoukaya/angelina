package angelina

type angeHook struct {
	kind   string // 'gamestate' or 'packet'
	target string
	event  bool
	owner  *Client
}

const gameStateHook = "gamestate"
const packetHook = "packet"
