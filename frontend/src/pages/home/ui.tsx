
import React from 'react';
import { Link } from 'react-router-dom';
import { Button } from '@/shared/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card';
import { MessageSquare } from 'lucide-react';

export const HomePage = () => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800 flex items-center justify-center p-4">
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <div className="bg-blue-500 p-3 rounded-full">
              <MessageSquare className="h-8 w-8 text-white" />
            </div>
          </div>
          <CardTitle className="text-3xl font-bold text-blue-600 dark:text-blue-400">
            Добро пожаловать
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 gap-4">
            <Link to="/login" className="w-full">
              <Button className="w-full bg-blue-600 hover:bg-blue-700 text-lg py-6">
                Войти в аккаунт
              </Button>
            </Link>
            <Link to="/register" className="w-full">
              <Button variant="outline" className="w-full text-lg py-6">
                Создать аккаунт
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
