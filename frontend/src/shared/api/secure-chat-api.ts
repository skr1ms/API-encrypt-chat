import { SecureApiClient } from './secure-client';

export interface Chat {
  id: number;
  name: string;
  is_group: boolean;
  created_by: number;
  creator: User;
  created_at: string;
  updated_at: string;
  members: User[];
  messages?: Message[];
}

export interface User {
  id: number;
  username: string;
  email: string;
  ecdsa_public_key: string;
  rsa_public_key: string;
  is_online: boolean;
  role?: string;
  last_seen?: string;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: number;
  chat_id: number;
  sender_id: number;
  sender: User;
  content: string;
  message_type: string;
  timestamp?: number;
  nonce: string;
  iv: string;
  hmac: string;
  ecdsa_signature: string;
  rsa_signature: string;
  is_edited: boolean;
  edited_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateChatRequest {
  name: string;
  is_group?: boolean;
  member_ids?: number[];
}

export interface SendMessageRequest {
  content: string;
  message_type?: string;
}

export interface PrivateChatRequest {
  user_id: number;
}

export class SecureChatAPI {
  constructor(private client: SecureApiClient) {}

  // === Chat Management ===

  async createChat(data: CreateChatRequest): Promise<Chat> {
    return this.client.authenticatedPost<Chat>('/chats', data);
  }

  async createOrGetPrivateChat(data: PrivateChatRequest): Promise<Chat> {
    return this.client.authenticatedPost<Chat>('/chats/private', data);
  }

  async getUserChats(): Promise<Chat[]> {
    return this.client.authenticatedGet<Chat[]>('/chats');
  }

  async getChatMessages(chatId: number, limit = 50, offset = 0): Promise<Message[]> {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });
    
    return this.client.authenticatedGet<Message[]>(`/chats/${chatId}/messages?${params}`);
  }

  async sendMessage(chatId: number, data: SendMessageRequest): Promise<Message> {
    return this.client.authenticatedPost<Message>(`/chats/${chatId}/messages`, data);
  }

  async getChatMembers(chatId: number): Promise<User[]> {
    return this.client.authenticatedGet<User[]>(`/chats/${chatId}/members`);
  }

  async addMember(chatId: number, userId: number): Promise<{ message: string }> {
    return this.client.authenticatedPost<{ message: string }>(`/chats/${chatId}/members`, {
      user_id: userId,
    });
  }

  async removeMember(chatId: number, userId: number): Promise<{ message: string }> {
    return this.client.authenticatedDelete<{ message: string }>(`/chats/${chatId}/members/${userId}`);
  }

  async setAdmin(chatId: number, userId: number): Promise<{ message: string }> {
    return this.client.authenticatedPut<{ message: string }>(`/chats/${chatId}/members/${userId}/admin`);
  }

  async removeAdmin(chatId: number, userId: number): Promise<{ message: string }> {
    return this.client.authenticatedDelete<{ message: string }>(`/chats/${chatId}/members/${userId}/admin`);
  }

  async leaveChat(chatId: number): Promise<{ message: string }> {
    return this.client.authenticatedPost<{ message: string }>(`/chats/${chatId}/leave`);
  }

  async deleteChat(chatId: number): Promise<{ message: string }> {
    return this.client.authenticatedDelete<{ message: string }>(`/chats/${chatId}`);
  }

  async deleteGroupChat(chatId: number): Promise<{ message: string }> {
    return this.client.authenticatedDelete<{ message: string }>(`/chats/${chatId}/delete`);
  }

  // === User Management ===

  async searchUsers(query: string, limit = 20): Promise<User[]> {
    const params = new URLSearchParams({
      q: query,
      limit: limit.toString(),
    });
    
    return this.client.authenticatedGet<User[]>(`/users/search?${params}`);
  }

  async getOnlineUsers(): Promise<User[]> {
    return this.client.authenticatedGet<User[]>('/users/online');
  }

  async getUser(userId: number): Promise<User> {
    return this.client.authenticatedGet<User>(`/users/${userId}`);
  }

  // === Authentication (использует обычные запросы, не зашифрованные) ===

  async login(credentials: {
    username: string;
    password: string;
    ecdhPublicKey: string;
    ecdsaPublicKey: string;
    rsaPublicKey: string;
  }): Promise<{
    token: string;
    expires_at: string;
    user: User;
  }> {
    const response = await this.client.post<{
      data: {
        token: string;
        expires_at: string;
        user: User;
      };
    }>('/auth/login', credentials);
    
    return response.data;
  }

  async register(userData: {
    username: string;
    email: string;
    password: string;
    ecdsaPublicKey: string;
    rsaPublicKey: string;
  }): Promise<{
    token: string;
    expires_at: string;
    user: User;
  }> {
    const response = await this.client.post<{
      data: {
        token: string;
        expires_at: string;
        user: User;
      };
    }>('/auth/register', userData);
    
    return response.data;
  }

  async logout(): Promise<{ message: string }> {
    return this.client.authenticatedPost<{ message: string }>('/auth/logout');
  }

  async getProfile(): Promise<User> {
    const response = await this.client.authenticatedGet<{
      data: { user: User };
    }>('/auth/profile');
    
    return response.data.user;
  }

  async changePassword(data: {
    current_password: string;
    new_password: string;
  }): Promise<{ message: string }> {
    return this.client.authenticatedPost<{ message: string }>('/auth/change-password', data);
  }
}

// Фабрика для создания экземпляра API
export const createSecureChatAPI = (client: SecureApiClient): SecureChatAPI => {
  return new SecureChatAPI(client);
};

// Хук для использования в React компонентах
import { useMemo } from 'react';
import { useSecureApiClient } from '../providers/encryption-provider';

export const useSecureChatAPI = (): SecureChatAPI => {
  const client = useSecureApiClient();
  
  return useMemo(() => createSecureChatAPI(client), [client]);
};
