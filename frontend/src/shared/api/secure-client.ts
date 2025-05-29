import { EncryptionService } from '../lib/crypto/encryption';

export interface ApiConfig {
  baseURL: string;
  userId?: number;
  autoInitEncryption?: boolean;
}

export class SecureApiClient {
  private encryption: EncryptionService;
  private baseURL: string;
  private userId: number | null;

  constructor(config: ApiConfig) {
    this.encryption = new EncryptionService();
    this.baseURL = config.baseURL;
    this.userId = config.userId || null;

    // Автоматически инициализируем шифрование если указан userId
    if (config.autoInitEncryption && config.userId) {
      this.initializeEncryption(config.userId);
    }
  }

  /**
   * Инициализирует шифрование для пользователя
   */
  async initializeEncryption(userId: number): Promise<boolean> {
    this.userId = userId;
    
    try {
      const success = await this.encryption.initiateKeyExchange(userId, this.baseURL);
      if (success) {
        console.log('Secure API client initialized successfully');
      }
      return success;
    } catch (error) {
      console.error('Failed to initialize encryption:', error);
      return false;
    }
  }

  /**
   * Обновляет ключи шифрования
   */
  async refreshEncryption(): Promise<boolean> {
    if (!this.userId) {
      throw new Error('User ID not set');
    }

    return await this.encryption.refreshSession(this.userId, this.baseURL);
  }

  /**
   * Проверяет состояние сессии
   */
  async validateSession(): Promise<boolean> {
    return await this.encryption.validateSession(this.baseURL);
  }

  /**
   * Завершает сессию
   */
  async logout(): Promise<void> {
    await this.encryption.revokeSession(this.baseURL);
    this.encryption.clearSession();
    this.userId = null;
  }

  /**
   * Выполняет GET запрос
   */
  async get<T>(endpoint: string, headers?: Record<string, string>): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    
    // Для GET запросов не используем шифрование тела, но добавляем sessionId в заголовки
    const requestHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      ...headers,
    };

    if (this.encryption.hasActiveSession()) {
      requestHeaders['X-Session-ID'] = this.encryption.getSessionId()!;
    }

    const response = await fetch(url, {
      method: 'GET',
      headers: requestHeaders,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Выполняет POST запрос с шифрованием
   */
  async post<T>(endpoint: string, data?: any, headers?: Record<string, string>): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;

    // Если шифрование доступно и это не эндпоинт обмена ключами, используем зашифрованный запрос
    if (this.encryption.hasActiveSession() && !endpoint.includes('/key-exchange/')) {
      return await this.encryption.encryptedFetch<T>(url, {
        method: 'POST',
        headers: headers,
      }, data);
    }

    // Обычный незашифрованный запрос
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: data ? JSON.stringify(data) : undefined,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Выполняет PUT запрос с шифрованием
   */
  async put<T>(endpoint: string, data?: any, headers?: Record<string, string>): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;

    if (this.encryption.hasActiveSession()) {
      return await this.encryption.encryptedFetch<T>(url, {
        method: 'PUT',
        headers: headers,
      }, data);
    }

    const response = await fetch(url, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: data ? JSON.stringify(data) : undefined,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Выполняет DELETE запрос
   */
  async delete<T>(endpoint: string, headers?: Record<string, string>): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;

    const requestHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      ...headers,
    };

    if (this.encryption.hasActiveSession()) {
      requestHeaders['X-Session-ID'] = this.encryption.getSessionId()!;
    }

    const response = await fetch(url, {
      method: 'DELETE',
      headers: requestHeaders,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Выполняет обычный fetch запрос без шифрования
   */
  async fetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Возвращает информацию о текущей сессии
   */
  getSessionInfo() {
    return {
      hasActiveSession: this.encryption.hasActiveSession(),
      sessionId: this.encryption.getSessionId(),
      userId: this.userId,
    };
  }

  /**
   * Устанавливает токен авторизации для запросов
   */
  setAuthToken(token: string) {
    // Можно добавить логику для сохранения токена и его использования в заголовках
    localStorage.setItem('authToken', token);
  }

  /**
   * Получает токен авторизации
   */
  getAuthToken(): string | null {
    return localStorage.getItem('authToken');
  }

  /**
   * Добавляет заголовок авторизации к запросу если токен доступен
   */
  private addAuthHeader(headers: Record<string, string> = {}): Record<string, string> {
    const token = this.getAuthToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
  }

  /**
   * Выполняет аутентифицированный GET запрос
   */
  async authenticatedGet<T>(endpoint: string, headers?: Record<string, string>): Promise<T> {
    return this.get<T>(endpoint, this.addAuthHeader(headers));
  }

  /**
   * Выполняет аутентифицированный POST запрос
   */
  async authenticatedPost<T>(endpoint: string, data?: any, headers?: Record<string, string>): Promise<T> {
    return this.post<T>(endpoint, data, this.addAuthHeader(headers));
  }

  /**
   * Выполняет аутентифицированный PUT запрос
   */
  async authenticatedPut<T>(endpoint: string, data?: any, headers?: Record<string, string>): Promise<T> {
    return this.put<T>(endpoint, data, this.addAuthHeader(headers));
  }

  /**
   * Выполняет аутентифицированный DELETE запрос
   */
  async authenticatedDelete<T>(endpoint: string, headers?: Record<string, string>): Promise<T> {
    return this.delete<T>(endpoint, this.addAuthHeader(headers));
  }

  /**
   * Выполняет PATCH запрос с шифрованием
   */
  async patch<T>(endpoint: string, data?: any, headers?: Record<string, string>): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;

    if (this.encryption.hasActiveSession()) {
      return await this.encryption.encryptedFetch<T>(url, {
        method: 'PATCH',
        headers: headers,
      }, data);
    }

    const response = await fetch(url, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: data ? JSON.stringify(data) : undefined,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  /**
   * Выполняет аутентифицированный PATCH запрос
   */
  async authenticatedPatch<T>(endpoint: string, data?: any, headers?: Record<string, string>): Promise<T> {
    return this.patch<T>(endpoint, data, this.addAuthHeader(headers));
  }
}

// Экспортируем создание глобального экземпляра API клиента
export const createSecureApiClient = (config: ApiConfig): SecureApiClient => {
  return new SecureApiClient(config);
};
