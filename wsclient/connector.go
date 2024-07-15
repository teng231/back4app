package wsclient

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type Conn struct {
	conn *websocket.Conn
	// URL ex: ws://localhost:8080/ws
	URL string
}

func NewConn(url string, headers map[string]string) (*Conn, error) {
	client := &Conn{URL: url}
	reqHeader := make(http.Header)
	if len(headers) > 0 {
		for key, val := range headers {
			reqHeader[key] = []string{val}
		}
	}
	c, _, err := websocket.DefaultDialer.Dial(url, reqHeader)
	if err != nil {
		return nil, err
	}
	client.conn = c
	return client, nil
}

func (c *Conn) Write(msg any) error {
	val, _ := json.Marshal(msg)
	err := c.conn.WriteMessage(websocket.TextMessage, val)
	if err != nil {
		log.Println("write:", err)
	}
	return err
}

func (c *Conn) Read(readerbuf chan []byte) {
	for {
		msgType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Close()
			}
			log.Println("read:", err)
			return
		}
		// log.Printf("recv: %d   %s", msgType, message)

		if msgType == websocket.BinaryMessage || msgType == websocket.TextMessage {
			readerbuf <- message
		}
	}
}

func (c *Conn) Close() {
	c.conn.Close()
}
