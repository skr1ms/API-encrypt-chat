package database

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/repository"

	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) repository.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *entities.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) GetByID(id uint) (*entities.Message, error) {
	var message entities.Message
	err := r.db.Preload("Sender").Preload("Chat").First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetChatMessages(chatID uint, limit, offset int) ([]entities.Message, error) {
	var messages []entities.Message
	err := r.db.
		Preload("Sender").
		Where("chat_id = ?", chatID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) Update(message *entities.Message) error {
	return r.db.Save(message).Error
}

func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&entities.Message{}, id).Error
}

func (r *messageRepository) GetUserMessages(userID uint, limit, offset int) ([]entities.Message, error) {
	var messages []entities.Message
	err := r.db.
		Preload("Sender").
		Preload("Chat").
		Where("sender_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}
