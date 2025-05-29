
import * as React from 'react';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Badge } from '@/shared/ui/badge';
import { Users } from 'lucide-react';
import { ChatSidebarMenu } from '@/widgets/messenger/chat-sidebar-menu';

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

interface ChatListProps {
  chats: Chat[];
  selectedChat: number | null;
  onSelectChat: (id: number) => void;
  onLeaveChat?: (chatId: number) => Promise<void>;
  onDeleteChat?: (chatId: number) => Promise<void>;
  onDeleteGroupChat?: (chatId: number) => Promise<void>;
  currentUserId?: number | null;
  isNarrow?: boolean;
  isVeryNarrow?: boolean;
}

export const ChatList = ({ chats, selectedChat, onSelectChat, onLeaveChat, onDeleteChat, onDeleteGroupChat, currentUserId, isNarrow = false, isVeryNarrow = false }: ChatListProps) => {
  console.log('ChatList rendered with chats:', chats);
  console.log('ChatList chats length:', chats.length);
  
  return (
    <div className="flex-1 overflow-y-auto telegram-scrollbar">
      {chats.length === 0 ? (
        <div className="p-4 text-center text-gray-500 dark:text-gray-400">
          <p>Нет чатов</p>
        </div>
      ) : (
        chats.map((chat) => {
          console.log('Rendering chat:', chat);
          return (
            <div
              key={chat.id}
              className={`group relative ${isVeryNarrow ? 'p-2' : 'p-4'} cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700 border-b border-gray-100 dark:border-gray-700 ${
                selectedChat === chat.id ? 'bg-blue-50 dark:bg-blue-900/20' : ''
              }`}
              onClick={() => onSelectChat(chat.id)}
              title={isVeryNarrow ? `${chat.name}: ${chat.lastMessage}` : undefined}
            >
            <div className="flex items-center space-x-3">
            <div className="relative flex-shrink-0">
              <Avatar className={`${isVeryNarrow ? 'h-8 w-8' : 'h-10 w-10'}`}>
                <AvatarFallback className="bg-blue-500 text-white">
                  {chat.isGroup ? <Users className={`${isVeryNarrow ? 'h-4 w-4' : 'h-5 w-5'}`} /> : chat.name.charAt(0)}
                </AvatarFallback>
              </Avatar>
              {chat.online && !chat.isGroup && (
                <div className="absolute bottom-0 right-0 h-3 w-3 bg-green-500 border-2 border-white dark:border-gray-800 rounded-full"></div>
              )}
            </div>
            {!isVeryNarrow && (
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between gap-2">
                  <p className="text-sm font-medium text-gray-900 dark:text-white truncate flex-1">
                    {chat.name}
                  </p>
                  <div className="flex items-center space-x-1 flex-shrink-0">
                    {!isNarrow && (
                      <span className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                        {chat.time}
                      </span>
                    )}
                    {/* Кнопка меню с тремя точками */}
                    <ChatSidebarMenu
                      chatId={chat.id}
                      chatName={chat.name}
                      isGroup={chat.isGroup || false}
                      isCreator={chat.isCreator || false}
                      onLeaveChat={onLeaveChat ? () => onLeaveChat(chat.id) : undefined}
                      onDeleteChat={onDeleteChat ? () => onDeleteChat(chat.id) : undefined}
                      onDeleteGroupChat={onDeleteGroupChat ? () => onDeleteGroupChat(chat.id) : undefined}
                    />
                  </div>
                </div>
                {!isNarrow && (
                  <div className="flex items-center justify-between mt-1">
                    <p className="text-sm text-gray-500 dark:text-gray-400 truncate flex-1 pr-2">
                      {chat.lastMessage}
                    </p>
                    {chat.unread > 0 && (
                      <Badge className="bg-blue-500 text-white text-xs flex-shrink-0">
                        {chat.unread}
                      </Badge>
                    )}
                  </div>
                )}
              </div>
            )}
            {isVeryNarrow && chat.unread > 0 && (
              <Badge className="bg-blue-500 text-white text-xs flex-shrink-0 absolute top-2 right-2">
                {chat.unread}
              </Badge>
            )}
          </div>
        </div>
      );
    })
  )}
    </div>
  );
};
