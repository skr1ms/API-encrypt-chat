package database

import (
	"sleek-chat-backend/internal/domain/entities"
	"sleek-chat-backend/internal/domain/repository"
	"time"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository - создает новый экземпляр репозитория пользователей
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// Create - создает нового пользователя в базе данных
func (r *userRepository) Create(user *entities.User) error {
	return r.db.Create(user).Error
}

// GetByID - получает пользователя по его ID
func (r *userRepository) GetByID(id uint) (*entities.User, error) {
	var user entities.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername - получает пользователя по имени пользователя
func (r *userRepository) GetByUsername(username string) (*entities.User, error) {
	var user entities.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail - получает пользователя по email адресу
func (r *userRepository) GetByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update - обновляет данные пользователя в базе данных
func (r *userRepository) Update(user *entities.User) error {
	return r.db.Save(user).Error
}

// Delete - удаляет пользователя из базы данных по ID
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&entities.User{}, id).Error
}

// UpdateOnlineStatus - обновляет статус пользователя (онлайн/оффлайн)
func (r *userRepository) UpdateOnlineStatus(userID uint, isOnline bool) error {
	updates := map[string]interface{}{
		"is_online": isOnline,
	}

	if !isOnline {
		updates["last_seen"] = time.Now()
	}

	return r.db.Model(&entities.User{}).Where("id = ?", userID).Updates(updates).Error
}

// GetOnlineUsers - получает список всех пользователей в онлайне
func (r *userRepository) GetOnlineUsers() ([]entities.User, error) {
	var users []entities.User
	err := r.db.Where("is_online = ?", true).Find(&users).Error
	return users, err
}

// SearchUsers - ищет пользователей по имени или email с исключением указанного пользователя
func (r *userRepository) SearchUsers(query string, excludeUserID uint, limit int) ([]entities.User, error) {
	var users []entities.User

	searchQuery := r.db.Where("(username ILIKE ? OR email ILIKE ?)", "%"+query+"%", "%"+query+"%")

	if excludeUserID != 0 {
		searchQuery = searchQuery.Where("id != ?", excludeUserID)
	}

	if limit > 0 {
		searchQuery = searchQuery.Limit(limit)
	}

	orderClause := "CASE WHEN username ILIKE '" + query + "%' THEN 1 WHEN email ILIKE '" + query + "%' THEN 2 ELSE 3 END"
	searchQuery = searchQuery.Order(orderClause)

	err := searchQuery.Find(&users).Error
	return users, err
}

// UpdatePassword - обновляет хеш пароля пользователя
func (r *userRepository) UpdatePassword(userID uint, passwordHash string) error {
	return r.db.Model(&entities.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}
