
import * as React from 'react';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Badge } from '@/shared/ui/badge';
import { Users } from 'lucide-react';

interface Chat {
  id: number;
  name: string;
  lastMessage: string;
  time: string;
  unread: number;
  online: boolean;
  isGroup?: boolean;
}

interface ChatListProps {
  chats: Chat[];
  selectedChat: number | null;
  onSelectChat: (id: number) => void;
}

export const ChatList = ({ chats, selectedChat, onSelectChat }: ChatListProps) => {
  return (
    <div className="flex-1 overflow-y-auto">
      {chats.map((chat) => (
        <div
          key={chat.id}
          onClick={() => onSelectChat(chat.id)}
          className={`p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700 border-b border-gray-100 dark:border-gray-700 ${
            selectedChat === chat.id ? 'bg-blue-50 dark:bg-blue-900/20' : ''
          }`}
        >
          <div className="flex items-center space-x-3">
            <div className="relative">
              <Avatar className="h-10 w-10">
                <AvatarFallback className="bg-blue-500 text-white">
                  {chat.isGroup ? <Users className="h-5 w-5" /> : chat.name.charAt(0)}
                </AvatarFallback>
              </Avatar>
              {chat.online && !chat.isGroup && (
                <div className="absolute bottom-0 right-0 h-3 w-3 bg-green-500 border-2 border-white dark:border-gray-800 rounded-full"></div>
              )}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                  {chat.name}
                </p>
                <span className="text-xs text-gray-500 dark:text-gray-400">
                  {chat.time}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <p className="text-sm text-gray-500 dark:text-gray-400 truncate">
                  {chat.lastMessage}
                </p>
                {chat.unread > 0 && (
                  <Badge className="bg-blue-500 text-white text-xs">
                    {chat.unread}
                  </Badge>
                )}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};
