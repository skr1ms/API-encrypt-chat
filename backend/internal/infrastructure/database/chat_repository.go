package database

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/repository"

	"gorm.io/gorm"
)

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) repository.ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Create(chat *entities.Chat) error {
	return r.db.Create(chat).Error
}

func (r *chatRepository) GetByID(id uint) (*entities.Chat, error) {
	var chat entities.Chat
	err := r.db.Preload("Creator").Preload("Members").First(&chat, id).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

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

func (r *chatRepository) Update(chat *entities.Chat) error {
	return r.db.Save(chat).Error
}

func (r *chatRepository) Delete(id uint) error {
	return r.db.Delete(&entities.Chat{}, id).Error
}

func (r *chatRepository) AddMember(chatID, userID uint, role string) error {
	member := &entities.ChatMember{
		ChatID: chatID,
		UserID: userID,
		Role:   role,
	}
	return r.db.Create(member).Error
}

func (r *chatRepository) RemoveMember(chatID, userID uint) error {
	return r.db.Where("chat_id = ? AND user_id = ?", chatID, userID).Delete(&entities.ChatMember{}).Error
}

func (r *chatRepository) GetMembers(chatID uint) ([]entities.User, error) {
	var users []entities.User
	err := r.db.
		Joins("JOIN chat_members ON users.id = chat_members.user_id").
		Where("chat_members.chat_id = ?", chatID).
		Find(&users).Error
	return users, err
}

func (r *chatRepository) IsMember(chatID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&entities.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *chatRepository) FindPrivateChat(userID1, userID2 uint) (*entities.Chat, error) {
	var chat entities.Chat

	// Ищем приватный чат между двумя пользователями
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
