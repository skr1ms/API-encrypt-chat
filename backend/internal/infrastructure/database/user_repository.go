package database

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/repository"
	"time"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *entities.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*entities.User, error) {
	var user entities.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*entities.User, error) {
	var user entities.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *entities.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&entities.User{}, id).Error
}

func (r *userRepository) UpdateOnlineStatus(userID uint, isOnline bool) error {
	updates := map[string]interface{}{
		"is_online": isOnline,
	}

	if !isOnline {
		updates["last_seen"] = time.Now()
	}

	return r.db.Model(&entities.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *userRepository) GetOnlineUsers() ([]entities.User, error) {
	var users []entities.User
	err := r.db.Where("is_online = ?", true).Find(&users).Error
	return users, err
}

func (r *userRepository) SearchUsers(query string, excludeUserID uint, limit int) ([]entities.User, error) {
	var users []entities.User

	// Строим запрос для поиска по username или email
	searchQuery := r.db.Where("(username ILIKE ? OR email ILIKE ?)", "%"+query+"%", "%"+query+"%")

	// Исключаем текущего пользователя из результатов
	if excludeUserID != 0 {
		searchQuery = searchQuery.Where("id != ?", excludeUserID)
	}

	// Ограничиваем количество результатов
	if limit > 0 {
		searchQuery = searchQuery.Limit(limit)
	}

	// Сортируем по релевантности: сначала точные совпадения username, потом email
	orderClause := "CASE WHEN username ILIKE '" + query + "%' THEN 1 WHEN email ILIKE '" + query + "%' THEN 2 ELSE 3 END"
	searchQuery = searchQuery.Order(orderClause)

	err := searchQuery.Find(&users).Error
	return users, err
}
