import { ECDHService } from '@/shared/lib/crypto/ecdh';

const API_BASE_URL = '/api/v1';

export interface LoginRequest {
  username: string;
  password: string;
  ecdhPublicKey: string;
  ecdsaPublicKey: string;
  rsaPublicKey: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  ecdsaPublicKey: string;
  rsaPublicKey: string;
}

export interface AuthResponse {
  message?: string;
  data?: {
    token: string;
    expires_at: string;
    user: {
      id: number;
      username: string;
      email: string;
      ecdsa_public_key: string;
      rsa_public_key: string;
      is_online: boolean;
      created_at: string;
    };
  };
  // Direct fields for backward compatibility
  token?: string;
  user?: {
    id: string;
    username: string;
    publicKey: string;
  };
}

class ChatAPI {
  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = localStorage.getItem('token');
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
      },
      ...options,
    };

    const response = await fetch(`${API_BASE_URL}${endpoint}`, config);
    
    if (!response.ok) {
      let errorMessage = 'Request failed';
      
      try {
        // Клонируем ответ, чтобы можно было прочитать тело дважды
        const responseClone = response.clone();
        const errorData = await responseClone.json();
        errorMessage = errorData.error || errorData.message || errorMessage;
      } catch {
        // Если не удалось распарсить JSON, используем текст ответа
        try {
          const responseText = await response.text();
          errorMessage = responseText || errorMessage;
        } catch {
          // Оставляем дефолтное сообщение
        }
      }
      
      throw new Error(errorMessage);
    }

    return response.json();
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.request<any>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    
    // Backend returns data in format: {message: "...", data: {...}}
    // We need to return the data field
    return response.data || response;
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.request<any>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    
    // Backend returns data in format: {message: "...", data: {...}}
    // We need to return the data field
    return response.data || response;
  }

  async getChats(): Promise<any[]> {
    const response = await this.request<any>('/chats');
    
    // Backend returns data in format: {data: [...]}
    // We need to return the data field
    return response.data || response;
  }

  async getMessages(chatId: string): Promise<any[]> {
    const response = await this.request<any>(`/chats/${chatId}/messages`);
    
    // Backend returns data in format: {data: [...]}
    // We need to return the data field
    return response.data || response;
  }

  async sendMessage(chatId: string, content: string, messageType: string = 'text'): Promise<any> {
    return this.request<any>(`/chats/${chatId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content, messageType }),
    });
  }

  async createChat(name: string, isGroup: boolean, participants: number[]): Promise<any> {
    return this.request<any>('/chats', {
      method: 'POST',
      body: JSON.stringify({ 
        name, 
        is_group: isGroup, 
        member_ids: participants 
      }),
    });
  }

  async createOrGetPrivateChat(userId: number, username: string): Promise<any> {
    return this.request<any>('/chats/private', {
      method: 'POST',
      body: JSON.stringify({ 
        user_id: userId,
        username: username
      }),
    });
  }

  async addMember(chatId: string, userId: number): Promise<any> {
    return this.request<any>(`/chats/${chatId}/members`, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId }),
    });
  }

  async removeMember(chatId: string, userId: number): Promise<any> {
    return this.request<any>(`/chats/${chatId}/members/${userId}`, {
      method: 'DELETE',
    });
  }

  async getChatMembers(chatId: string): Promise<any[]> {
    const response = await this.request<any>(`/chats/${chatId}/members`);
    
    // Backend returns data in format: {data: [...]}
    // We need to return the data field
    return response.data || response;
  }
  
  async setAdmin(chatId: string, userId: number): Promise<any> {
    return this.request<any>(`/chats/${chatId}/members/${userId}/admin`, {
      method: 'PUT',
      body: JSON.stringify({ role: 'admin' }),
    });
  }

  async removeAdmin(chatId: string, userId: number): Promise<any> {
    return this.request<any>(`/chats/${chatId}/members/${userId}/admin`, {
      method: 'DELETE',
    });
  }

  async logout(): Promise<any> {
    return this.request<any>('/auth/logout', {
      method: 'POST',
    });
  }

  async getProfile(): Promise<any> {
    return this.request<any>('/auth/profile');
  }

  async changePassword(oldPassword: string, newPassword: string): Promise<any> {
    return this.request<any>('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({
        oldPassword,
        newPassword
      })
    });
  }

  // Выход из группового чата (пользователь покидает группу)
  async leaveChat(chatId: string): Promise<any> {
    return this.request<any>(`/chats/${chatId}/leave`, {
      method: 'POST',
    });
  }

  // Удаление приватного чата (скрывает чат для пользователя)
  async deleteChat(chatId: string): Promise<any> {
    return this.request<any>(`/chats/${chatId}`, {
      method: 'DELETE',
    });
  }

  // Полное удаление группового чата (только для создателя)
  async deleteGroupChat(chatId: string): Promise<any> {
    return this.request<any>(`/chats/${chatId}/delete`, {
      method: 'DELETE',
    });
  }
}

export const chatAPI = new ChatAPI();
