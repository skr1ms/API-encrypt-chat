package usecase

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/repository"
	"errors"
	"strings"
)

type UserUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase - создает новый экземпляр сервиса для работы с пользователями
func NewUserUseCase(userRepo repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
	}
}

type SearchUsersRequest struct {
	Query  string `json:"query" binding:"required,min=1"`
	Limit  int    `json:"limit"`
	UserID uint   `json:"-"`
}

type SearchUsersResponse struct {
	Users []UserSearchResult `json:"users"`
	Total int                `json:"total"`
}

type UserSearchResult struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsOnline bool   `json:"is_online"`
}

// SearchUsers - осуществляет поиск пользователей по запросу
func (uc *UserUseCase) SearchUsers(req SearchUsersRequest) (*SearchUsersResponse, error) {
	if strings.TrimSpace(req.Query) == "" {
		return nil, errors.New("поисковый запрос не может быть пустым")
	}

	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 10
	}

	query := strings.TrimSpace(req.Query)

	users, err := uc.userRepo.SearchUsers(query, req.UserID, req.Limit)
	if err != nil {
		return nil, err
	}

	searchResults := make([]UserSearchResult, 0, len(users))
	for _, user := range users {
		searchResults = append(searchResults, UserSearchResult{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsOnline: user.IsOnline,
		})
	}

	return &SearchUsersResponse{
		Users: searchResults,
		Total: len(searchResults),
	}, nil
}

// GetUserByID - получает данные пользователя по его идентификатору
func (uc *UserUseCase) GetUserByID(userID uint) (*entities.User, error) {
	return uc.userRepo.GetByID(userID)
}

// GetUserByUsername - получает данные пользователя по имени пользователя
func (uc *UserUseCase) GetUserByUsername(username string) (*entities.User, error) {
	return uc.userRepo.GetByUsername(username)
}

// GetOnlineUsers - получает список всех пользователей, находящихся в сети
func (uc *UserUseCase) GetOnlineUsers() ([]entities.User, error) {
	return uc.userRepo.GetOnlineUsers()
}
