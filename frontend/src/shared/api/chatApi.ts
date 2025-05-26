import { ECDHService } from '@/shared/lib/crypto/ecdh';

const API_BASE_URL = 'http://localhost:8080/api/v1';

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
  token: string;
  user: {
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
        const errorData = await response.json();
        errorMessage = errorData.error || errorMessage;
      } catch {
        // Если не удалось распарсить JSON, используем текст ответа
        errorMessage = await response.text() || errorMessage;
      }
      
      throw new Error(errorMessage);
    }

    return response.json();
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getChats(): Promise<any[]> {
    return this.request<any[]>('/chats');
  }

  async getMessages(chatId: string): Promise<any[]> {
    return this.request<any[]>(`/chats/${chatId}/messages`);
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
      body: JSON.stringify({ name, isGroup, participants }),
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

  async searchUsers(query: string): Promise<any[]> {
    return this.request<any[]>(`/users/search?q=${encodeURIComponent(query)}`);
  }
}

export const chatAPI = new ChatAPI();
