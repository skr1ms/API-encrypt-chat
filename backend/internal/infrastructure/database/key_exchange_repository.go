package database

import (
	"sleek-chat-backend/internal/domain/entities"
	"sleek-chat-backend/internal/domain/repository"

	"gorm.io/gorm"
)

type keyExchangeRepository struct {
	db *gorm.DB
}

// NewKeyExchangeRepository создает новый экземпляр репозитория обмена ключами
func NewKeyExchangeRepository(db *gorm.DB) repository.KeyExchangeRepository {
	return &keyExchangeRepository{db: db}
}

// Create создает новую запись обмена ключами в базе данных
func (r *keyExchangeRepository) Create(keyExchange *entities.KeyExchange) error {
	return r.db.Create(keyExchange).Error
}

// GetByID получает запись обмена ключами по ID
func (r *keyExchangeRepository) GetByID(id uint) (*entities.KeyExchange, error) {
	var keyExchange entities.KeyExchange
	err := r.db.Preload("UserA").Preload("UserB").First(&keyExchange, id).Error
	if err != nil {
		return nil, err
	}
	return &keyExchange, nil
}

// GetByUsers получает запись обмена ключами между двумя пользователями
func (r *keyExchangeRepository) GetByUsers(userAID, userBID uint) (*entities.KeyExchange, error) {
	var keyExchange entities.KeyExchange

	err := r.db.Preload("UserA").Preload("UserB").
		Where("(user_a_id = ? AND user_b_id = ?) OR (user_a_id = ? AND user_b_id = ?)",
			userAID, userBID, userBID, userAID).
		First(&keyExchange).Error

	if err != nil {
		return nil, err
	}

	return &keyExchange, nil
}

// Update обновляет данные обмена ключами в базе данных
func (r *keyExchangeRepository) Update(keyExchange *entities.KeyExchange) error {
	return r.db.Save(keyExchange).Error
}

// Delete удаляет запись обмена ключами по ID
func (r *keyExchangeRepository) Delete(id uint) error {
	return r.db.Delete(&entities.KeyExchange{}, id).Error
}

// DeleteByUsers удаляет запись обмена ключами между пользователями
func (r *keyExchangeRepository) DeleteByUsers(userAID, userBID uint) error {
	return r.db.Where("(user_a_id = ? AND user_b_id = ?) OR (user_a_id = ? AND user_b_id = ?)",
		userAID, userBID, userBID, userAID).
		Delete(&entities.KeyExchange{}).Error
}

// GetActiveExchanges получает все активные обмены ключами для пользователя
func (r *keyExchangeRepository) GetActiveExchanges(userID uint) ([]entities.KeyExchange, error) {
	var exchanges []entities.KeyExchange

	err := r.db.Preload("UserA").Preload("UserB").
		Where("(user_a_id = ? OR user_b_id = ?) AND status = ?",
			userID, userID, "active").
		Find(&exchanges).Error

	return exchanges, err
}

// UpdateStatus обновляет статус обмена ключами
func (r *keyExchangeRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&entities.KeyExchange{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// GetPendingExchanges получает все ожидающие обмены ключами для пользователя
func (r *keyExchangeRepository) GetPendingExchanges(userID uint) ([]entities.KeyExchange, error) {
	var exchanges []entities.KeyExchange

	err := r.db.Preload("UserA").Preload("UserB").
		Where("(user_a_id = ? OR user_b_id = ?) AND status = ?",
			userID, userID, "pending").
		Find(&exchanges).Error

	return exchanges, err
}
