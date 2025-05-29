import { useState, useEffect, useCallback } from 'react';
import { SecureApiClient, createSecureApiClient } from '../api/secure-client';

interface UseSecureApiOptions {
  baseURL?: string;
  userId?: number;
  autoInit?: boolean;
}

interface UseSecureApiReturn {
  client: SecureApiClient | null;
  isInitialized: boolean;
  isLoading: boolean;
  error: string | null;
  initializeEncryption: (userId: number) => Promise<boolean>;
  refreshEncryption: () => Promise<boolean>;
  validateSession: () => Promise<boolean>;
  logout: () => Promise<void>;
  clearError: () => void;
}

export const useSecureApi = (options: UseSecureApiOptions = {}): UseSecureApiReturn => {
  const { 
    baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
    userId,
    autoInit = false
  } = options;

  const [client, setClient] = useState<SecureApiClient | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Создаем клиент при монтировании компонента
  useEffect(() => {
    const apiClient = createSecureApiClient({
      baseURL,
      userId,
      autoInitEncryption: false // Инициализируем вручную для лучшего контроля
    });
    setClient(apiClient);
  }, [baseURL, userId]);

  // Автоматическая инициализация при наличии userId
  useEffect(() => {
    if (autoInit && userId && client && !isInitialized && !isLoading) {
      initializeEncryption(userId);
    }
  }, [autoInit, userId, client, isInitialized, isLoading]);

  const initializeEncryption = useCallback(async (targetUserId: number): Promise<boolean> => {
    if (!client) {
      setError('API client not available');
      return false;
    }

    setIsLoading(true);
    setError(null);

    try {
      const success = await client.initializeEncryption(targetUserId);
      
      if (success) {
        setIsInitialized(true);
        console.log('Encryption initialized successfully');
      } else {
        setError('Failed to initialize encryption');
      }

      return success;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
      setError(`Encryption initialization failed: ${errorMessage}`);
      return false;
    } finally {
      setIsLoading(false);
    }
  }, [client]);

  const refreshEncryption = useCallback(async (): Promise<boolean> => {
    if (!client) {
      setError('API client not available');
      return false;
    }

    if (!isInitialized) {
      setError('Encryption not initialized');
      return false;
    }

    setIsLoading(true);
    setError(null);

    try {
      const success = await client.refreshEncryption();
      
      if (!success) {
        setError('Failed to refresh encryption');
        setIsInitialized(false);
      }

      return success;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
      setError(`Encryption refresh failed: ${errorMessage}`);
      setIsInitialized(false);
      return false;
    } finally {
      setIsLoading(false);
    }
  }, [client, isInitialized]);

  const validateSession = useCallback(async (): Promise<boolean> => {
    if (!client) {
      return false;
    }

    try {
      const isValid = await client.validateSession();
      
      if (!isValid) {
        setIsInitialized(false);
      }

      return isValid;
    } catch (err) {
      console.error('Session validation failed:', err);
      setIsInitialized(false);
      return false;
    }
  }, [client]);

  const logout = useCallback(async (): Promise<void> => {
    if (!client) {
      return;
    }

    setIsLoading(true);
    
    try {
      await client.logout();
      setIsInitialized(false);
      setError(null);
    } catch (err) {
      console.error('Logout failed:', err);
      // Очищаем состояние даже если logout не удался
      setIsInitialized(false);
    } finally {
      setIsLoading(false);
    }
  }, [client]);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return {
    client,
    isInitialized,
    isLoading,
    error,
    initializeEncryption,
    refreshEncryption,
    validateSession,
    logout,
    clearError
  };
};

// Хук для проверки статуса шифрования
export const useEncryptionStatus = (client: SecureApiClient | null) => {
  const [status, setStatus] = useState({
    hasActiveSession: false,
    sessionId: null as string | null,
    userId: null as number | null,
  });

  useEffect(() => {
    if (client) {
      const sessionInfo = client.getSessionInfo();
      setStatus({
        hasActiveSession: sessionInfo.hasActiveSession,
        sessionId: sessionInfo.sessionId,
        userId: sessionInfo.userId,
      });
    }
  }, [client]);

  return status;
};
