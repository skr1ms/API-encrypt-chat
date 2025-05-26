package websocket

import (
	"crypto-chat-backend/internal/domain/entities"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, user *entities.User) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: user.ID,
		user:   user,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		// Обрабатываем входящее сообщение
		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Добавляем дополнительные сообщения из очереди
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(data []byte) {
	var message WSMessage
	if err := json.Unmarshal(data, &message); err != nil {
		c.hub.logger.Errorf("Failed to unmarshal message: %v", err)
		c.sendError("Invalid message format")
		return
	}

	message.From = c.userID
	message.Timestamp = time.Now().Unix()

	switch message.Type {
	case MessageTypeChat:
		c.handleChatMessage(message)
	case MessageTypeKeyExchange:
		c.handleKeyExchange(message)
	default:
		c.sendError("Unknown message type")
	}
}

func (c *Client) handleChatMessage(message WSMessage) {
	// В реальном приложении здесь должна быть обработка через use case
	// Пока что просто пересылаем сообщение в чат
	if message.ChatID == 0 {
		c.sendError("Chat ID is required")
		return
	}

	// Отправляем сообщение всем участникам чата кроме отправителя
	c.hub.SendToChat(message.ChatID, message, c.userID)
}

func (c *Client) handleKeyExchange(message WSMessage) {
	// Обработка обмена ключами
	if message.To == 0 {
		c.sendError("Recipient ID is required for key exchange")
		return
	}

	// Отправляем сообщение конкретному пользователю
	c.hub.SendToUser(message.To, message)
}

func (c *Client) sendError(errMsg string) {
	errorMessage := WSMessage{
		Type: MessageTypeError,
		Data: map[string]string{
			"error": errMsg,
		},
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(errorMessage)
	if err != nil {
		c.hub.logger.Errorf("Failed to marshal error message: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
	}
}
