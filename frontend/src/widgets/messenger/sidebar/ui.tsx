
import * as React from 'react';
import { useState } from 'react';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { ThemeToggle } from '@/shared/ui/theme-toggle';
import { ChatList } from '@/features/messenger/ui/chat-list';
import { UserSearch } from '@/shared/ui/user-search';
import { User } from '@/shared/api/userApi';
import { chatAPI } from '@/shared/api/chatApi';
import { useToast } from '@/hooks/use-toast';
import { usePanelWidth } from '@/hooks/use-panel-width';
import { Users, Edit, Settings, Search, X } from 'lucide-react';

interface Chat {
  id: number;
  name: string;
  lastMessage: string;
  time: string;
  unread: number;
  online: boolean;
  isGroup?: boolean;
  isCreator?: boolean;
}

interface MessengerSidebarProps {
  chats: Chat[];
  selectedChat: number | null;
  onSelectChat: (id: number) => void;
  onShowProfile: () => void;
  onCreateChat?: (chatName: string, isGroup: boolean, participants: number[]) => Promise<any>;
  onRefreshChats?: () => Promise<void>;
  onShowCreateGroup?: () => void;
  onLeaveChat?: (chatId: number) => Promise<void>;
  onDeleteChat?: (chatId: number) => Promise<void>;
  onDeleteGroupChat?: (chatId: number) => Promise<void>;
  currentUserId?: number | null;
  groupUsers?: any[];
  isLoading?: boolean;
}

export const MessengerSidebar = ({ 
  chats, 
  selectedChat, 
  onSelectChat, 
  onShowProfile, 
  onCreateChat, 
  onRefreshChats, 
  onShowCreateGroup, 
  onLeaveChat, 
  onDeleteChat, 
  onDeleteGroupChat, 
  currentUserId,
  groupUsers,
  isLoading 
}: MessengerSidebarProps) => {
  const [showUserSearch, setShowUserSearch] = useState(false);
  const { toast } = useToast();
  const { ref: panelRef, width, isNarrow, isVeryNarrow } = usePanelWidth();

  const handleUserSelect = async (user: User) => {
    try {
      // Используем новый API для создания или получения приватного чата
      const response = await chatAPI.createOrGetPrivateChat(user.id, user.username);
      
      if (response.data?.chat) {
        // Обновляем список чатов
        if (onRefreshChats) {
          await onRefreshChats();
        }
        
        // Если чат был создан или найден, выбираем его
        onSelectChat(response.data.chat.id);
        setShowUserSearch(false);
        
        if (response.data.created) {
          toast({
            title: "Чат создан",
            description: `Создан новый чат с ${user.username}`,
          });
        } else {
          toast({
            title: "Чат найден",
            description: `Открыт существующий чат с ${user.username}`,
          });
        }
      }
    } catch (error) {
      console.error('Ошибка создания/получения чата:', error);
      toast({
        title: "Ошибка",
        description: "Не удалось создать чат",
        variant: "destructive",
      });
    }
  };

  const handleNewMessage = () => {
    setShowUserSearch(true);
  };

  const handleBackToChats = () => {
    setShowUserSearch(false);
  };
  
  const handleCreateGroup = () => {
    if (onShowCreateGroup) {
      onShowCreateGroup();
    }
  };

  return (
    <div ref={panelRef} className="w-full h-full bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col">
      {/* Заголовок */}
      <div className="p-4 border-b border-gray-200 dark:border-gray-700 flex-shrink-0">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-2 min-w-0 flex-1">
            {showUserSearch && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleBackToChats}
                className="flex-shrink-0"
              >
                <X className="h-4 w-4" />
              </Button>
            )}
            {!isVeryNarrow && (
              <h1 className="text-xl font-semibold text-gray-900 dark:text-white truncate">
                {showUserSearch ? 'Новое сообщение' : 'Мессенджер'}
              </h1>
            )}
          </div>
          <div className="flex items-center space-x-2 flex-shrink-0">
            {!isVeryNarrow && <ThemeToggle />}
            <Button
              variant="ghost"
              size="sm"
              onClick={onShowProfile}
              title="Настройки профиля"
            >
              <Settings className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {showUserSearch ? (
        /* Поиск пользователей */
        <div className="flex-1 overflow-y-auto">
          <div className={`${isNarrow ? 'p-2' : 'p-4'}`}>
            <UserSearch 
              onUserSelect={handleUserSelect}
              showEmail={!isVeryNarrow}
            />
          </div>
        </div>
      ) : (
        <>
          {/* Кнопки действий */}
          <div className="p-4 space-y-2 flex-shrink-0">
            <Button 
              variant="outline" 
              className="w-full justify-start min-w-0"
              onClick={handleCreateGroup}
              title="Создать группу"
            >
              <Users className="mr-2 h-4 w-4 flex-shrink-0" />
              {!isNarrow && <span className="truncate">Создать группу</span>}
            </Button>
            <Button 
              variant="outline" 
              className="w-full justify-start min-w-0"
              onClick={handleNewMessage}
              title="Новое сообщение"
            >
              <Edit className="mr-2 h-4 w-4 flex-shrink-0" />
              {!isNarrow && <span className="truncate">Новое сообщение</span>}
            </Button>
          </div>

          {/* Список чатов */}
          {isLoading ? (
            <div className="flex-1 flex items-center justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
          ) : (
            <ChatList 
              chats={chats} 
              selectedChat={selectedChat} 
              onSelectChat={onSelectChat}
              onLeaveChat={onLeaveChat}
              onDeleteChat={onDeleteChat}
              onDeleteGroupChat={onDeleteGroupChat}
              currentUserId={currentUserId}
              isNarrow={isNarrow}
              isVeryNarrow={isVeryNarrow}
            />
          )}
        </>
      )}
    </div>
  );
};
