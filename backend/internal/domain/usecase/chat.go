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

// NewChatUseCase - создает новый экземпляр сервиса для работы с чатами
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

// CreateChat - создает новый чат (групповой или приватный)
func (uc *ChatUseCase) CreateChat(creatorID uint, req *CreateChatRequest) (*entities.Chat, error) {
	creator, err := uc.userRepo.GetByID(creatorID)
	if err != nil {
		return nil, errors.New("creator not found")
	}

	chat := &entities.Chat{
		Name:      req.Name,
		IsGroup:   req.IsGroup,
		CreatedBy: creatorID,
		Creator:   *creator,
	}

	if err := uc.chatRepo.Create(chat); err != nil {
		return nil, fmt.Errorf("failed to create chat: %v", err)
	}

	if err := uc.chatRepo.AddMember(chat.ID, creatorID, "admin"); err != nil {
		return nil, fmt.Errorf("failed to add creator to chat: %v", err)
	}

	for _, memberID := range req.MemberIDs {
		if memberID != creatorID {
			if err := uc.chatRepo.AddMember(chat.ID, memberID, "member"); err != nil {
				return nil, fmt.Errorf("failed to add member %d to chat: %v", memberID, err)
			}
		}
	}

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

// GetUserChats - получает список всех чатов пользователя
func (uc *ChatUseCase) GetUserChats(userID uint) ([]entities.Chat, error) {
	chats, err := uc.chatRepo.GetUserChats(userID)
	if err != nil {
		return nil, err
	}

	for i := range chats {
		if !chats[i].IsGroup {
			members, err := uc.chatRepo.GetMembers(chats[i].ID)
			if err != nil {
				continue
			}

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

// CreateOrGetPrivateChat - создает новый приватный чат или возвращает существующий
func (uc *ChatUseCase) CreateOrGetPrivateChat(userID1, userID2 uint, otherUserName string) (*PrivateChatResponse, error) {
	existingChat, err := uc.chatRepo.FindPrivateChat(userID1, userID2)
	if err == nil {
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

	chatName := "Private Chat"
	req := &CreateChatRequest{
		Name:      chatName,
		IsGroup:   false,
		MemberIDs: []uint{userID2},
	}

	newChat, err := uc.CreateChat(userID1, req)
	if err != nil {
		return nil, err
	}

	newChat.Name = fmt.Sprintf("Chat with %s", otherUserName)

	return &PrivateChatResponse{
		Chat:    newChat,
		Created: true,
	}, nil
}

// SendMessage - отправляет зашифрованное сообщение в чат
func (uc *ChatUseCase) SendMessage(chatID, senderID uint, req *SendMessageRequest, senderECDSAPrivateKey *ecdsa.PrivateKey, senderRSAPrivateKey *rsa.PrivateKey) (*entities.Message, error) {
	isMember, err := uc.chatRepo.IsMember(chatID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("sender is not a member of the chat")
	}

	members, err := uc.chatRepo.GetMembers(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat members: %v", err)
	}

	sender, err := uc.userRepo.GetByID(senderID)
	if err != nil {
		return nil, errors.New("sender not found")
	}

	var sharedSecret []byte
	var recipientID uint = senderID

	if senderECDSAPrivateKey != nil && len(members) > 1 {
		var recipientPublicKey []byte
		for _, member := range members {
			if member.ID != senderID {
				recipientPublicKey, err = hex.DecodeString(member.ECDSAPublicKey)
				if err != nil {
					return nil, fmt.Errorf("failed to decode recipient public key: %v", err)
				}
				recipientID = member.ID
				break
			}
		}

		if len(recipientPublicKey) > 0 {
			sharedSecret, err = crypto.ComputeECDHSharedSecret(senderECDSAPrivateKey, recipientPublicKey)
		}
	}

	if len(sharedSecret) == 0 {
		sharedSecret = make([]byte, 64)
		copy(sharedSecret, "default-shared-secret-for-single-user-or-error")
	}

	secureMsg, err := crypto.CreateSecureMessage(
		fmt.Sprintf("%d", senderID),
		fmt.Sprintf("%d", recipientID),
		[]byte(req.Content),
		sharedSecret,
		senderECDSAPrivateKey,
		senderRSAPrivateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure message: %v", err)
	}

	message := &entities.Message{
		ChatID:         chatID,
		SenderID:       senderID,
		Content:        secureMsg.Ciphertext,
		MessageType:    req.MessageType,
		Timestamp:      &secureMsg.Timestamp,
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

	message.Sender = *sender
	chat, _ := uc.chatRepo.GetByID(chatID)
	if chat != nil {
		message.Chat = *chat
	}

	return message, nil
}

// GetChatMessages - получает список сообщений чата с расшифровкой для пользователя
func (uc *ChatUseCase) GetChatMessages(chatID, userID uint, limit, offset int) ([]MessageResponse, error) {
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

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	var responses []MessageResponse
	for _, msg := range messages {
		response := MessageResponse{
			Message: &msg,
		}

		decryptedContent, err := uc.decryptMessage(&msg, user)
		if err != nil {
			response.DecryptedContent = msg.Content
		} else {
			response.DecryptedContent = decryptedContent
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// decryptMessage - расшифровывает зашифрованное сообщение для конкретного пользователя
func (uc *ChatUseCase) decryptMessage(msg *entities.Message, user *entities.User) (string, error) {
	if msg.Content == "" || msg.IV == "" || msg.HMAC == "" {
		return msg.Content, nil
	}

	sender, err := uc.userRepo.GetByID(msg.SenderID)
	if err != nil {
		return "", fmt.Errorf("sender not found: %v", err)
	}

	userECDSAPrivateKey, err := crypto.DeserializeECDSAPrivateKey([]byte(user.ECDSAPrivateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse user ECDSA private key: %v", err)
	}

	senderECDSAPublicKeyBytes, err := hex.DecodeString(sender.ECDSAPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode sender ECDSA public key: %v", err)
	}

	senderRSAPublicKeyBytes, err := hex.DecodeString(sender.RSAPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode sender RSA public key: %v", err)
	}

	members, err := uc.chatRepo.GetMembers(msg.ChatID)
	if err != nil {
		return "", fmt.Errorf("failed to get chat members: %v", err)
	}

	var sharedSecret []byte
	if user.ID == msg.SenderID {
		for _, member := range members {
			if member.ID != msg.SenderID {
				recipientPublicKeyBytes, err := hex.DecodeString(member.ECDSAPublicKey)
				if err != nil {
					return "", fmt.Errorf("failed to decode recipient public key: %v", err)
				}
				sharedSecret, err = crypto.ComputeECDHSharedSecret(userECDSAPrivateKey, recipientPublicKeyBytes)
				break
			}
		}
	} else {
		sharedSecret, err = crypto.ComputeECDHSharedSecret(userECDSAPrivateKey, senderECDSAPublicKeyBytes)
	}

	if len(sharedSecret) == 0 {
		sharedSecret = make([]byte, 64)
		copy(sharedSecret, "default-shared-secret-for-single-user-or-error")
	}

	var recipientID uint = msg.SenderID

	if len(members) > 1 {
		for _, member := range members {
			if member.ID != msg.SenderID {
				recipientID = member.ID
				break
			}
		}
	}

	timestamp := msg.CreatedAt.Unix()
	if msg.Timestamp != nil {
		timestamp = *msg.Timestamp
	}

	secureMsg := &crypto.SecureMessage{
		Ciphertext:     msg.Content,
		IV:             msg.IV,
		HMAC:           msg.HMAC,
		ECDSASignature: msg.ECDSASignature,
		RSASignature:   msg.RSASignature,
		Nonce:          msg.Nonce,
		Timestamp:      timestamp,
		SenderID:       fmt.Sprintf("%d", msg.SenderID),
		RecipientID:    fmt.Sprintf("%d", recipientID),
	}

	plaintext, err := crypto.VerifyAndDecryptMessage(secureMsg, sharedSecret, senderECDSAPublicKeyBytes, senderRSAPublicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt message: %v", err)
	}

	return string(plaintext), nil
}

// AddMember - добавляет нового участника в чат
func (uc *ChatUseCase) AddMember(chatID, requesterID, newMemberID uint) error {
	isMember, err := uc.chatRepo.IsMember(chatID, requesterID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("you are not a member of this chat")
	}

	isAlreadyMember, err := uc.chatRepo.IsMember(chatID, newMemberID)
	if err != nil {
		return err
	}
	if isAlreadyMember {
		return errors.New("user is already a member of this chat")
	}

	err = uc.chatRepo.AddMember(chatID, newMemberID, "member")
	if err != nil {
		return err
	}

	newUser, err := uc.userRepo.GetByID(newMemberID)
	if err != nil {
		return nil
	}

	systemMessageText := fmt.Sprintf("%s присоединился к группе", newUser.Username)

	err = uc.createSystemMessage(chatID, systemMessageText)
	if err != nil {
	}

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

// AddMemberWithUserData - добавляет нового участника в чат и возвращает данные пользователя
func (uc *ChatUseCase) AddMemberWithUserData(chatID, requesterID, newMemberID uint) (*entities.User, error) {
	isMember, err := uc.chatRepo.IsMember(chatID, requesterID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("you are not a member of this chat")
	}

	isAlreadyMember, err := uc.chatRepo.IsMember(chatID, newMemberID)
	if err != nil {
		return nil, err
	}
	if isAlreadyMember {
		return nil, errors.New("user is already a member of this chat")
	}

	err = uc.chatRepo.AddMember(chatID, newMemberID, "member")
	if err != nil {
		return nil, err
	}

	newUser, err := uc.userRepo.GetByID(newMemberID)
	if err != nil {
		return nil, err
	}

	systemMessageText := fmt.Sprintf("%s присоединился к группе", newUser.Username)

	err = uc.createSystemMessage(chatID, systemMessageText)
	if err != nil {
	}

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

// RemoveMember - удаляет участника из чата (только админы и создатель)
func (uc *ChatUseCase) RemoveMember(chatID, actorID, memberID uint) error {
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

	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	actorRole, err := uc.chatRepo.GetMemberRole(chatID, actorID)
	if err != nil {
		return err
	}

	targetRole, err := uc.chatRepo.GetMemberRole(chatID, memberID)
	if err != nil {
		return err
	}

	if chat.CreatedBy == actorID {
		removedUser, err := uc.userRepo.GetByID(memberID)
		if err != nil {
			return err
		}
		actorUser, err := uc.userRepo.GetByID(actorID)
		if err != nil {
			return err
		}

		systemMessageText := fmt.Sprintf("%s был(а) удален(а) из группы создателем %s", removedUser.Username, actorUser.Username)

		err = uc.createSystemMessage(chatID, systemMessageText)
		if err != nil {
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

	if actorRole == "admin" && targetRole == "member" {
		removedUser, err := uc.userRepo.GetByID(memberID)
		if err != nil {
			return err
		}
		actorUser, err := uc.userRepo.GetByID(actorID)
		if err != nil {
			return err
		}

		systemMessageText := fmt.Sprintf("%s был(а) удален(а) из группы администратором %s", removedUser.Username, actorUser.Username)

		err = uc.createSystemMessage(chatID, systemMessageText)
		if err != nil {
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

	if actorRole == "member" {
		return errors.New("regular members cannot remove users")
	}

	return errors.New("you don't have permission to remove this user")
}

// GetChatMembers - получает список всех участников чата с их ролями
func (uc *ChatUseCase) GetChatMembers(chatID, userID uint) ([]*entities.User, error) {
	if userID != 0 {
		isMember, err := uc.chatRepo.IsMember(chatID, userID)
		if err != nil {
			return nil, err
		}

		if !isMember {
			return nil, errors.New("user is not a member of this chat")
		}
	}

	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, err
	}

	members, err := uc.chatRepo.GetMembersWithRoles(chatID)
	if err != nil {
		return nil, err
	}

	for i := range members {
		if members[i].ID == chat.CreatedBy {
			members[i].Role = "creator"
		}
	}

	return members, nil
}

// SetAdmin - назначает пользователя администратором чата (только создатель)
func (uc *ChatUseCase) SetAdmin(chatID, requesterID, targetUserID uint) error {
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %v", err)
	}

	if chat.CreatedBy != requesterID {
		return errors.New("only chat creator can assign admin rights")
	}

	isMember, err := uc.chatRepo.IsMember(chatID, targetUserID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of this chat")
	}

	currentRole, err := uc.chatRepo.GetMemberRole(chatID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %v", err)
	}

	if currentRole == "admin" {
		return nil
	}

	return uc.chatRepo.UpdateMemberRole(chatID, targetUserID, "admin")
}

// RemoveAdmin - снимает права администратора с пользователя (только создатель)
func (uc *ChatUseCase) RemoveAdmin(chatID, requesterID, targetUserID uint) error {
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %v", err)
	}

	if chat.CreatedBy != requesterID {
		return errors.New("only chat creator can remove admin rights")
	}

	isMember, err := uc.chatRepo.IsMember(chatID, targetUserID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of this chat")
	}

	currentRole, err := uc.chatRepo.GetMemberRole(chatID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %v", err)
	}

	if currentRole != "admin" {
		return nil
	}

	return uc.chatRepo.UpdateMemberRole(chatID, targetUserID, "member")
}

// LeaveChat - позволяет пользователю покинуть групповой чат
func (uc *ChatUseCase) LeaveChat(chatID, userID uint) error {
	isMember, err := uc.chatRepo.IsMember(chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("you are not a member of this chat")
	}

	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	if !chat.IsGroup {
		return errors.New("you can only leave group chats")
	}

	if chat.CreatedBy == userID {
		return errors.New("chat creator cannot leave the chat, please delete it instead")
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	systemMessageText := fmt.Sprintf("%s покинул(а) группу", user.Username)

	err = uc.createSystemMessage(chatID, systemMessageText)
	if err != nil {
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

	err = uc.chatRepo.RemoveMember(chatID, userID)
	if err != nil {
		return err
	}
	return nil
}

// DeletePrivateChat - удаляет приватный чат для пользователя
func (uc *ChatUseCase) DeletePrivateChat(chatID, userID uint) error {
	isMember, err := uc.chatRepo.IsMember(chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("you are not a member of this chat")
	}

	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	if chat.IsGroup {
		return errors.New("you can only delete private chats, use leave for group chats")
	}

	return uc.chatRepo.RemoveMember(chatID, userID)
}

// DeleteGroupChat - полностью удаляет групповой чат (только создатель)
func (uc *ChatUseCase) DeleteGroupChat(chatID, userID uint) error {
	chat, err := uc.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}

	if !chat.IsGroup {
		return errors.New("you can only delete group chats with this method")
	}

	if chat.CreatedBy != userID {
		return errors.New("only chat creator can delete the group chat")
	}

	creator, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	systemMessageText := fmt.Sprintf("Группа \"%s\" была удалена создателем %s", chat.Name, creator.Username)

	err = uc.createSystemMessage(chatID, systemMessageText)
	if err != nil {
	}

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

	return uc.chatRepo.Delete(chatID)
}

// createSystemMessage - создает системное сообщение в чате
func (uc *ChatUseCase) createSystemMessage(chatID uint, content string) error {
	systemMessage := &entities.Message{
		ChatID:      chatID,
		SenderID:    0,
		Content:     content,
		MessageType: "system",
	}

	return uc.messageRepo.Create(systemMessage)
}
