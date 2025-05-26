package usecase

import (
	"crypto-chat-backend/internal/crypto"
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/internal/domain/repository"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
)

type ChatUseCase struct {
	chatRepo        repository.ChatRepository
	messageRepo     repository.MessageRepository
	userRepo        repository.UserRepository
	keyExchangeRepo repository.KeyExchangeRepository
}

func NewChatUseCase(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	keyExchangeRepo repository.KeyExchangeRepository,
) *ChatUseCase {
	return &ChatUseCase{
		chatRepo:        chatRepo,
		messageRepo:     messageRepo,
		userRepo:        userRepo,
		keyExchangeRepo: keyExchangeRepo,
	}
}

type CreateChatRequest struct {
	Name      string `json:"name" binding:"required"`
	IsGroup   bool   `json:"is_group"`
	MemberIDs []uint `json:"member_ids" binding:"required"`
}

type SendMessageRequest struct {
	Content     string `json:"content" binding:"required"`
	MessageType string `json:"message_type"`
}

type MessageResponse struct {
	*entities.Message
	DecryptedContent string `json:"decrypted_content,omitempty"`
}

func (uc *ChatUseCase) CreateChat(creatorID uint, req *CreateChatRequest) (*entities.Chat, error) {
	// Проверяем, что создатель существует
	creator, err := uc.userRepo.GetByID(creatorID)
	if err != nil {
		return nil, errors.New("creator not found")
	}

	// Создаем чат
	chat := &entities.Chat{
		Name:      req.Name,
		IsGroup:   req.IsGroup,
		CreatedBy: creatorID,
		Creator:   *creator,
	}

	if err := uc.chatRepo.Create(chat); err != nil {
		return nil, fmt.Errorf("failed to create chat: %v", err)
	}

	// Добавляем создателя как администратора
	if err := uc.chatRepo.AddMember(chat.ID, creatorID, "admin"); err != nil {
		return nil, fmt.Errorf("failed to add creator to chat: %v", err)
	}

	// Добавляем остальных участников
	for _, memberID := range req.MemberIDs {
		if memberID != creatorID {
			if err := uc.chatRepo.AddMember(chat.ID, memberID, "member"); err != nil {
				return nil, fmt.Errorf("failed to add member %d to chat: %v", memberID, err)
			}
		}
	}

	return chat, nil
}

func (uc *ChatUseCase) GetUserChats(userID uint) ([]entities.Chat, error) {
	return uc.chatRepo.GetUserChats(userID)
}

func (uc *ChatUseCase) SendMessage(chatID, senderID uint, req *SendMessageRequest, senderECDSAPrivateKey *ecdsa.PrivateKey, senderRSAPrivateKey *rsa.PrivateKey) (*entities.Message, error) {
	// Проверяем, что отправитель является участником чата
	isMember, err := uc.chatRepo.IsMember(chatID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("sender is not a member of the chat")
	}

	// Получаем участников чата для шифрования сообщения
	members, err := uc.chatRepo.GetMembers(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat members: %v", err)
	}

	// Получаем отправителя
	sender, err := uc.userRepo.GetByID(senderID)
	if err != nil {
		return nil, errors.New("sender not found")
	}

	// Для простоты, будем использовать общий секрет на основе первого участника
	// В реальном приложении нужно было бы шифровать сообщение для каждого участника отдельно
	var sharedSecret []byte
	if len(members) > 1 {
		// Находим первого участника, который не является отправителем
		var recipientPublicKey []byte
		for _, member := range members {
			if member.ID != senderID {
				recipientPublicKey, err = hex.DecodeString(member.ECDSAPublicKey)
				if err != nil {
					return nil, fmt.Errorf("failed to decode recipient public key: %v", err)
				}
				break
			}
		}

		if len(recipientPublicKey) > 0 {
			sharedSecret, err = crypto.ComputeECDHSharedSecret(senderECDSAPrivateKey, recipientPublicKey)
			if err != nil {
				return nil, fmt.Errorf("failed to compute shared secret: %v", err)
			}
		}
	}

	// Если общий секрет не создан, используем заглушку (не рекомендуется в продакшене)
	if len(sharedSecret) == 0 {
		sharedSecret = make([]byte, 64) // 32 для AES + 32 для HMAC
		copy(sharedSecret, "default-shared-secret-for-single-user-or-error")
	}

	// Создаем защищенное сообщение
	secureMsg, err := crypto.CreateSecureMessage(
		fmt.Sprintf("%d", senderID),
		fmt.Sprintf("%d", chatID),
		[]byte(req.Content),
		sharedSecret,
		senderECDSAPrivateKey,
		senderRSAPrivateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure message: %v", err)
	}

	// Создаем сообщение в базе данных
	message := &entities.Message{
		ChatID:         chatID,
		SenderID:       senderID,
		Content:        secureMsg.Ciphertext,
		MessageType:    req.MessageType,
		Nonce:          secureMsg.Nonce,
		IV:             secureMsg.IV,
		HMAC:           secureMsg.HMAC,
		ECDSASignature: secureMsg.ECDSASignature,
		RSASignature:   secureMsg.RSASignature,
	}

	if message.MessageType == "" {
		message.MessageType = "text"
	}

	if err := uc.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("failed to save message: %v", err)
	}

	// Загружаем связанные данные
	message.Sender = *sender
	chat, _ := uc.chatRepo.GetByID(chatID)
	if chat != nil {
		message.Chat = *chat
	}

	return message, nil
}

func (uc *ChatUseCase) GetChatMessages(chatID, userID uint, limit, offset int) ([]MessageResponse, error) {
	// Проверяем, что пользователь является участником чата
	isMember, err := uc.chatRepo.IsMember(chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not a member of the chat")
	}

	messages, err := uc.messageRepo.GetChatMessages(chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []MessageResponse
	for _, msg := range messages {
		response := MessageResponse{
			Message: &msg,
		}

		// Здесь можно добавить логику расшифровки сообщения
		// Для этого нужны приватные ключи пользователя и публичные ключи отправителя
		// response.DecryptedContent = decryptedContent

		responses = append(responses, response)
	}

	return responses, nil
}

func (uc *ChatUseCase) AddMember(chatID, adminID, newMemberID uint) error {
	// Проверяем, что admin является администратором чата
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	if chat.CreatedBy != adminID {
		return errors.New("only chat creator can add members")
	}

	return uc.chatRepo.AddMember(chatID, newMemberID, "member")
}

func (uc *ChatUseCase) RemoveMember(chatID, adminID, memberID uint) error {
	// Проверяем, что admin является администратором чата
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	if chat.CreatedBy != adminID {
		return errors.New("only chat creator can remove members")
	}

	return uc.chatRepo.RemoveMember(chatID, memberID)
}
