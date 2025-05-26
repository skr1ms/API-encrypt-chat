
import * as React from 'react';
import { useState } from 'react';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Button } from '@/shared/ui/button';
import { MessengerSidebar } from '@/widgets/messenger/sidebar';
import { MessageList } from '@/features/messenger/ui/message-list';
import { MessageInput } from '@/features/messenger/ui/message-input';
import { GroupSettingsModal } from '@/widgets/messenger/group-settings-modal';
import { MessageSquare, Users } from 'lucide-react';

export const MessengerPage = () => {
  const [selectedChat, setSelectedChat] = useState<number | null>(null);
  const [message, setMessage] = useState('');
  const [showProfile, setShowProfile] = useState(false);
  const [showGroupSettings, setShowGroupSettings] = useState(false);

  // Пустой массив чатов - пользователь будет добавлять их сам
  const chats: any[] = [];

  const groupUsers: any[] = [];

  const messages: any[] = [];

  const handleSendMessage = () => {
    if (message.trim()) {
      console.log('Отправка сообщения:', message);
      setMessage('');
    }
  };

  const handleChatHeaderClick = () => {
    const currentChat = chats.find(c => c.id === selectedChat);
    if (currentChat && currentChat.isGroup) {
      setShowGroupSettings(true);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex">
      {/* Боковая панель */}
      <MessengerSidebar 
        chats={chats}
        selectedChat={selectedChat}
        onSelectChat={setSelectedChat}
        onShowProfile={() => setShowProfile(!showProfile)}
      />

      {/* Основная область чата */}
      <div className="flex-1 flex flex-col">
        {selectedChat ? (
          <>
            {/* Заголовок чата */}
            <div 
              className={`bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-4 ${
                chats.find(c => c.id === selectedChat)?.isGroup ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700' : ''
              }`}
              onClick={handleChatHeaderClick}
            >
              <div className="flex items-center space-x-3">
                <Avatar className="h-8 w-8">
                  <AvatarFallback className="bg-blue-500 text-white text-sm">
                    {chats.find(c => c.id === selectedChat)?.isGroup ? 
                      <Users className="h-4 w-4" /> : 
                      chats.find(c => c.id === selectedChat)?.name.charAt(0)
                    }
                  </AvatarFallback>
                </Avatar>
                <div>
                  <h2 className="font-medium text-gray-900 dark:text-white">
                    {chats.find(c => c.id === selectedChat)?.name}
                  </h2>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {chats.find(c => c.id === selectedChat)?.isGroup 
                      ? `${groupUsers.filter(u => u.online).length} участников в сети`
                      : chats.find(c => c.id === selectedChat)?.online ? 'в сети' : 'был в сети недавно'
                    }
                  </p>
                </div>
              </div>
            </div>

            {/* Сообщения */}
            <MessageList messages={messages} />

            {/* Поле ввода */}
            <MessageInput 
              message={message}
              onMessageChange={setMessage}
              onSendMessage={handleSendMessage}
            />
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center bg-gray-50 dark:bg-gray-900">
            <div className="text-center">
              <MessageSquare className="h-16 w-16 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
                Добро пожаловать в мессенджер!
              </h3>
              <p className="text-gray-500 dark:text-gray-400 mb-4">
                У вас пока нет чатов. Найдите пользователей через поиск в боковой панели.
              </p>
              <div className="text-sm text-gray-400 space-y-1">
                <p>• Используйте поиск по логину или email</p>
                <p>• Создавайте групповые чаты</p>
                <p>• Все сообщения зашифрованы</p>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Панель профиля */}
      {showProfile && (
        <div className="w-80 bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 p-4">
          <div className="text-center">
            <Avatar className="h-20 w-20 mx-auto mb-4">
              <AvatarFallback className="bg-blue-500 text-white text-xl">И</AvatarFallback>
            </Avatar>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white">Иван Иванов</h3>
            <p className="text-sm text-gray-500 dark:text-gray-400">@ivanov</p>
          </div>
          
          <div className="mt-6 space-y-4">
            <Button variant="outline" className="w-full">
              Редактировать профиль
            </Button>
            <Button variant="outline" className="w-full">
              Настройки
            </Button>
            <Button variant="outline" className="w-full text-red-600 hover:text-red-700">
              Выйти
            </Button>
          </div>
        </div>
      )}

      {/* Модальное окно настроек группы */}
      <GroupSettingsModal
        isOpen={showGroupSettings}
        onClose={() => setShowGroupSettings(false)}
        groupName={chats.find(c => c.id === selectedChat)?.name || ''}
        users={groupUsers}
      />
    </div>
  );
};
