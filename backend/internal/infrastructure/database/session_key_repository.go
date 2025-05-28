package database

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/repository"
	"time"

	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository - создает новый экземпляр репозитория сессий
func NewSessionRepository(db *gorm.DB) repository.SessionRepository {
	return &sessionRepository{db: db}
}

// Create - создает новую сессию в базе данных
func (r *sessionRepository) Create(session *entities.Session) error {
	return r.db.Create(session).Error
}

// GetByToken - получает сессию по токену с загрузкой пользователя
func (r *sessionRepository) GetByToken(token string) (*entities.Session, error) {
	var session entities.Session
	err := r.db.Preload("User").Where("token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetUserSessions - получает все сессии пользователя
func (r *sessionRepository) GetUserSessions(userID uint) ([]entities.Session, error) {
	var sessions []entities.Session
	err := r.db.Where("user_id = ?", userID).Find(&sessions).Error
	return sessions, err
}

// Update - обновляет данные сессии в базе данных
func (r *sessionRepository) Update(session *entities.Session) error {
	return r.db.Save(session).Error
}

// Delete - удаляет сессию по токену
func (r *sessionRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&entities.Session{}).Error
}

// DeleteExpired - удаляет все истекшие сессии
func (r *sessionRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&entities.Session{}).Error
}

// UpdateActivity - обновляет время последней активности сессии
func (r *sessionRepository) UpdateActivity(token string, lastActivity time.Time) error {
	return r.db.Model(&entities.Session{}).
		Where("token = ?", token).
		Update("last_activity", lastActivity).Error
}

type keyExchangeRepository struct {
	db *gorm.DB
}

// NewKeyExchangeRepository - создает новый экземпляр репозитория обмена ключами
func NewKeyExchangeRepository(db *gorm.DB) repository.KeyExchangeRepository {
	return &keyExchangeRepository{db: db}
}

// Create - создает новый обмен ключами в базе данных
func (r *keyExchangeRepository) Create(keyExchange *entities.KeyExchange) error {
	return r.db.Create(keyExchange).Error
}

// GetByID - получает обмен ключами по ID с загрузкой пользователей
func (r *keyExchangeRepository) GetByID(id uint) (*entities.KeyExchange, error) {
	var keyExchange entities.KeyExchange
	err := r.db.Preload("UserA").Preload("UserB").First(&keyExchange, id).Error
	if err != nil {
		return nil, err
	}
	return &keyExchange, nil
}

// GetByUsers - получает обмен ключами между двумя пользователями
func (r *keyExchangeRepository) GetByUsers(userAID, userBID uint) (*entities.KeyExchange, error) {
	var keyExchange entities.KeyExchange
	err := r.db.
		Preload("UserA").
		Preload("UserB").
		Where("(user_a_id = ? AND user_b_id = ?) OR (user_a_id = ? AND user_b_id = ?)",
			userAID, userBID, userBID, userAID).
		First(&keyExchange).Error
	if err != nil {
		return nil, err
	}
	return &keyExchange, nil
}

// Update - обновляет данные обмена ключами в базе данных
func (r *keyExchangeRepository) Update(keyExchange *entities.KeyExchange) error {
	return r.db.Save(keyExchange).Error
}

// Delete - удаляет обмен ключами из базы данных по ID
func (r *keyExchangeRepository) Delete(id uint) error {
	return r.db.Delete(&entities.KeyExchange{}, id).Error
}

// GetPendingExchanges - получает все ожидающие обмены ключами для пользователя
func (r *keyExchangeRepository) GetPendingExchanges(userID uint) ([]entities.KeyExchange, error) {
	var keyExchanges []entities.KeyExchange
	err := r.db.
		Preload("UserA").
		Preload("UserB").
		Where("(user_a_id = ? OR user_b_id = ?) AND status = ?", userID, userID, "pending").
		Find(&keyExchanges).Error
	return keyExchanges, err
}
