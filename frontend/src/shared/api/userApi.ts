const API_BASE_URL = '/api/v1';

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
    console.log('UserAPI making request to:', endpoint);
    console.log('UserAPI token exists:', !!token);
    console.log('UserAPI token preview:', token ? `${token.substring(0, 20)}...` : 'no token');
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
      },
      ...options,
    };

    console.log('UserAPI request headers:', config.headers);

    const response = await fetch(`${API_BASE_URL}${endpoint}`, config);
    
    console.log('UserAPI response status:', response.status);
    console.log('UserAPI response ok:', response.ok);
    
    if (!response.ok) {
      let errorMessage = 'Request failed';
      
      try {
        const errorData = await response.json();
        errorMessage = errorData.error || errorMessage;
        console.log('UserAPI error data:', errorData);
      } catch {
        // Ignore JSON parsing errors
      }
      
      // Handle authentication errors specifically
      if (response.status === 401) {
        // Clear invalid token
        localStorage.removeItem('token');
        errorMessage = 'Authentication required. Please login again.';
        console.log('UserAPI: 401 error, token cleared from localStorage');
      }
      
      console.error('UserAPI request failed:', response.status, errorMessage);
      throw new Error(errorMessage);
    }

    const responseData = await response.json();
    console.log('UserAPI response data:', responseData);
    return responseData;
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
