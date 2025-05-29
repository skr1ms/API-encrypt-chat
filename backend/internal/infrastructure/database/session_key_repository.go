package database

import (
	"sleek-chat-backend/internal/domain/entities"
	"sleek-chat-backend/internal/domain/repository"
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
