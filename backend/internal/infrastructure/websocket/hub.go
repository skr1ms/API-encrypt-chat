package websocket

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/usecase"
	"crypto-chat-backend/pkg/logger"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// В продакшене нужно проверять origin
		return true
	},
}

type Hub struct {
	clients     map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	logger      *logger.Logger
	chatUseCase *usecase.ChatUseCase
	mu          sync.RWMutex
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID uint
	user   *entities.User
}

type MessageType string

const (
	MessageTypeChat         MessageType = "chat"
	MessageTypeNotification MessageType = "notification"
	MessageTypeUserStatus   MessageType = "user_status"
	MessageTypeKeyExchange  MessageType = "key_exchange"
	MessageTypeError        MessageType = "error"
)

type WSMessage struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	From      uint        `json:"from,omitempty"`
	To        uint        `json:"to,omitempty"`
	ChatID    uint        `json:"chat_id,omitempty"`
}

type ChatMessage struct {
	ID             uint   `json:"id"`
	ChatID         uint   `json:"chat_id"`
	SenderID       uint   `json:"sender_id"`
	Content        string `json:"content"`
	MessageType    string `json:"message_type"`
	Nonce          string `json:"nonce"`
	IV             string `json:"iv"`
	HMAC           string `json:"hmac"`
	ECDSASignature string `json:"ecdsa_signature"`
	RSASignature   string `json:"rsa_signature"`
	Timestamp      int64  `json:"timestamp"`
}

type UserStatusMessage struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	IsOnline bool   `json:"is_online"`
}

func NewHub(logger *logger.Logger, chatUseCase *usecase.ChatUseCase) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		logger:      logger,
		chatUseCase: chatUseCase,
	}
}

// SetChatUseCase устанавливает chatUseCase после создания Hub
// Это нужно для разрешения циклической зависимости между Hub и ChatUseCase
func (h *Hub) SetChatUseCase(chatUseCase *usecase.ChatUseCase) {
	h.chatUseCase = chatUseCase
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			h.logger.Infof("Client connected: user_id=%d", client.userID)

			// Уведомляем других клиентов о том, что пользователь онлайн
			h.broadcastUserStatus(client.userID, client.user.Username, true)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

			h.logger.Infof("Client disconnected: user_id=%d", client.userID)

			// Уведомляем других клиентов о том, что пользователь офлайн
			h.broadcastUserStatus(client.userID, client.user.Username, false)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) broadcastUserStatus(userID uint, username string, isOnline bool) {
	message := WSMessage{
		Type: MessageTypeUserStatus,
		Data: UserStatusMessage{
			UserID:   userID,
			Username: username,
			IsOnline: isOnline,
		},
		Timestamp: getTimestamp(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Errorf("Failed to marshal user status message: %v", err)
		return
	}

	h.broadcast <- data
}

func (h *Hub) SendToUser(userID uint, message WSMessage) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}

	return nil
}

func (h *Hub) SendToChat(chatID uint, message WSMessage, excludeUserID uint) error {
	// В реальном приложении нужно получить список участников чата
	// Пока что отправляем всем клиентам кроме отправителя
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	for client := range h.clients {
		if client.userID != excludeUserID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}

	return nil
}

func (h *Hub) BroadcastMessage(message WSMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- data
	return nil
}

func (h *Hub) GetOnlineUsers() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var userIDs []uint
	for client := range h.clients {
		userIDs = append(userIDs, client.userID)
	}

	return userIDs
}

func (h *Hub) SendNotificationToChat(chatID uint, notification *entities.Notification) {
	// Для отправки уведомлений всем пользователям чата нам не обязательно
	// запрашивать всю информацию о чате, достаточно получить список участников

	// Используем системный userID (0) для получения списка участников
	// Это безопасно, так как для системных сообщений мы хотим получить всех участников
	members, err := h.chatUseCase.GetChatMembers(chatID, 0)
	if err != nil {
		h.logger.Errorf("Failed to get chat members for notification: %v", err)
		return
	}

	// Создаем WebSocket сообщение
	wsMsg := entities.WebSocketMessage{
		Type:         "notification",
		ChatID:       chatID,
		Notification: notification,
	}

	data, err := json.Marshal(wsMsg)
	if err != nil {
		h.logger.Errorf("Failed to marshal notification: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// Отправляем уведомление всем онлайн участникам чата
	for client := range h.clients {
		for _, member := range members {
			if client.userID == member.ID {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
				break
			}
		}
	}
}

func getTimestamp() int64 {
	return getCurrentTimestamp()
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
