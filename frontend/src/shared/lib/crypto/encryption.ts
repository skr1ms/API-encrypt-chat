import CryptoJS from 'crypto-js';
import { ECDHService } from './ecdh';

export interface EncryptedRequest {
  data: string;      // Зашифрованные данные в base64
  iv: string;        // Вектор инициализации в base64
  hmac: string;      // HMAC для проверки целостности
  sessionId: string; // ID сессии
}

export interface EncryptedResponse {
  data: string; // Зашифрованные данные в base64
  iv: string;   // Вектор инициализации в base64
  hmac: string; // HMAC для проверки целостности
}

export interface KeyExchangeRequest {
  clientPublicKey: string;
  userId: number;
}

export interface KeyExchangeResponse {
  serverPublicKey: string;
  sessionId: string;
  expiresAt: number;
}

export interface SessionKeys {
  aesKey: string;
  hmacKey: string;
}

export class EncryptionService {
  private sessionId: string | null = null;
  private sessionKeys: SessionKeys | null = null;
  private ecdhService: ECDHService;

  constructor() {
    this.ecdhService = new ECDHService();
  }

  /**
   * Инициирует обмен ключами с сервером
   */
  async initiateKeyExchange(userId: number, baseURL: string = ''): Promise<boolean> {
    try {
      // Генерируем пару ключей ECDH
      const keyPair = this.ecdhService.generateKeyPair();
      
      const request: KeyExchangeRequest = {
        clientPublicKey: keyPair.publicKey,
        userId: userId
      };

      // Отправляем запрос на обмен ключами
      const response = await fetch(`${baseURL}/api/key-exchange/initiate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request)
      });

      if (!response.ok) {
        throw new Error(`Key exchange failed: ${response.statusText}`);
      }

      const keyExchangeResponse: KeyExchangeResponse = await response.json();

      // Вычисляем общий секрет
      const sharedSecret = this.ecdhService.computeSharedSecret(keyExchangeResponse.serverPublicKey);

      // Деривируем ключи сессии
      this.sessionKeys = this.deriveSessionKeys(sharedSecret);
      this.sessionId = keyExchangeResponse.sessionId;

      console.log('Key exchange completed successfully', {
        sessionId: this.sessionId,
        expiresAt: new Date(keyExchangeResponse.expiresAt * 1000).toISOString()
      });

      return true;
    } catch (error) {
      console.error('Key exchange failed:', error);
      return false;
    }
  }

  /**
   * Обновляет ключи существующей сессии
   */
  async refreshSession(userId: number, baseURL: string = ''): Promise<boolean> {
    if (!this.sessionId) {
      throw new Error('No active session to refresh');
    }

    try {
      // Генерируем новую пару ключей ECDH
      const keyPair = this.ecdhService.generateKeyPair();
      
      const request: KeyExchangeRequest = {
        clientPublicKey: keyPair.publicKey,
        userId: userId
      };

      const response = await fetch(`${baseURL}/api/key-exchange/refresh/${this.sessionId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request)
      });

      if (!response.ok) {
        throw new Error(`Session refresh failed: ${response.statusText}`);
      }

      const keyExchangeResponse: KeyExchangeResponse = await response.json();

      // Вычисляем новый общий секрет
      const sharedSecret = this.ecdhService.computeSharedSecret(keyExchangeResponse.serverPublicKey);

      // Деривируем новые ключи сессии
      this.sessionKeys = this.deriveSessionKeys(sharedSecret);
      this.sessionId = keyExchangeResponse.sessionId;

      console.log('Session refreshed successfully', {
        sessionId: this.sessionId,
        expiresAt: new Date(keyExchangeResponse.expiresAt * 1000).toISOString()
      });

      return true;
    } catch (error) {
      console.error('Session refresh failed:', error);
      return false;
    }
  }

  /**
   * Шифрует HTTP-запрос
   */
  encryptRequest(data: any): EncryptedRequest {
    if (!this.sessionId || !this.sessionKeys) {
      throw new Error('Session not established. Call initiateKeyExchange first.');
    }

    // Сериализуем данные
    const jsonData = JSON.stringify(data);
    
    // Генерируем случайный IV
    const iv = CryptoJS.lib.WordArray.random(16);
    
    // Шифруем данные с помощью AES
    const encrypted = CryptoJS.AES.encrypt(jsonData, CryptoJS.enc.Hex.parse(this.sessionKeys.aesKey), {
      iv: iv,
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7
    });

    // Генерируем HMAC
    const encryptedBytes = CryptoJS.enc.Base64.parse(encrypted.toString());
    const hmac = CryptoJS.HmacSHA256(encryptedBytes, CryptoJS.enc.Hex.parse(this.sessionKeys.hmacKey));

    return {
      data: encrypted.toString(),
      iv: CryptoJS.enc.Base64.stringify(iv),
      hmac: CryptoJS.enc.Base64.stringify(hmac),
      sessionId: this.sessionId
    };
  }

  /**
   * Расшифровывает HTTP-ответ
   */
  decryptResponse(encryptedResponse: EncryptedResponse): any {
    if (!this.sessionKeys) {
      throw new Error('Session not established');
    }

    try {
      // Проверяем HMAC
      const encryptedBytes = CryptoJS.enc.Base64.parse(encryptedResponse.data);
      const calculatedHmac = CryptoJS.HmacSHA256(encryptedBytes, CryptoJS.enc.Hex.parse(this.sessionKeys.hmacKey));
      const providedHmac = CryptoJS.enc.Base64.parse(encryptedResponse.hmac);      if (CryptoJS.lib.WordArray.create(calculatedHmac.words).toString() !== 
          CryptoJS.lib.WordArray.create(providedHmac.words).toString()) {
        throw new Error('HMAC verification failed');
      }

      // Расшифровываем данные
      const iv = CryptoJS.enc.Base64.parse(encryptedResponse.iv);
      const decrypted = CryptoJS.AES.decrypt(encryptedResponse.data, CryptoJS.enc.Hex.parse(this.sessionKeys.aesKey), {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      });

      const decryptedText = decrypted.toString(CryptoJS.enc.Utf8);
      return JSON.parse(decryptedText);
    } catch (error) {
      console.error('Failed to decrypt response:', error);
      throw new Error('Failed to decrypt response');
    }
  }

  /**
   * Выполняет зашифрованный HTTP-запрос
   */
  async encryptedFetch<T>(
    url: string, 
    options: RequestInit = {}, 
    data?: any
  ): Promise<T> {
    if (!this.sessionId || !this.sessionKeys) {
      throw new Error('Session not established. Call initiateKeyExchange first.');
    }

    let body: string | undefined;
    
    if (data) {
      const encryptedRequest = this.encryptRequest(data);
      body = JSON.stringify(encryptedRequest);
    }

    const response = await fetch(url, {
      ...options,
      method: options.method || 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      body: body,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const responseData = await response.json();

    // Проверяем, является ли ответ зашифрованным
    if (this.isEncryptedResponse(responseData)) {
      return this.decryptResponse(responseData as EncryptedResponse);
    }

    // Если ответ не зашифрован, возвращаем как есть
    return responseData;
  }

  /**
   * Проверяет статус сессии
   */
  async validateSession(baseURL: string = ''): Promise<boolean> {
    if (!this.sessionId) {
      return false;
    }

    try {
      const response = await fetch(`${baseURL}/api/key-exchange/validate/${this.sessionId}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        return false;
      }

      const result = await response.json();
      return result.valid === true;
    } catch (error) {
      console.error('Session validation failed:', error);
      return false;
    }
  }

  /**
   * Отзывает текущую сессию
   */
  async revokeSession(baseURL: string = ''): Promise<boolean> {
    if (!this.sessionId) {
      return true; // Нет сессии для отзыва
    }

    try {
      const response = await fetch(`${baseURL}/api/key-exchange/revoke/${this.sessionId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        this.clearSession();
        return true;
      }

      return false;
    } catch (error) {
      console.error('Session revocation failed:', error);
      return false;
    }
  }

  /**
   * Очищает данные сессии
   */
  clearSession(): void {
    this.sessionId = null;
    this.sessionKeys = null;
  }

  /**
   * Возвращает ID текущей сессии
   */
  getSessionId(): string | null {
    return this.sessionId;
  }

  /**
   * Проверяет, установлена ли сессия
   */
  hasActiveSession(): boolean {
    return this.sessionId !== null && this.sessionKeys !== null;
  }

  /**
   * Деривирует ключи сессии из общего секрета
   */
  private deriveSessionKeys(sharedSecret: string): SessionKeys {
    // Используем HKDF-подобный процесс с SHA-256
    const salt = CryptoJS.enc.Utf8.parse('sleek-chat-salt');
    const info = CryptoJS.enc.Utf8.parse('sleek-chat-session-keys');
    
    // Симулируем HKDF с помощью HMAC
    const secretWordArray = CryptoJS.enc.Hex.parse(sharedSecret);
    const prk = CryptoJS.HmacSHA256(secretWordArray, salt);
    
    // Расширяем ключ до 64 байт (32 для AES + 32 для HMAC)
    const okm1 = CryptoJS.HmacSHA256(info, prk);
    const okm2 = CryptoJS.HmacSHA256(CryptoJS.lib.WordArray.create([0x01]).concat(info), prk);
    
    const expandedKey = okm1.concat(okm2);
    
    // Первые 32 байта для AES, следующие 32 для HMAC
    const aesKey = CryptoJS.lib.WordArray.create(expandedKey.words.slice(0, 8));
    const hmacKey = CryptoJS.lib.WordArray.create(expandedKey.words.slice(8, 16));
    
    return {
      aesKey: aesKey.toString(CryptoJS.enc.Hex),
      hmacKey: hmacKey.toString(CryptoJS.enc.Hex)
    };
  }

  /**
   * Проверяет, является ли ответ зашифрованным
   */
  private isEncryptedResponse(response: any): boolean {
    return response && 
           typeof response.data === 'string' && 
           typeof response.iv === 'string' && 
           typeof response.hmac === 'string';
  }
}
