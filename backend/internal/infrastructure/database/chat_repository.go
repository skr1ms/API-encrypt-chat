package database

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/repository"

	"gorm.io/gorm"
)

type chatRepository struct {
	db *gorm.DB
}

// NewChatRepository - создает новый экземпляр репозитория чатов
func NewChatRepository(db *gorm.DB) repository.ChatRepository {
	return &chatRepository{db: db}
}

// Create - создает новый чат в базе данных
func (r *chatRepository) Create(chat *entities.Chat) error {
	return r.db.Create(chat).Error
}

// GetByID - получает чат по его ID с загрузкой создателя и участников
func (r *chatRepository) GetByID(id uint) (*entities.Chat, error) {
	var chat entities.Chat
	err := r.db.Preload("Creator").Preload("Members").First(&chat, id).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// GetUserChats - получает все чаты пользователя
func (r *chatRepository) GetUserChats(userID uint) ([]entities.Chat, error) {
	var chats []entities.Chat
	err := r.db.
		Preload("Creator").
		Preload("Members").
		Joins("JOIN chat_members ON chats.id = chat_members.chat_id").
		Where("chat_members.user_id = ?", userID).
		Find(&chats).Error
	return chats, err
}

// Update - обновляет данные чата в базе данных
func (r *chatRepository) Update(chat *entities.Chat) error {
	return r.db.Save(chat).Error
}

// Delete - удаляет чат из базы данных по ID
func (r *chatRepository) Delete(id uint) error {
	return r.db.Delete(&entities.Chat{}, id).Error
}

// AddMember - добавляет участника в чат с указанной ролью
func (r *chatRepository) AddMember(chatID, userID uint, role string) error {
	member := &entities.ChatMember{
		ChatID: chatID,
		UserID: userID,
		Role:   role,
	}
	return r.db.Create(member).Error
}

// RemoveMember - удаляет участника из чата
func (r *chatRepository) RemoveMember(chatID, userID uint) error {
	return r.db.Where("chat_id = ? AND user_id = ?", chatID, userID).Delete(&entities.ChatMember{}).Error
}

// GetMembers - получает список всех участников чата
func (r *chatRepository) GetMembers(chatID uint) ([]entities.User, error) {
	var users []entities.User
	err := r.db.
		Joins("JOIN chat_members ON users.id = chat_members.user_id").
		Where("chat_members.chat_id = ?", chatID).
		Find(&users).Error
	return users, err
}

// IsMember - проверяет, является ли пользователь участником чата
func (r *chatRepository) IsMember(chatID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&entities.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	return count > 0, err
}

// FindPrivateChat - находит приватный чат между двумя пользователями
func (r *chatRepository) FindPrivateChat(userID1, userID2 uint) (*entities.Chat, error) {
	var chat entities.Chat

	err := r.db.
		Preload("Creator").
		Preload("Members").
		Where("is_group = false").
		Joins("JOIN chat_members cm1 ON chats.id = cm1.chat_id AND cm1.user_id = ?", userID1).
		Joins("JOIN chat_members cm2 ON chats.id = cm2.chat_id AND cm2.user_id = ?", userID2).
		First(&chat).Error

	if err != nil {
		return nil, err
	}

	return &chat, nil
}

// GetMembersWithRoles - получает список участников чата с их ролями
func (r *chatRepository) GetMembersWithRoles(chatID uint) ([]*entities.User, error) {
	type userWithRole struct {
		entities.User
		Role string `gorm:"column:role"`
	}

	var usersWithRoles []userWithRole

	err := r.db.Model(&entities.User{}).
		Select("users.*, chat_members.role").
		Joins("JOIN chat_members ON users.id = chat_members.user_id").
		Where("chat_members.chat_id = ?", chatID).
		Scan(&usersWithRoles).Error

	if err != nil {
		return nil, err
	}

	result := make([]*entities.User, len(usersWithRoles))
	for i, ur := range usersWithRoles {
		user := ur.User
		user.Role = ur.Role
		result[i] = &user
	}

	return result, nil
}

// UpdateMemberRole - обновляет роль участника чата
func (r *chatRepository) UpdateMemberRole(chatID, userID uint, role string) error {
	return r.db.Model(&entities.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Update("role", role).
		Error
}

// GetMemberRole - получает роль участника в чате
func (r *chatRepository) GetMemberRole(chatID, userID uint) (string, error) {
	var member entities.ChatMember
	err := r.db.
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&member).Error

	if err != nil {
		return "", err
	}

	return member.Role, nil
}
