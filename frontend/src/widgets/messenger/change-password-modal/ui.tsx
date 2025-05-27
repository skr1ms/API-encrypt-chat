import React, { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/shared/ui/dialog';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import { useToast } from '@/hooks/use-toast';
import { chatAPI } from '@/shared/api/chatApi';
import { validatePassword, getPasswordStrength } from '@/shared/lib/validation/password';
import { Loader2, Eye, EyeOff } from 'lucide-react';

interface ChangePasswordModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const ChangePasswordModal: React.FC<ChangePasswordModalProps> = ({
  isOpen,
  onClose
}) => {  const [formData, setFormData] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  });
  const [passwordErrors, setPasswordErrors] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const { toast } = useToast();
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });

    // Валидация пароля в реальном времени
    if (name === 'newPassword') {
      const validation = validatePassword(value);
      setPasswordErrors(validation.errors);
    }

    // Очищаем ошибку при изменении данных
    if (error) setError('');
  };
  const getPasswordStrengthColor = (strength: 'weak' | 'medium' | 'strong') => {
    switch (strength) {
      case 'weak': return 'bg-red-500';
      case 'medium': return 'bg-yellow-500';
      case 'strong': return 'bg-green-500';
    }
  };

  const validateForm = () => {
    if (!formData.currentPassword.trim()) {
      setError('Введите текущий пароль');
      return false;
    }

    if (!formData.newPassword.trim()) {
      setError('Введите новый пароль');
      return false;
    }

    // Валидация нового пароля с использованием тех же правил, что и при регистрации
    const passwordValidation = validatePassword(formData.newPassword);
    if (!passwordValidation.isValid) {
      setError('Новый пароль не соответствует требованиям безопасности');
      return false;
    }

    if (formData.newPassword !== formData.confirmPassword) {
      setError('Пароли не совпадают');
      return false;
    }

    if (formData.currentPassword === formData.newPassword) {
      setError('Новый пароль должен отличаться от текущего');
      return false;
    }

    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!validateForm()) {
      return;
    }

    setIsLoading(true);    try {
      await chatAPI.changePassword(
        formData.currentPassword,
        formData.newPassword
      );

      toast({
        title: "Пароль изменён",
        description: "Ваш пароль успешно обновлён",
      });

      // Сброс формы и закрытие модального окна
      setFormData({
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      });
      onClose();
    } catch (error) {
      console.error('Ошибка смены пароля:', error);
      
      const errorMessage = error instanceof Error ? error.message : 'Не удалось изменить пароль';
      
      // Обрабатываем специфичные ошибки
      if (errorMessage.includes('INVALID_CURRENT_PASSWORD') || errorMessage.includes('Incorrect current password')) {
        setError('Неверный текущий пароль');
      } else if (errorMessage.includes('PASSWORD_TOO_WEAK')) {
        setError('Пароль слишком слабый. Используйте более сложный пароль');
      } else {
        setError(errorMessage);
      }
    } finally {
      setIsLoading(false);
    }
  };
  const handleClose = () => {
    setFormData({
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    });
    setPasswordErrors([]);
    setError('');
    onClose();
  };

  const passwordStrength = getPasswordStrength(formData.newPassword);

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Изменить пароль</DialogTitle>
        </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4" autoComplete="off">
          {/* Скрытое поле для предотвращения автозаполнения */}
          <input type="password" style={{ display: 'none' }} />
          
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-md text-sm">
              {error}
            </div>
          )}<div>
            <Label htmlFor="currentPassword">Текущий пароль</Label>            <Input
              id="currentPassword"
              name="currentPassword"
              type="password"
              value={formData.currentPassword}
              onChange={handleInputChange}
              required
              disabled={isLoading}
              className="mt-1"
              placeholder="Введите текущий пароль"
              autoComplete="current-password"
              data-form-type="password"
              data-lpignore="true"
            />
          </div><div>
            <Label htmlFor="newPassword">Новый пароль</Label>            <Input
              id="newPassword"
              name="newPassword"
              type="password"
              value={formData.newPassword}
              onChange={handleInputChange}
              required
              disabled={isLoading}
              className="mt-1"
              placeholder="Введите новый пароль"
              autoComplete="new-password"
            />
            
            {/* Индикатор силы пароля */}
            {formData.newPassword && (
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
                    {passwordStrength === 'weak' ? 'Слабый' :
                     passwordStrength === 'medium' ? 'Средний' : 'Сильный'}
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
          </div>          <div>
            <Label htmlFor="confirmPassword">Подтвердите новый пароль</Label>            <Input
              id="confirmPassword"
              name="confirmPassword"
              type="password"
              value={formData.confirmPassword}
              onChange={handleInputChange}
              required
              disabled={isLoading}
              className="mt-1"
              placeholder="Повторите новый пароль"
              autoComplete="new-password"
            />
            
            {/* Проверка совпадения паролей */}
            {formData.confirmPassword && formData.newPassword !== formData.confirmPassword && (
              <div className="text-xs text-red-500 mt-1">
                Пароли не совпадают
              </div>
            )}
          </div>

          <div className="flex space-x-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={isLoading}
              className="flex-1"
            >
              Отмена
            </Button>            <Button
              type="submit"
              disabled={isLoading || passwordErrors.length > 0 || formData.newPassword !== formData.confirmPassword || !formData.currentPassword || !formData.newPassword || !formData.confirmPassword}
              className="flex-1"
            >
              {isLoading ? (
                <div className="flex items-center space-x-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>Изменение...</span>
                </div>
              ) : (
                'Изменить пароль'
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>  );
};
