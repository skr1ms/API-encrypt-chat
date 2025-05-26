
import React from 'react';
import { Link } from 'react-router-dom';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card';
import { LoginForm } from '@/features/auth/ui/login-form';
import { ThemeToggle } from '@/shared/ui/theme-toggle';

export const LoginPage = () => {
  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center relative">
          <div className="absolute top-4 right-4">
            <ThemeToggle />
          </div>
          <CardTitle className="text-2xl font-bold text-blue-600 dark:text-blue-400">
            Вход в аккаунт
          </CardTitle>
          <CardDescription>
            Введите ваши данные для входа
          </CardDescription>
        </CardHeader>
        <CardContent>
          <LoginForm />
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Нет аккаунта?{' '}
              <Link to="/register" className="text-blue-600 hover:text-blue-700 font-medium">
                Зарегистрироваться
              </Link>
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
