import React, { createContext, useContext, useEffect, useState } from 'react';
import { SecureApiClient } from '../api/secure-client';
import { useSecureApi } from '../hooks/use-secure-api';

interface EncryptionContextType {
  client: SecureApiClient | null;
  isInitialized: boolean;
  isLoading: boolean;
  error: string | null;
  userId: number | null;
  initializeEncryption: (userId: number) => Promise<boolean>;
  refreshEncryption: () => Promise<boolean>;
  validateSession: () => Promise<boolean>;
  logout: () => Promise<void>;
  clearError: () => void;
}

const EncryptionContext = createContext<EncryptionContextType | null>(null);

interface EncryptionProviderProps {
  children: React.ReactNode;
  userId?: number | null;
  autoInit?: boolean;
}

export const EncryptionProvider: React.FC<EncryptionProviderProps> = ({
  children,
  userId = null,
  autoInit = false
}) => {
  const {
    client,
    isInitialized,
    isLoading,
    error,
    initializeEncryption,
    refreshEncryption,
    validateSession,
    logout,
    clearError
  } = useSecureApi({
    userId: userId || undefined,
    autoInit: autoInit && !!userId
  });

  const [currentUserId, setCurrentUserId] = useState<number | null>(userId);

  // Обновляем userId когда он изменяется
  useEffect(() => {
    setCurrentUserId(userId);
  }, [userId]);

  // Автоматическая инициализация при смене пользователя
  useEffect(() => {
    if (autoInit && userId && userId !== currentUserId && !isInitialized && !isLoading) {
      initializeEncryption(userId);
      setCurrentUserId(userId);
    }
  }, [autoInit, userId, currentUserId, isInitialized, isLoading, initializeEncryption]);

  const contextValue: EncryptionContextType = {
    client,
    isInitialized,
    isLoading,
    error,
    userId: currentUserId,
    initializeEncryption,
    refreshEncryption,
    validateSession,
    logout,
    clearError
  };

  return (
    <EncryptionContext.Provider value={contextValue}>
      {children}
    </EncryptionContext.Provider>
  );
};

export const useEncryption = (): EncryptionContextType => {
  const context = useContext(EncryptionContext);
  
  if (!context) {
    throw new Error('useEncryption must be used within an EncryptionProvider');
  }
  
  return context;
};

// Хук для безопасных API вызовов с автоматическим шифрованием
export const useSecureRequest = () => {
  const { client, isInitialized, error } = useEncryption();

  const secureRequest = async <T = any>(
    method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH',
    url: string,
    data?: any,
    config?: any
  ): Promise<T> => {
    if (!client) {
      throw new Error('Secure API client not available');
    }

    if (!isInitialized) {
      throw new Error('Encryption not initialized');
    }

    if (error) {
      throw new Error(`Encryption error: ${error}`);
    }

    try {
      switch (method) {
        case 'GET':
          return await client.get<T>(url, config);
        case 'POST':
          return await client.post<T>(url, data, config);
        case 'PUT':
          return await client.put<T>(url, data, config);
        case 'DELETE':
          return await client.delete<T>(url, config);
        case 'PATCH':
          return await client.patch<T>(url, data, config);
        default:
          throw new Error(`Unsupported HTTP method: ${method}`);
      }
    } catch (err) {
      console.error(`Secure ${method} request failed:`, err);
      throw err;
    }
  };

  return {
    secureRequest,
    isReady: isInitialized && !error && !!client,
    error
  };
};

export default EncryptionProvider;
