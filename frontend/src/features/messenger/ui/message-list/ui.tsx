
import React, { useEffect, useRef } from 'react';

interface Message {
  id: number | string;
  text: string;
  isOwn: boolean;
  time: string;
  isSystem?: boolean; // флаг для системных сообщений
  type?: 'user_left' | 'group_deleted' | 'user_joined' | 'system'; // тип системного сообщения
  user?: {
    name: string;
  };
}

interface MessageListProps {
  messages: Message[];
}

export const MessageList = ({ messages }: MessageListProps) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const formatMessageText = (text: string) => {
    // Сначала разделяем текст на строки для обработки переносов
    const lines = text.split('\n');
    
    return lines.map((line, lineIndex) => {
      // Обрабатываем упоминания в каждой строке
      const mentionRegex = /@(\w+)/g;
      const parts = line.split(mentionRegex);
      
      const formattedLine = parts.map((part, partIndex) => {
        if (partIndex % 2 === 1) {
          return (
            <span key={`${lineIndex}-${partIndex}`} className="font-bold text-blue-500 dark:text-blue-400">
              @{part}
            </span>
          );
        }
        return part;
      });
      
      // Если это не последняя строка, добавляем <br />
      return (
        <React.Fragment key={lineIndex}>
          {formattedLine}
          {lineIndex < lines.length - 1 && <br />}
        </React.Fragment>
      );
    });
  };

  if (messages.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="text-center">
          <div className="text-gray-400 mb-2">
            <svg className="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <p className="text-gray-500 dark:text-gray-400">Пока нет сообщений</p>
          <p className="text-sm text-gray-400 dark:text-gray-500">Начните общение, отправив первое сообщение</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-50 dark:bg-gray-900 telegram-scrollbar">
      {messages.map((msg) => (
        <div
          key={msg.id}
          className={`flex ${msg.isSystem ? 'justify-center' : msg.isOwn ? 'justify-end' : 'justify-start'}`}
        >
          {msg.isSystem ? (
            <div className={`max-w-xs lg:max-w-md px-3 py-1 rounded-lg text-center text-sm ${
              msg.type === 'group_deleted' 
                ? 'bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400' 
                : msg.type === 'user_left'
                  ? 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-600 dark:text-yellow-400'
                  : msg.type === 'user_joined'
                    ? 'bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-400'
                    : 'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400'
            }`}>
              <span>{formatMessageText(msg.text)}</span>
              <div className="text-xs text-gray-400 dark:text-gray-500 mt-1">{msg.time}</div>
            </div>
          ) : (
            <div
              className={`max-w-xs lg:max-w-md px-4 py-2 rounded-2xl ${
                msg.isOwn
                  ? 'bg-blue-500 text-white rounded-br-sm'
                : 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-700 rounded-bl-sm'
            }`}
          >
            {!msg.isOwn && msg.user?.name && (
              <p className="text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
                {msg.user.name}
              </p>
            )}
            <p className="text-sm">{formatMessageText(msg.text)}</p>
            <p className={`text-xs mt-1 ${
              msg.isOwn ? 'text-blue-100' : 'text-gray-500 dark:text-gray-400'
            }`}>
              {msg.time}
            </p>
          </div>
          )}
        </div>
      ))}
      <div ref={messagesEndRef} />
    </div>
  );
};
