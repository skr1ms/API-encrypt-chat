
import * as React from 'react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import { loginStart, loginSuccess, loginFailure } from '@/shared/store/slices/authSlice';
import { chatAPI } from '@/shared/api/chatApi';
import { ECDHService } from '@/shared/lib/crypto/ecdh';
import { ECDSAService } from '@/shared/lib/crypto/ecdsa';
import { RSAService } from '@/shared/lib/crypto/rsa';
import { websocketService } from '@/shared/lib/websocket/websocketService';
import { getErrorMessage } from '@/shared/lib/errors/errorMessages';
import { Loader2 } from 'lucide-react';

export const LoginForm = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
    // Очищаем ошибку при изменении данных
    if (error) setError('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);
    
    dispatch(loginStart());

    try {
      // Валидация формы
      if (!formData.username.trim() || !formData.password.trim()) {
        setError('Please fill in all required fields');
        setIsLoading(false);
        return;
      }

      // Generate ECDH key pair for secure communication
      const ecdhKeyPair = ECDHService.generateKeyPair();

      // Generate ECDSA key pair for authentication
      const ecdsaKeyPair = ECDSAService.generateStaticKeyPair();

      // Generate RSA key pair for encryption
      const rsaKeyPair = await RSAService.generateKeyPair();

      const response = await chatAPI.login({
        username: formData.username,
        password: formData.password,
        ecdhPublicKey: ecdhKeyPair.publicKey,
        ecdsaPublicKey: ecdsaKeyPair.publicKey,
        rsaPublicKey: rsaKeyPair.publicKey,
      });

      console.log('Login response:', response);

      const token = response.token;
      const user = response.user;

      console.log('Token to save:', token);
      console.log('User data:', user);

      if (!token) {
        throw new Error('No token received from server');
      }

      if (!user) {
        throw new Error('No user data received from server');
      }

      // Store token
      localStorage.setItem('token', token);
      localStorage.setItem('ecdhPrivateKey', ecdhKeyPair.privateKey);
      localStorage.setItem('ecdsaPrivateKey', ecdsaKeyPair.privateKey);
      localStorage.setItem('rsaPrivateKey', rsaKeyPair.privateKey);

      console.log('Token saved to localStorage:', localStorage.getItem('token'));

      // Handle different user data formats from backend
      const userId = typeof user === 'object' && user && 'id' in user ? String((user as any).id) : '';
      const username = typeof user === 'object' && user && 'username' in user ? (user as any).username : '';
      const publicKey = typeof user === 'object' && user && 'ecdsa_public_key' in user 
        ? (user as any).ecdsa_public_key 
        : typeof user === 'object' && user && 'publicKey' in user 
          ? (user as any).publicKey 
          : '';
      
      dispatch(loginSuccess({
        user: {
          id: userId,
          username: username,
          publicKey: publicKey,
        },
        privateKey: ecdhKeyPair.privateKey,
        publicKey: ecdhKeyPair.publicKey,
      }));

      // Connect to WebSocket
      websocketService.connect(token);

      navigate('/messenger');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Login failed';
      const userFriendlyError = getErrorMessage(errorMessage);
      setError(userFriendlyError);
      dispatch(loginFailure(userFriendlyError));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-md">
          {error}
        </div>
      )}
      
      <div>
        <Label htmlFor="username">Логин</Label>
        <Input
          id="username"
          name="username"
          type="text"
          value={formData.username}
          onChange={handleInputChange}
          required
          disabled={isLoading}
          className="mt-1"
          placeholder="Введите ваш логин"
        />
      </div>
      
      <div>
        <Label htmlFor="password">Пароль</Label>
        <Input
          id="password"
          name="password"
          type="password"
          value={formData.password}
          onChange={handleInputChange}
          required
          disabled={isLoading}
          className="mt-1"
          placeholder="Введите ваш пароль"
        />
      </div>
      
      <Button 
        type="submit" 
        className="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50" 
        disabled={isLoading}
      >
        {isLoading ? (
          <div className="flex items-center space-x-2">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Signing in...</span>
          </div>
        ) : (
          'Sign In'
        )}
      </Button>
    </form>
  );
};
