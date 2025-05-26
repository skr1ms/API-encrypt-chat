
import React from 'react';

interface Message {
  id: number;
  text: string;
  isOwn: boolean;
  time: string;
  user?: {
    name: string;
  };
}

interface MessageListProps {
  messages: Message[];
}

export const MessageList = ({ messages }: MessageListProps) => {
  const formatMessageText = (text: string) => {
    const mentionRegex = /@(\w+)/g;
    const parts = text.split(mentionRegex);
    
    return parts.map((part, index) => {
      if (index % 2 === 1) {
        return (
          <span key={index} className="font-bold text-blue-500 dark:text-blue-400">
            @{part}
          </span>
        );
      }
      return part;
    });
  };

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-50 dark:bg-gray-900">
      {messages.map((msg) => (
        <div
          key={msg.id}
          className={`flex ${msg.isOwn ? 'justify-end' : 'justify-start'}`}
        >
          <div
            className={`max-w-xs lg:max-w-md px-4 py-2 rounded-2xl ${
              msg.isOwn
                ? 'bg-blue-500 text-white'
                : 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white'
            }`}
          >
            <p className="text-sm">{formatMessageText(msg.text)}</p>
            <p className={`text-xs mt-1 ${
              msg.isOwn ? 'text-blue-100' : 'text-gray-500 dark:text-gray-400'
            }`}>
              {msg.time}
            </p>
          </div>
        </div>
      ))}
    </div>
  );
};
