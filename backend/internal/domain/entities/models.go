package entities

import (
	"time"

	"gorm.io/gorm"
)

// User представляет пользователя в системе
type User struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	Username       string `gorm:"unique;not null" json:"username"`
	Email          string `gorm:"unique;not null" json:"email"`
	PasswordHash   string `gorm:"not null" json:"-"`
	ECDSAPublicKey string `gorm:"type:text" json:"ecdsa_public_key"`
	RSAPublicKey   string `gorm:"type:text" json:"rsa_public_key"`
	// ВНИМАНИЕ: Хранение приватных ключей на сервере небезопасно в продакшене
	// В реальном приложении ключи должны храниться только на клиенте
	ECDSAPrivateKey string         `gorm:"type:text" json:"-"`
	RSAPrivateKey   string         `gorm:"type:text" json:"-"`
	IsOnline        bool           `gorm:"default:false" json:"is_online"`
	Role            string         `gorm:"-" json:"role,omitempty"` // Используется для представления роли пользователя в чате
	LastSeen        *time.Time     `json:"last_seen"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Chat представляет чат между пользователями
type Chat struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	IsGroup   bool           `gorm:"default:false" json:"is_group"`
	CreatedBy uint           `gorm:"not null" json:"created_by"`
	Creator   User           `gorm:"foreignKey:CreatedBy" json:"creator"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Связи
	Members  []User    `gorm:"many2many:chat_members;" json:"members"`
	Messages []Message `gorm:"foreignKey:ChatID" json:"messages"`
}

// Message представляет сообщение в чате
type Message struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ChatID      uint   `gorm:"not null" json:"chat_id"`
	Chat        Chat   `gorm:"foreignKey:ChatID" json:"chat"`
	SenderID    uint   `gorm:"not null" json:"sender_id"`
	Sender      User   `gorm:"foreignKey:SenderID" json:"sender"`
	Content     string `gorm:"type:text" json:"content"` // Зашифрованный контент
	MessageType string `gorm:"default:'text'" json:"message_type"`

	// Криптографические поля
	Nonce          string `gorm:"type:text" json:"nonce"`
	IV             string `gorm:"type:text" json:"iv"`
	HMAC           string `gorm:"type:text" json:"hmac"`
	ECDSASignature string `gorm:"type:text" json:"ecdsa_signature"`
	RSASignature   string `gorm:"type:text" json:"rsa_signature"`

	IsEdited  bool           `gorm:"default:false" json:"is_edited"`
	EditedAt  *time.Time     `json:"edited_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ChatMember представляет участника чата
type ChatMember struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	ChatID   uint      `gorm:"not null" json:"chat_id"`
	UserID   uint      `gorm:"not null" json:"user_id"`
	Role     string    `gorm:"default:'member'" json:"role"` // admin, member
	JoinedAt time.Time `json:"joined_at"`

	// Уникальный индекс для пары chat_id, user_id
	Chat User `gorm:"foreignKey:ChatID" json:"-"`
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// KeyExchange представляет обмен ключами между пользователями
type KeyExchange struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserAID          uint      `gorm:"not null" json:"user_a_id"`
	UserBID          uint      `gorm:"not null" json:"user_b_id"`
	UserA            User      `gorm:"foreignKey:UserAID" json:"user_a"`
	UserB            User      `gorm:"foreignKey:UserBID" json:"user_b"`
	SharedSecretHash string    `gorm:"type:text" json:"-"`              // Хэш общего секрета для проверки
	Status           string    `gorm:"default:'pending'" json:"status"` // pending, completed, failed
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Session представляет активную сессию пользователя
type Session struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"user"`
	Token        string    `gorm:"unique;not null" json:"token"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// Notification представляет системные уведомления для WebSocket
type Notification struct {
	Type    string                 `json:"type"` // "user_left", "group_deleted", "user_added", etc.
	ChatID  uint                   `json:"chat_id"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// WebSocketMessage представляет сообщение для WebSocket
type WebSocketMessage struct {
	Type         string        `json:"type"` // "message", "notification", "typing", etc.
	ChatID       uint          `json:"chat_id,omitempty"`
	Message      *Message      `json:"message,omitempty"`
	Notification *Notification `json:"notification,omitempty"`
}

// TableName методы для явного указания имен таблиц
func (User) TableName() string        { return "users" }
func (Chat) TableName() string        { return "chats" }
func (Message) TableName() string     { return "messages" }
func (ChatMember) TableName() string  { return "chat_members" }
func (KeyExchange) TableName() string { return "key_exchanges" }
func (Session) TableName() string     { return "sessions" }
