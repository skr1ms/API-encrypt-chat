package websocket

import (
	"sleek-chat-backend/internal/crypto"
	"sleek-chat-backend/internal/domain/entities"
	"sleek-chat-backend/internal/domain/usecase"
	"crypto/ecdsa"
	"crypto/rsa"
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

// ServeWS - обрабатывает WebSocket подключения и создает нового клиента
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

// readPump - читает сообщения от WebSocket клиента
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

		c.handleMessage(message)
	}
}

// writePump - отправляет сообщения WebSocket клиенту
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

// handleMessage - обрабатывает входящие WebSocket сообщения
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

// handleChatMessage - обрабатывает сообщения чата и отправляет их через usecase
func (c *Client) handleChatMessage(message WSMessage) {
	if message.ChatID == 0 {
		c.sendError("Chat ID is required")
		return
	}

	var chatData map[string]interface{}
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		c.sendError("Invalid message data format")
		return
	}

	if err := json.Unmarshal(dataBytes, &chatData); err != nil {
		c.sendError("Invalid message data format")
		return
	}

	content, ok := chatData["content"].(string)
	if !ok {
		c.sendError("Message content is required")
		return
	}

	messageType, ok := chatData["message_type"].(string)
	if !ok {
		messageType = "text"
	}

	req := &usecase.SendMessageRequest{
		Content:     content,
		MessageType: messageType,
	}

	var ecdsaPrivateKey *ecdsa.PrivateKey
	var rsaPrivateKey *rsa.PrivateKey

	if c.user.ECDSAPrivateKey != "" {
		ecdsaPrivateKey, err = crypto.DeserializeECDSAPrivateKey([]byte(c.user.ECDSAPrivateKey))
		if err != nil {
			c.hub.logger.Errorf("Failed to deserialize ECDSA private key for user %d: %v", c.userID, err)
			c.sendError("Failed to process cryptographic keys")
			return
		}
	}

	if c.user.RSAPrivateKey != "" {
		rsaPrivateKey, err = crypto.DeserializeRSAPrivateKey([]byte(c.user.RSAPrivateKey))
		if err != nil {
			c.hub.logger.Errorf("Failed to deserialize RSA private key for user %d: %v", c.userID, err)
			c.sendError("Failed to process cryptographic keys")
			return
		}
	}
	sentMessage, err := c.hub.chatUseCase.SendMessage(message.ChatID, c.userID, req, ecdsaPrivateKey, rsaPrivateKey)
	if err != nil {
		c.hub.logger.Errorf("Failed to send message via usecase: %v", err)
		c.sendError("Failed to send message: " + err.Error())
		return
	}

	wsMessage := WSMessage{
		Type:   MessageTypeChat,
		ChatID: message.ChatID,
		From:   c.userID,
		Data: ChatMessage{
			ID:             sentMessage.ID,
			ChatID:         sentMessage.ChatID,
			SenderID:       sentMessage.SenderID,
			Content:        content,
			MessageType:    sentMessage.MessageType,
			Nonce:          sentMessage.Nonce,
			IV:             sentMessage.IV,
			HMAC:           sentMessage.HMAC,
			ECDSASignature: sentMessage.ECDSASignature,
			RSASignature:   sentMessage.RSASignature,
			Timestamp:      sentMessage.CreatedAt.Unix(),
		},
		Timestamp: time.Now().Unix(),
	}

	c.hub.SendToChat(message.ChatID, wsMessage, c.userID)
}

// handleKeyExchange - обрабатывает сообщения обмена ключами между пользователями
func (c *Client) handleKeyExchange(message WSMessage) {
	if message.To == 0 {
		c.sendError("Recipient ID is required for key exchange")
		return
	}

	c.hub.SendToUser(message.To, message)
}

// sendError - отправляет сообщение об ошибке клиенту
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
