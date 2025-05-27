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

// NotificationSender интерфейс для отправки уведомлений через WebSocket
type NotificationSender interface {
	SendNotificationToChat(chatID uint, notification *entities.Notification)
}

type ChatUseCase struct {
	chatRepo           repository.ChatRepository
	messageRepo        repository.MessageRepository
	userRepo           repository.UserRepository
	keyExchangeRepo    repository.KeyExchangeRepository
	notificationSender NotificationSender
}

func NewChatUseCase(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	keyExchangeRepo repository.KeyExchangeRepository,
	notificationSender NotificationSender,
) *ChatUseCase {
	return &ChatUseCase{
		chatRepo:           chatRepo,
		messageRepo:        messageRepo,
		userRepo:           userRepo,
		keyExchangeRepo:    keyExchangeRepo,
		notificationSender: notificationSender,
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

type PrivateChatResponse struct {
	Chat    *entities.Chat `json:"chat"`
	Created bool           `json:"created"`
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

	// Отправляем уведомление о создании чата, если это групповой чат
	if req.IsGroup && uc.notificationSender != nil {
		notification := &entities.Notification{
			Type:    "group_created",
			ChatID:  chat.ID,
			Message: fmt.Sprintf("Группа \"%s\" была создана пользователем %s", chat.Name, creator.Username),
			Data: map[string]interface{}{
				"creator_id":   creatorID,
				"creator_name": creator.Username,
				"chat_name":    chat.Name,
			},
		}
		uc.notificationSender.SendNotificationToChat(chat.ID, notification)
	}

	return chat, nil
}

func (uc *ChatUseCase) GetUserChats(userID uint) ([]entities.Chat, error) {
	chats, err := uc.chatRepo.GetUserChats(userID)
	if err != nil {
		return nil, err
	}

	// Для приватных чатов нужно адаптировать название под текущего пользователя
	for i := range chats {
		if !chats[i].IsGroup {
			// Получаем участников приватного чата
			members, err := uc.chatRepo.GetMembers(chats[i].ID)
			if err != nil {
				continue // Пропускаем чат если не удалось получить участников
			}

			// Находим собеседника (не текущего пользователя)
			for _, member := range members {
				if member.ID != userID {
					chats[i].Name = fmt.Sprintf("Chat with %s", member.Username)
					break
				}
			}
		}
	}

	return chats, nil
}

func (uc *ChatUseCase) CreateOrGetPrivateChat(userID1, userID2 uint, otherUserName string) (*PrivateChatResponse, error) {
	// Сначала пытаемся найти существующий приватный чат
	existingChat, err := uc.chatRepo.FindPrivateChat(userID1, userID2)
	if err == nil {
		// Чат найден, адаптируем его название для текущего пользователя
		members, err := uc.chatRepo.GetMembers(existingChat.ID)
		if err == nil {
			for _, member := range members {
				if member.ID != userID1 {
					existingChat.Name = fmt.Sprintf("Chat with %s", member.Username)
					break
				}
			}
		}

		return &PrivateChatResponse{
			Chat:    existingChat,
			Created: false,
		}, nil
	}

	// Чат не найден, создаем новый с нейтральным названием
	chatName := "Private Chat" // Используем нейтральное название
	req := &CreateChatRequest{
		Name:      chatName,
		IsGroup:   false,
		MemberIDs: []uint{userID2}, // Добавляем только второго пользователя, создатель добавится автоматически
	}

	newChat, err := uc.CreateChat(userID1, req)
	if err != nil {
		return nil, err
	}

	// Адаптируем название для создателя чата
	newChat.Name = fmt.Sprintf("Chat with %s", otherUserName)

	return &PrivateChatResponse{
		Chat:    newChat,
		Created: true,
	}, nil
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

	// Только вычисляем общий секрет если есть приватный ключ отправителя
	if senderECDSAPrivateKey != nil && len(members) > 1 {
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
			err = nil
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
	// Получаем пользователя для расшифровки
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	fmt.Printf("DEBUG: About to decrypt %d messages for user %d\n", len(messages), userID)

	var responses []MessageResponse
	for _, msg := range messages {
		response := MessageResponse{
			Message: &msg,
		}
		// Расшифровываем сообщение
		decryptedContent, err := uc.decryptMessage(&msg, user)
		if err != nil {
			// Если не удалось расшифровать, возвращаем оригинальный контент как есть
			// Это может происходить для старых сообщений или при ошибках ключей
			fmt.Printf("Failed to decrypt message %d: %v\n", msg.ID, err)
			response.DecryptedContent = msg.Content
		} else {
			fmt.Printf("Successfully decrypted message %d: %s\n", msg.ID, decryptedContent)
			response.DecryptedContent = decryptedContent
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// decryptMessage расшифровывает сообщение для пользователя
func (uc *ChatUseCase) decryptMessage(msg *entities.Message, user *entities.User) (string, error) {
	// Проверяем, есть ли зашифрованные данные
	if msg.Content == "" || msg.IV == "" || msg.HMAC == "" {
		return msg.Content, nil // Возвращаем как есть, если нет данных для расшифровки
	}

	// Получаем отправителя сообщения
	sender, err := uc.userRepo.GetByID(msg.SenderID)
	if err != nil {
		return "", fmt.Errorf("sender not found: %v", err)
	}

	// Парсим приватный ключ пользователя
	userECDSAPrivateKey, err := crypto.DeserializeECDSAPrivateKey([]byte(user.ECDSAPrivateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse user ECDSA private key: %v", err)
	}

	// Парсим публичный ключ отправителя
	senderECDSAPublicKeyBytes, err := hex.DecodeString(sender.ECDSAPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode sender ECDSA public key: %v", err)
	}

	senderRSAPublicKeyBytes, err := hex.DecodeString(sender.RSAPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode sender RSA public key: %v", err)
	}

	// Вычисляем общий секрет
	sharedSecret, err := crypto.ComputeECDHSharedSecret(userECDSAPrivateKey, senderECDSAPublicKeyBytes)
	if err != nil {
		// Если не удалось вычислить общий секрет, используем заглушку
		sharedSecret = make([]byte, 64)
		copy(sharedSecret, "default-shared-secret-for-single-user-or-error")
	}

	// Создаем объект SecureMessage из данных в базе
	secureMsg := &crypto.SecureMessage{
		Ciphertext:     msg.Content,
		IV:             msg.IV,
		HMAC:           msg.HMAC,
		ECDSASignature: msg.ECDSASignature,
		RSASignature:   msg.RSASignature,
		Nonce:          msg.Nonce,
		Timestamp:      msg.CreatedAt.Unix(),
		SenderID:       fmt.Sprintf("%d", msg.SenderID),
		RecipientID:    fmt.Sprintf("%d", user.ID),
	}

	// Расшифровываем сообщение
	plaintext, err := crypto.VerifyAndDecryptMessage(secureMsg, sharedSecret, senderECDSAPublicKeyBytes, senderRSAPublicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt message: %v", err)
	}

	return string(plaintext), nil
}

func (uc *ChatUseCase) AddMember(chatID, requesterID, newMemberID uint) error {
	// Проверяем, что requester является участником чата
	isMember, err := uc.chatRepo.IsMember(chatID, requesterID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("you are not a member of this chat")
	}

	// Проверяем, что новый пользователь ещё не является участником
	isAlreadyMember, err := uc.chatRepo.IsMember(chatID, newMemberID)
	if err != nil {
		return err
	}
	if isAlreadyMember {
		return errors.New("user is already a member of this chat")
	}

	// Добавляем участника в чат
	err = uc.chatRepo.AddMember(chatID, newMemberID, "member")
	if err != nil {
		return err
	}
	// Получаем информацию о добавленном пользователе для уведомления
	newUser, err := uc.userRepo.GetByID(newMemberID)
	if err != nil {
		// Не возвращаем ошибку, так как пользователь уже добавлен
		return nil
	}

	// Создаем системное сообщение в базе данных и отправляем уведомление ПОСЛЕ добавления пользователя
	systemMessageText := fmt.Sprintf("%s присоединился к группе", newUser.Username)

	// Создаем системное сообщение в базе данных
	err = uc.createSystemMessage(chatID, systemMessageText)
	if err != nil {
		// Логируем ошибку, но не останавливаем процесс
	}

	// Отправляем WebSocket уведомление всем участникам чата
	if uc.notificationSender != nil {
		notification := &entities.Notification{
			Type:    "user_joined",
			Message: systemMessageText,
			Data: map[string]interface{}{
				"user_id":  newMemberID,
				"username": newUser.Username,
				"chat_id":  chatID,
			},
		}

		uc.notificationSender.SendNotificationToChat(chatID, notification)
	}

	return nil
}

// AddMemberWithUserData добавляет участника в чат и возвращает данные о добавленном пользователе
func (uc *ChatUseCase) AddMemberWithUserData(chatID, requesterID, newMemberID uint) (*entities.User, error) {
	// Проверяем, что requester является участником чата
	isMember, err := uc.chatRepo.IsMember(chatID, requesterID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("you are not a member of this chat")
	}

	// Проверяем, что новый пользователь ещё не является участником
	isAlreadyMember, err := uc.chatRepo.IsMember(chatID, newMemberID)
	if err != nil {
		return nil, err
	}
	if isAlreadyMember {
		return nil, errors.New("user is already a member of this chat")
	}

	// Добавляем участника в чат
	err = uc.chatRepo.AddMember(chatID, newMemberID, "member")
	if err != nil {
		return nil, err
	}
	// Получаем информацию о добавленном пользователе для уведомления
	newUser, err := uc.userRepo.GetByID(newMemberID)
	if err != nil {
		return nil, err
	}

	// Создаем системное сообщение в базе данных и отправляем уведомление ПОСЛЕ добавления пользователя
	systemMessageText := fmt.Sprintf("%s присоединился к группе", newUser.Username)

	// Создаем системное сообщение в базе данных
	err = uc.createSystemMessage(chatID, systemMessageText)
	if err != nil {
		// Логируем ошибку, но не останавливаем процесс
	}

	// Отправляем WebSocket уведомление всем участникам чата
	if uc.notificationSender != nil {
		notification := &entities.Notification{
			Type:    "user_joined",
			Message: systemMessageText,
			Data: map[string]interface{}{
				"user_id":  newMemberID,
				"username": newUser.Username,
				"chat_id":  chatID,
			},
		}

		uc.notificationSender.SendNotificationToChat(chatID, notification)
	}

	return newUser, nil
}

func (uc *ChatUseCase) RemoveMember(chatID, actorID, memberID uint) error {
	// Проверяем, что оба пользователя являются участниками чата
	isMemberActor, err := uc.chatRepo.IsMember(chatID, actorID)
	if err != nil {
		return err
	}
	if !isMemberActor {
		return errors.New("you are not a member of this chat")
	}

	isMemberTarget, err := uc.chatRepo.IsMember(chatID, memberID)
	if err != nil {
		return err
	}
	if !isMemberTarget {
		return errors.New("target user is not a member of this chat")
	}

	// Получаем информацию о чате
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	// Получаем роль актора и целевого пользователя
	actorRole, err := uc.chatRepo.GetMemberRole(chatID, actorID)
	if err != nil {
		return err
	}

	targetRole, err := uc.chatRepo.GetMemberRole(chatID, memberID)
	if err != nil {
		return err
	}
	// Проверяем права на удаление:

	// 1. Создатель может удалить любого
	if chat.CreatedBy == actorID {
		// Получаем информацию о пользователях для уведомления
		removedUser, err := uc.userRepo.GetByID(memberID)
		if err != nil {
			return err
		}
		actorUser, err := uc.userRepo.GetByID(actorID)
		if err != nil {
			return err
		}
		// Создаем системное сообщение в базе данных и отправляем уведомление ДО удаления пользователя
		systemMessageText := fmt.Sprintf("%s был(а) удален(а) из группы создателем %s", removedUser.Username, actorUser.Username)

		// Создаем системное сообщение в базе данных
		err = uc.createSystemMessage(chatID, systemMessageText)
		if err != nil {
			// Логируем ошибку, но не останавливаем процесс
		}

		if uc.notificationSender != nil {
			notification := &entities.Notification{
				Type:    "user_removed",
				ChatID:  chatID,
				Message: systemMessageText,
				Data: map[string]interface{}{
					"removed_user_id":  memberID,
					"removed_username": removedUser.Username,
					"actor_id":         actorID,
					"actor_username":   actorUser.Username,
					"chat_name":        chat.Name,
				},
			}
			uc.notificationSender.SendNotificationToChat(chatID, notification)
		}

		return uc.chatRepo.RemoveMember(chatID, memberID)
	}
	// 2. Администратор может удалить только обычных пользователей
	if actorRole == "admin" && targetRole == "member" {
		// Получаем информацию о пользователях для уведомления
		removedUser, err := uc.userRepo.GetByID(memberID)
		if err != nil {
			return err
		}
		actorUser, err := uc.userRepo.GetByID(actorID)
		if err != nil {
			return err
		}

		// Создаем системное сообщение в базе данных и отправляем уведомление ДО удаления пользователя
		systemMessageText := fmt.Sprintf("%s был(а) удален(а) из группы администратором %s", removedUser.Username, actorUser.Username)

		// Создаем системное сообщение в базе данных
		err = uc.createSystemMessage(chatID, systemMessageText)
		if err != nil {
			// Логируем ошибку, но не останавливаем процесс
		}

		// Отправляем уведомление ДО удаления пользователя
		if uc.notificationSender != nil {
			notification := &entities.Notification{
				Type:    "user_removed",
				ChatID:  chatID,
				Message: systemMessageText,
				Data: map[string]interface{}{
					"removed_user_id":  memberID,
					"removed_username": removedUser.Username,
					"actor_id":         actorID,
					"actor_username":   actorUser.Username,
					"chat_name":        chat.Name,
				},
			}
			uc.notificationSender.SendNotificationToChat(chatID, notification)
		}

		return uc.chatRepo.RemoveMember(chatID, memberID)
	}

	// 3. Обычный участник не может удалять других пользователей
	if actorRole == "member" {
		return errors.New("regular members cannot remove users")
	}

	// 4. Администратор не может удалить создателя или другого администратора
	return errors.New("you don't have permission to remove this user")
}

// GetChatMembers возвращает список участников чата с их ролями
func (uc *ChatUseCase) GetChatMembers(chatID, userID uint) ([]*entities.User, error) {
	// Специальная обработка для системных вызовов
	if userID != 0 {
		// Проверяем, что пользователь является участником чата
		isMember, err := uc.chatRepo.IsMember(chatID, userID)
		if err != nil {
			return nil, err
		}

		if !isMember {
			return nil, errors.New("user is not a member of this chat")
		}
	}

	// Получаем базовую информацию о чате
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, err
	}

	// Получаем участников чата с информацией о ролях
	members, err := uc.chatRepo.GetMembersWithRoles(chatID)
	if err != nil {
		return nil, err
	}

	// Добавляем информацию о роли Creator для создателя чата
	for i := range members {
		if members[i].ID == chat.CreatedBy {
			members[i].Role = "creator"
		}
	}

	return members, nil
}

// SetAdmin назначает пользователя администратором чата
func (uc *ChatUseCase) SetAdmin(chatID, requesterID, targetUserID uint) error {
	// Проверяем, что инициатор запроса является создателем чата
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %v", err)
	}

	if chat.CreatedBy != requesterID {
		return errors.New("only chat creator can assign admin rights")
	}

	// Проверяем, что целевой пользователь является участником чата
	isMember, err := uc.chatRepo.IsMember(chatID, targetUserID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of this chat")
	}

	// Получаем текущую роль пользователя
	currentRole, err := uc.chatRepo.GetMemberRole(chatID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %v", err)
	}

	// Если пользователь уже администратор, ничего не делаем
	if currentRole == "admin" {
		return nil
	}

	// Обновляем роль пользователя
	return uc.chatRepo.UpdateMemberRole(chatID, targetUserID, "admin")
}

// RemoveAdmin снимает права администратора с пользователя
func (uc *ChatUseCase) RemoveAdmin(chatID, requesterID, targetUserID uint) error {
	// Проверяем, что инициатор запроса является создателем чата
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %v", err)
	}

	if chat.CreatedBy != requesterID {
		return errors.New("only chat creator can remove admin rights")
	}

	// Проверяем, что целевой пользователь является участником чата
	isMember, err := uc.chatRepo.IsMember(chatID, targetUserID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of this chat")
	}

	// Получаем текущую роль пользователя
	currentRole, err := uc.chatRepo.GetMemberRole(chatID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %v", err)
	}

	// Если пользователь не администратор, ничего не делаем
	if currentRole != "admin" {
		return nil
	}

	// Обновляем роль пользователя на обычного участника
	return uc.chatRepo.UpdateMemberRole(chatID, targetUserID, "member")
}

// LeaveChat позволяет пользователю покинуть групповой чат
func (uc *ChatUseCase) LeaveChat(chatID, userID uint) error {
	// Проверяем, что пользователь является участником чата
	isMember, err := uc.chatRepo.IsMember(chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("you are not a member of this chat")
	}

	// Получаем информацию о чате
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	// Проверяем, что это групповой чат
	if !chat.IsGroup {
		return errors.New("you can only leave group chats")
	}

	// Создатель не может покинуть чат, он может только удалить его
	if chat.CreatedBy == userID {
		return errors.New("chat creator cannot leave the chat, please delete it instead")
	}
	// Получаем информацию о пользователе для уведомления
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	// Отправляем уведомление в чат о том, что пользователь покинул группу
	// ВАЖНО: отправляем ДО удаления пользователя из чата, чтобы все участники получили уведомление
	systemMessageText := fmt.Sprintf("%s покинул(а) группу", user.Username)

	// Создаем системное сообщение в базе данных
	err = uc.createSystemMessage(chatID, systemMessageText)
	if err != nil {
		// Логируем ошибку, но не останавливаем процесс
		// Пользователь все равно должен покинуть группу
	}

	if uc.notificationSender != nil {
		notification := &entities.Notification{
			Type:    "user_left",
			ChatID:  chatID,
			Message: systemMessageText,
			Data: map[string]interface{}{
				"user_id":   userID,
				"username":  user.Username,
				"chat_name": chat.Name,
			},
		}
		uc.notificationSender.SendNotificationToChat(chatID, notification)
	}

	// Удаляем пользователя из чата
	err = uc.chatRepo.RemoveMember(chatID, userID)
	if err != nil {
		return err
	}
	return nil
}

// DeletePrivateChat скрывает приватный чат для пользователя
func (uc *ChatUseCase) DeletePrivateChat(chatID, userID uint) error {
	// Проверяем, что пользователь является участником чата
	isMember, err := uc.chatRepo.IsMember(chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("you are not a member of this chat")
	}

	// Получаем информацию о чате
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	// Проверяем, что это приватный чат
	if chat.IsGroup {
		return errors.New("you can only delete private chats, use leave for group chats")
	}

	// Удаляем пользователя из чата (скрываем чат)
	return uc.chatRepo.RemoveMember(chatID, userID)
}

// DeleteGroupChat полностью удаляет групповой чат (только для создателя)
func (uc *ChatUseCase) DeleteGroupChat(chatID, userID uint) error {
	// Получаем информацию о чате
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	// Проверяем, что это групповой чат
	if !chat.IsGroup {
		return errors.New("you can only delete group chats with this method")
	}

	// Проверяем, что пользователь является создателем чата
	if chat.CreatedBy != userID {
		return errors.New("only chat creator can delete the group chat")
	} // Получаем информацию о создателе для уведомления
	creator, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Создаем системное сообщение в базе данных и отправляем уведомление ДО удаления чата
	systemMessageText := fmt.Sprintf("Группа \"%s\" была удалена создателем %s", chat.Name, creator.Username)

	// Создаем системное сообщение в базе данных
	err = uc.createSystemMessage(chatID, systemMessageText)
	if err != nil {
		// Логируем ошибку, но не останавливаем процесс
	}

	// Отправляем уведомление участникам группы о том, что группа была удалена
	// ВАЖНО: отправляем ДО удаления чата, чтобы все участники получили уведомление
	if uc.notificationSender != nil {
		notification := &entities.Notification{
			Type:    "group_deleted",
			ChatID:  chatID,
			Message: systemMessageText,
			Data: map[string]interface{}{
				"creator_id":   userID,
				"creator_name": creator.Username,
				"chat_name":    chat.Name,
			},
		}
		uc.notificationSender.SendNotificationToChat(chatID, notification)
	}

	// Удаляем чат полностью (это также удалит все связанные записи благодаря ON DELETE CASCADE)
	return uc.chatRepo.Delete(chatID)
}

// createSystemMessage создает системное сообщение в базе данных
func (uc *ChatUseCase) createSystemMessage(chatID uint, content string) error {
	systemMessage := &entities.Message{
		ChatID:      chatID,
		SenderID:    0, // 0 означает системное сообщение
		Content:     content,
		MessageType: "system",
	}

	return uc.messageRepo.Create(systemMessage)
}
