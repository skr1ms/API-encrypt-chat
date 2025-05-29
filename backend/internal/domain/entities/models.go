package entities

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Username        string         `gorm:"unique;not null" json:"username"`
	Email           string         `gorm:"unique;not null" json:"email"`
	PasswordHash    string         `gorm:"not null" json:"-"`
	ECDSAPublicKey  string         `gorm:"type:text" json:"ecdsa_public_key"`
	RSAPublicKey    string         `gorm:"type:text" json:"rsa_public_key"`
	ECDSAPrivateKey string         `gorm:"type:text" json:"-"`
	RSAPrivateKey   string         `gorm:"type:text" json:"-"`
	IsOnline        bool           `gorm:"default:false" json:"is_online"`
	Role            string         `gorm:"-" json:"role,omitempty"`
	LastSeen        *time.Time     `json:"last_seen"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type Chat struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	IsGroup   bool           `gorm:"default:false" json:"is_group"`
	CreatedBy uint           `gorm:"not null" json:"created_by"`
	Creator   User           `gorm:"foreignKey:CreatedBy" json:"creator"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Members   []User         `gorm:"many2many:chat_members;" json:"members"`
	Messages  []Message      `gorm:"foreignKey:ChatID" json:"messages"`
}

type Message struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	ChatID         uint   `gorm:"not null" json:"chat_id"`
	Chat           Chat   `gorm:"foreignKey:ChatID" json:"chat"`
	SenderID       uint   `gorm:"not null" json:"sender_id"`
	Sender         User   `gorm:"foreignKey:SenderID" json:"sender"`
	Content        string `gorm:"type:text" json:"content"`
	MessageType    string `gorm:"default:'text'" json:"message_type"`
	Timestamp      *int64 `gorm:"default:null" json:"timestamp"`
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

type ChatMember struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	ChatID   uint      `gorm:"not null" json:"chat_id"`
	UserID   uint      `gorm:"not null" json:"user_id"`
	Role     string    `gorm:"default:'member'" json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	Chat     User      `gorm:"foreignKey:ChatID" json:"-"`
	User     User      `gorm:"foreignKey:UserID" json:"-"`
}

type KeyExchange struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserAID          uint      `gorm:"not null" json:"user_a_id"`
	UserBID          uint      `gorm:"not null" json:"user_b_id"`
	UserA            User      `gorm:"foreignKey:UserAID" json:"user_a"`
	UserB            User      `gorm:"foreignKey:UserBID" json:"user_b"`
	SharedSecretHash string    `gorm:"type:text" json:"-"`
	Status           string    `gorm:"default:'pending'" json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Session struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"user"`
	Token        string    `gorm:"unique;not null" json:"token"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastActivity time.Time `json:"last_activity"`
}

type Notification struct {
	Type    string                 `json:"type"`
	ChatID  uint                   `json:"chat_id"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type WebSocketMessage struct {
	Type         string        `json:"type"`
	ChatID       uint          `json:"chat_id,omitempty"`
	Message      *Message      `json:"message,omitempty"`
	Notification *Notification `json:"notification,omitempty"`
}

// TableName - возвращает имя таблицы для пользователей
func (User) TableName() string { return "users" }

// TableName - возвращает имя таблицы для чатов
func (Chat) TableName() string { return "chats" }

// TableName - возвращает имя таблицы для сообщений
func (Message) TableName() string { return "messages" }

// TableName - возвращает имя таблицы для участников чата
func (ChatMember) TableName() string { return "chat_members" }

// TableName - возвращает имя таблицы для обмена ключами
func (KeyExchange) TableName() string { return "key_exchanges" }

// TableName - возвращает имя таблицы для сессий
func (Session) TableName() string { return "sessions" }
