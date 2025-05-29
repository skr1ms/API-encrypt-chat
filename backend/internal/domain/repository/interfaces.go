package repository

import (
	"sleek-chat-backend/internal/domain/entities"
	"time"
)

type UserRepository interface {
	Create(user *entities.User) error
	GetByID(id uint) (*entities.User, error)
	GetByUsername(username string) (*entities.User, error)
	GetByEmail(email string) (*entities.User, error)
	Update(user *entities.User) error
	Delete(id uint) error
	UpdateOnlineStatus(userID uint, isOnline bool) error
	UpdatePassword(userID uint, passwordHash string) error
	GetOnlineUsers() ([]entities.User, error)
	SearchUsers(query string, excludeUserID uint, limit int) ([]entities.User, error)
}

type ChatRepository interface {
	Create(chat *entities.Chat) error
	GetByID(id uint) (*entities.Chat, error)
	GetUserChats(userID uint) ([]entities.Chat, error)
	Update(chat *entities.Chat) error
	Delete(id uint) error
	AddMember(chatID, userID uint, role string) error
	RemoveMember(chatID, userID uint) error
	GetMembers(chatID uint) ([]entities.User, error)
	GetMembersWithRoles(chatID uint) ([]*entities.User, error)
	IsMember(chatID, userID uint) (bool, error)
	FindPrivateChat(userID1, userID2 uint) (*entities.Chat, error)
	UpdateMemberRole(chatID, userID uint, role string) error
	GetMemberRole(chatID, userID uint) (string, error)
}

type MessageRepository interface {
	Create(message *entities.Message) error
	GetByID(id uint) (*entities.Message, error)
	GetChatMessages(chatID uint, limit, offset int) ([]entities.Message, error)
	Update(message *entities.Message) error
	Delete(id uint) error
	GetUserMessages(userID uint, limit, offset int) ([]entities.Message, error)
}

type KeyExchangeRepository interface {
	Create(keyExchange *entities.KeyExchange) error
	GetByID(id uint) (*entities.KeyExchange, error)
	GetByUsers(userAID, userBID uint) (*entities.KeyExchange, error)
	Update(keyExchange *entities.KeyExchange) error
	Delete(id uint) error
	DeleteByUsers(userAID, userBID uint) error
	GetActiveExchanges(userID uint) ([]entities.KeyExchange, error)
	GetPendingExchanges(userID uint) ([]entities.KeyExchange, error)
	UpdateStatus(id uint, status string) error
}

type SessionRepository interface {
	Create(session *entities.Session) error
	GetByToken(token string) (*entities.Session, error)
	GetUserSessions(userID uint) ([]entities.Session, error)
	Update(session *entities.Session) error
	Delete(token string) error
	DeleteExpired() error
	UpdateActivity(token string, lastActivity time.Time) error
}

type Repository struct {
	User        UserRepository
	Chat        ChatRepository
	Message     MessageRepository
	KeyExchange KeyExchangeRepository
	Session     SessionRepository
}
