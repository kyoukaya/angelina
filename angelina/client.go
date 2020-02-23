package angelina

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/gorilla/websocket"
	"github.com/kyoukaya/rhine/proxy"
	"github.com/kyoukaya/rhine/proxy/gamestate"

	"github.com/kyoukaya/angelina/angelina/msg"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub         *Hub
	mod         *proxy.RhineModule
	hookCounter uint64
	hooks       map[uint64]*clientHook
	listener    chan gamestate.StateEvent

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *Client) sendWrapper(data []byte) {
	select {
	case c.send <- data:
	default:
		c.hub.Warnf("[Ange] Failed to send data to %p", c)
	}
}

func (c *Client) removeHook(id uint64) error {
	hook := c.hooks[id]
	if hook == nil {
		return fmt.Errorf("Unable to find hook ID %d to unhook", id)
	}
	hook.Unhook()
	delete(c.hooks, id)
	return nil
}

func (c *Client) unhookAll() {
	for k, hook := range c.hooks {
		hook.Unhook()
		delete(c.hooks, k)
	}
}

func (c *Client) addHook(data *msg.Hook) error {
	var hook proxy.Hooker
	switch data.Kind {
	case packetHook:
		hook = c.mod.Hook(data.Target, 0, c.hookHandler)
	case gameStateHook:
		hook = c.mod.StateHook(data.Target, c.listener, !data.Event)
	default:
		return fmt.Errorf("Unknown hook type '%s'", data.Kind)
	}
	c.hooks[c.hookCounter] = &clientHook{
		kind:   data.Kind,
		target: data.Target,
		event:  data.Event,
		hook:   hook,
	}

	ret, err := msg.ServerHooked(c.hookCounter, data.Kind, data.Target, data.Event)
	if err != nil {
		return err
	}
	c.hookCounter++
	c.sendWrapper(ret)
	return nil
}

func (c *Client) hookHandler(op string, data []byte, pktCtx *goproxy.ProxyCtx) []byte {
	b, err := msg.ServerHookEvt(packetHook, op, json.RawMessage(data))
	if err != nil {
		c.hub.Warnln("[Ange] ", err)
		return data
	}
	c.sendWrapper(b)
	return data
}

func (c *Client) stateListener() {
	for {
		l, open := <-c.listener
		if !open {
			return
		}
		b, err := msg.ServerHookEvt(gameStateHook, l.Path, l.Payload)
		if err != nil {
			c.hub.Warnln("[Ange] ", err)
			continue
		}
		c.sendWrapper(b)
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	// c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		c.hub.Warnln("[Ange] ", err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				c.hub.Warnln("[Ange] ", err)
			}
			break
		}
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		c.hub.messages <- &messageT{
			client:  c,
			payload: msg,
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				return
			}
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				c.hub.Warnln("[Ange] ", err)
				continue
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, err = w.Write(message)
			if err != nil {
				c.hub.Warnln("[Ange] ", err)
				continue
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err = w.Write(newline)
				if err != nil {
					c.hub.Warnln("[Ange] ", err)
					continue
				}
				_, err = w.Write(<-c.send)
				if err != nil {
					c.hub.Warnln("[Ange] ", err)
					continue
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				c.hub.Warnln("[Ange] ", err)
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func (hub *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.Warnf("[Ange] %s", err.Error())
		return
	}
	client := &Client{
		hub:      hub,
		hooks:    make(map[uint64]*clientHook),
		listener: make(chan gamestate.StateEvent, 32),
		conn:     conn,
		send:     make(chan []byte, 128),
	}
	client.hub.register <- client

	go client.stateListener()
	go client.writePump()
	go client.readPump()
}
