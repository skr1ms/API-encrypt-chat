
import * as React from 'react';
import { useState, useCallback, useEffect } from 'react';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { ThemeToggle } from '@/shared/ui/theme-toggle';
import { ChatList } from '@/features/messenger/ui/chat-list';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { chatAPI } from '@/shared/api/chatApi';
import { Search, Users, Edit, Settings, MessageCircle, X } from 'lucide-react';

interface Chat {
  id: number;
  name: string;
  lastMessage: string;
  time: string;
  unread: number;
  online: boolean;
  isGroup?: boolean;
}

interface User {
  id: number;
  username: string;
  email: string;
  online: boolean;
}

interface MessengerSidebarProps {
  chats: Chat[];
  selectedChat: number | null;
  onSelectChat: (id: number) => void;
  onShowProfile: () => void;
}

export const MessengerSidebar = ({ chats, selectedChat, onSelectChat, onShowProfile }: MessengerSidebarProps) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [searchMode, setSearchMode] = useState(false);

  const handleSearch = useCallback(async (query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      setSearchMode(false);
      return;
    }

    setIsSearching(true);
    setSearchMode(true);
    
    try {
      const results = await chatAPI.searchUsers(query);
      setSearchResults(results);
    } catch (error) {
      console.error('Ошибка поиска:', error);
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  }, []);

  useEffect(() => {
    const delayedSearch = setTimeout(() => {
      handleSearch(searchQuery);
    }, 300);

    return () => clearTimeout(delayedSearch);
  }, [searchQuery, handleSearch]);

  const handleCreateChat = async (user: User) => {
    try {
      await chatAPI.createChat(user.username, false, [user.id]);
      setSearchQuery('');
      setSearchMode(false);
      // Здесь можно добавить обновление списка чатов
    } catch (error) {
      console.error('Ошибка создания чата:', error);
    }
  };

  const clearSearch = () => {
    setSearchQuery('');
    setSearchMode(false);
    setSearchResults([]);
  };

  return (
    <div className="w-80 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col">
      {/* Заголовок */}
      <div className="p-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-xl font-semibold text-gray-900 dark:text-white">Мессенджер</h1>
          <div className="flex items-center space-x-2">
            <ThemeToggle />
            <Button
              variant="ghost"
              size="sm"
              onClick={onShowProfile}
            >
              <Settings className="h-4 w-4" />
            </Button>
          </div>
        </div>
        
        {/* Поиск */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
          <Input
            placeholder="Поиск пользователей..."
            className="pl-10 pr-10"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          {searchMode && (
            <button
              onClick={clearSearch}
              className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      {searchMode ? (
        /* Результаты поиска */
        <div className="flex-1 overflow-y-auto">
          <div className="p-4">
            <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-3">
              Результаты поиска
            </h3>
            {isSearching ? (
              <div className="text-center py-4">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500 mx-auto"></div>
                <p className="text-sm text-gray-500 mt-2">Поиск...</p>
              </div>
            ) : searchResults.length > 0 ? (
              <div className="space-y-2">
                {searchResults.map((user) => (
                  <div
                    key={user.id}
                    className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
                  >
                    <div className="flex items-center space-x-3">
                      <Avatar className="h-10 w-10">
                        <AvatarFallback className="bg-blue-500 text-white">
                          {user.username.charAt(0).toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <p className="font-medium text-gray-900 dark:text-white">
                          {user.username}
                        </p>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          {user.email}
                        </p>
                      </div>
                    </div>
                    <Button
                      size="sm"
                      onClick={() => handleCreateChat(user)}
                      className="text-xs"
                    >
                      <MessageCircle className="h-3 w-3 mr-1" />
                      Чат
                    </Button>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <Users className="h-12 w-12 text-gray-400 mx-auto mb-3" />
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Пользователи не найдены
                </p>
                <p className="text-xs text-gray-400 mt-1">
                  Попробуйте другой запрос
                </p>
              </div>
            )}
          </div>
        </div>
      ) : (
        <>
          {/* Кнопки действий */}
          <div className="p-4 space-y-2">
            <Button variant="outline" className="w-full justify-start">
              <Users className="mr-2 h-4 w-4" />
              Создать группу
            </Button>
            <Button variant="outline" className="w-full justify-start">
              <Edit className="mr-2 h-4 w-4" />
              Новое сообщение
            </Button>
          </div>

          {/* Список чатов */}
          <ChatList chats={chats} selectedChat={selectedChat} onSelectChat={onSelectChat} />
        </>
      )}
    </div>
  );
};
