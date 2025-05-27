import * as React from 'react';
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { PasswordInput } from '@/components/ui/password-input';
import { Label } from '@/shared/ui/label';
import { chatAPI } from '@/shared/api/chatApi';
import { ECDSAService } from '@/shared/lib/crypto/ecdsa';
import { RSAService } from '@/shared/lib/crypto/rsa';
import { validatePassword, getPasswordStrength } from '@/shared/lib/validation/password';
import { getErrorMessage } from '@/shared/lib/errors/errorMessages';
import { CheckCircle, Loader2 } from 'lucide-react';

export const RegisterForm = () => {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
  });
  const [passwordErrors, setPasswordErrors] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [countdown, setCountdown] = useState(0);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });

    // Валидация пароля в реальном времени
    if (name === 'password') {
      const validation = validatePassword(value);
      setPasswordErrors(validation.errors);
    }

    // Очищаем ошибки при изменении полей
    if (error) setError('');
  };

  const getPasswordStrengthColor = (strength: 'weak' | 'medium' | 'strong') => {
    switch (strength) {
      case 'weak': return 'bg-red-500';
      case 'medium': return 'bg-yellow-500';
      case 'strong': return 'bg-green-500';
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Валидация формы
    if (!formData.username.trim() || !formData.email.trim() || !formData.password || !formData.confirmPassword) {
      setError('Please fill in all required fields');
      return;
    }

    // Валидация пароля
    const passwordValidation = validatePassword(formData.password);
    if (!passwordValidation.isValid) {
      setError('Password does not meet security requirements');
      return;
    }
    
    if (formData.password !== formData.confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    setIsLoading(true);

    try {
      // Generate ECDSA key pair for authentication
      const ecdsaKeyPair = ECDSAService.generateStaticKeyPair();

      // Generate RSA key pair for encryption
      const rsaKeyPair = await RSAService.generateKeyPair();

      const response = await chatAPI.register({
        username: formData.username,
        email: formData.email,
        password: formData.password,
        ecdsaPublicKey: ecdsaKeyPair.publicKey,
        rsaPublicKey: rsaKeyPair.publicKey,
      });

      // Успешная регистрация - показываем сообщение и запускаем таймер
      setSuccessMessage('Registration successful! Redirecting to login page...');
      setCountdown(5);
      
      // Запускаем таймер обратного отсчета
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer);
            navigate('/login');
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Registration failed';
      setError(getErrorMessage(errorMessage));
    } finally {
      setIsLoading(false);
    }
  };

  const passwordStrength = getPasswordStrength(formData.password);

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-md">
          {error}
        </div>
      )}

      {successMessage && (
        <div className="bg-green-50 border border-green-200 text-green-600 px-4 py-3 rounded-md flex items-center space-x-2">
          <CheckCircle className="h-5 w-5" />
          <div>
            <p>{successMessage}</p>
            {countdown > 0 && (
              <p className="text-sm">Redirecting in {countdown} seconds...</p>
            )}
          </div>
        </div>
      )}
      
      <div>
        <Label htmlFor="username">Имя пользователя</Label>
        <Input
          id="username"
          name="username"
          type="text"
          value={formData.username}
          onChange={handleInputChange}
          required
          className="mt-1"
          placeholder="Введите имя пользователя"
        />
      </div>

      <div>
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          name="email"
          type="email"
          value={formData.email}
          onChange={handleInputChange}
          required
          className="mt-1"
          placeholder="Введите email"
        />
      </div>

      <div>
        <Label htmlFor="password">Пароль</Label>
        <PasswordInput
          id="password"
          name="password"
          value={formData.password}
          onChange={handleInputChange}
          required
          className="mt-1"
          placeholder="Введите пароль"
        />
        
        {/* Индикатор силы пароля */}
        {formData.password && (
          <div className="mt-2">
            <div className="flex items-center space-x-2 mb-1">
              <div className="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
                <div 
                  className={`h-full transition-all duration-300 ${getPasswordStrengthColor(passwordStrength)}`}
                  style={{ 
                    width: passwordStrength === 'weak' ? '33%' : 
                           passwordStrength === 'medium' ? '66%' : '100%' 
                  }}
                />
              </div>
              <span className={`text-xs font-medium ${
                passwordStrength === 'weak' ? 'text-red-500' :
                passwordStrength === 'medium' ? 'text-yellow-500' : 'text-green-500'
              }`}>
                {passwordStrength === 'weak' ? 'Weak' :
                 passwordStrength === 'medium' ? 'Medium' : 'Strong'}
              </span>
            </div>
            
            {/* Ошибки валидации пароля */}
            {passwordErrors.length > 0 && (
              <div className="text-xs text-red-500 space-y-1">
                {passwordErrors.map((error, index) => (
                  <div key={index}>• {error}</div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      <div>
        <Label htmlFor="confirmPassword">Подтвердите пароль</Label>
        <PasswordInput
          id="confirmPassword"
          name="confirmPassword"
          value={formData.confirmPassword}
          onChange={handleInputChange}
          required
          className="mt-1"
          placeholder="Повторите пароль"
        />
        
        {/* Проверка совпадения паролей */}
        {formData.confirmPassword && formData.password !== formData.confirmPassword && (
          <div className="text-xs text-red-500 mt-1">
            Passwords do not match
          </div>
        )}
      </div>

      <Button 
        type="submit" 
        className="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50"
        disabled={isLoading || passwordErrors.length > 0 || formData.password !== formData.confirmPassword || successMessage !== ''}
      >
        {isLoading ? (
          <div className="flex items-center space-x-2">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Creating account...</span>
          </div>
        ) : (
          'Create Account'
        )}
      </Button>
    </form>
  );
};
