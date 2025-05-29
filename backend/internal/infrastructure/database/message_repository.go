package database

import (
	"sleek-chat-backend/internal/domain/entities"
	"sleek-chat-backend/internal/domain/repository"

	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository - создает новый экземпляр репозитория сообщений
func NewMessageRepository(db *gorm.DB) repository.MessageRepository {
	return &messageRepository{db: db}
}

// Create - создает новое сообщение в базе данных
func (r *messageRepository) Create(message *entities.Message) error {
	return r.db.Create(message).Error
}

// GetByID - получает сообщение по его ID с загрузкой отправителя и чата
func (r *messageRepository) GetByID(id uint) (*entities.Message, error) {
	var message entities.Message
	err := r.db.Preload("Sender").Preload("Chat").First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetChatMessages - получает сообщения чата с пагинацией (отсортированные по дате)
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

// Update - обновляет данные сообщения в базе данных
func (r *messageRepository) Update(message *entities.Message) error {
	return r.db.Save(message).Error
}

// Delete - удаляет сообщение из базы данных по ID
func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&entities.Message{}, id).Error
}

// GetUserMessages - получает все сообщения пользователя с пагинацией
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
