import React, { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Loader2, Shield, ShieldCheck, ShieldX, RefreshCw } from 'lucide-react';
import { useSecureApi, useEncryptionStatus } from '@/shared/hooks/use-secure-api';

interface EncryptionManagerProps {
  userId: number;
  onEncryptionReady?: () => void;
  onEncryptionError?: (error: string) => void;
  autoInit?: boolean;
}

export const EncryptionManager: React.FC<EncryptionManagerProps> = ({
  userId,
  onEncryptionReady,
  onEncryptionError,
  autoInit = true
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
    userId,
    autoInit
  });

  const encryptionStatus = useEncryptionStatus(client);
  const [lastValidation, setLastValidation] = useState<Date | null>(null);

  // Вызываем коллбэки при изменении состояния
  useEffect(() => {
    if (isInitialized && onEncryptionReady) {
      onEncryptionReady();
    }
  }, [isInitialized, onEncryptionReady]);

  useEffect(() => {
    if (error && onEncryptionError) {
      onEncryptionError(error);
    }
  }, [error, onEncryptionError]);

  // Периодическая проверка сессии
  useEffect(() => {
    if (!isInitialized) return;

    const interval = setInterval(async () => {
      const isValid = await validateSession();
      setLastValidation(new Date());
      
      if (!isValid) {
        console.warn('Session validation failed, encryption may be invalid');
      }
    }, 30000); // Проверяем каждые 30 секунд

    return () => clearInterval(interval);
  }, [isInitialized, validateSession]);

  const handleInitialize = async () => {
    await initializeEncryption(userId);
  };

  const handleRefresh = async () => {
    await refreshEncryption();
  };

  const handleValidate = async () => {
    const isValid = await validateSession();
    setLastValidation(new Date());
    
    if (isValid) {
      clearError();
    }
  };

  const getStatusIcon = () => {
    if (isLoading) {
      return <Loader2 className="h-4 w-4 animate-spin" />;
    }
    
    if (error) {
      return <ShieldX className="h-4 w-4 text-destructive" />;
    }
    
    if (isInitialized) {
      return <ShieldCheck className="h-4 w-4 text-green-600" />;
    }
    
    return <Shield className="h-4 w-4 text-muted-foreground" />;
  };

  const getStatusText = () => {
    if (isLoading) return 'Initializing...';
    if (error) return 'Error';
    if (isInitialized) return 'Secure';
    return 'Not Initialized';
  };

  const getStatusVariant = (): "default" | "secondary" | "destructive" | "outline" => {
    if (error) return 'destructive';
    if (isInitialized) return 'default';
    return 'secondary';
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {getStatusIcon()}
          End-to-End Encryption
          <Badge variant={getStatusVariant()}>
            {getStatusText()}
          </Badge>
        </CardTitle>
        <CardDescription>
          Secure communication using ECDH key exchange and AES-256 encryption
        </CardDescription>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <ShieldX className="h-4 w-4" />
            <AlertDescription>
              {error}
            </AlertDescription>
          </Alert>
        )}

        {isInitialized && (
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Session ID:</span>
              <span className="font-mono text-xs">
                {encryptionStatus.sessionId?.substring(0, 16)}...
              </span>
            </div>
            
            {lastValidation && (
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Last validated:</span>
                <span className="text-xs">
                  {lastValidation.toLocaleTimeString()}
                </span>
              </div>
            )}
          </div>
        )}

        <div className="flex flex-wrap gap-2">
          {!isInitialized ? (
            <Button 
              onClick={handleInitialize}
              disabled={isLoading}
              className="flex items-center gap-2"
            >
              {isLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Shield className="h-4 w-4" />
              )}
              Initialize Encryption
            </Button>
          ) : (
            <>
              <Button 
                variant="outline"
                onClick={handleRefresh}
                disabled={isLoading}
                className="flex items-center gap-2"
              >
                <RefreshCw className="h-4 w-4" />
                Refresh Keys
              </Button>
              
              <Button 
                variant="outline"
                onClick={handleValidate}
                disabled={isLoading}
                className="flex items-center gap-2"
              >
                <ShieldCheck className="h-4 w-4" />
                Validate Session
              </Button>
              
              <Button 
                variant="destructive"
                onClick={logout}
                disabled={isLoading}
                className="flex items-center gap-2"
              >
                <ShieldX className="h-4 w-4" />
                End Session
              </Button>
            </>
          )}
        </div>

        {error && (
          <Button 
            variant="ghost" 
            size="sm"
            onClick={clearError}
            className="w-full"
          >
            Clear Error
          </Button>
        )}
      </CardContent>
    </Card>
  );
};

export default EncryptionManager;
