const API_BASE_URL = 'http://localhost:8080/api/v1';

export interface User {
  id: number;
  username: string;
  email: string;
  is_online: boolean;
  last_seen?: string;
  ecdsa_public_key?: string;
  rsa_public_key?: string;
  created_at?: string;
}

export interface SearchUsersResponse {
  users: User[];
  total: number;
}

export interface OnlineUsersResponse {
  users: User[];
  total: number;
}

class UserAPI {  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = localStorage.getItem('token');
    console.log('Making request to:', endpoint, 'with token:', token ? `${token.substring(0, 20)}...` : 'no token');
    
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
        // Ignore JSON parsing errors
      }
      
      // Handle authentication errors specifically
      if (response.status === 401) {
        // Clear invalid token
        localStorage.removeItem('token');
        errorMessage = 'Authentication required. Please login again.';
      }
      
      console.error('Request failed:', response.status, errorMessage);
      throw new Error(errorMessage);
    }

    return response.json();
  }

  async searchUsers(query: string, limit: number = 10): Promise<SearchUsersResponse> {
    const params = new URLSearchParams({
      q: query,
      limit: limit.toString()
    });
    
    return this.request<SearchUsersResponse>(`/users/search?${params}`);
  }

  async getUserById(id: number): Promise<User> {
    return this.request<User>(`/users/${id}`);
  }

  async getOnlineUsers(): Promise<OnlineUsersResponse> {
    return this.request<OnlineUsersResponse>('/users/online');
  }
}

export const userAPI = new UserAPI();
