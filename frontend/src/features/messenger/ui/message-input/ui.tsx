
import React, { useState } from 'react';
import { Button } from '@/shared/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Send, Smile } from 'lucide-react';

interface MessageInputProps {
  message: string;
  onMessageChange: (message: string) => void;
  onSendMessage: () => void;
  disabled?: boolean;
}

export const MessageInput = ({ message, onMessageChange, onSendMessage, disabled = false }: MessageInputProps) => {
  const [isTextarea, setIsTextarea] = useState(false);

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (message.trim() && !disabled) {
        onSendMessage();
      }
    } else if (e.key === 'Enter' && e.shiftKey) {
      // Shift+Enter для новой строки - переключаемся на textarea
      if (!isTextarea) {
        setIsTextarea(true);
      }
    }
  };

  const handleSend = () => {
    if (message.trim() && !disabled) {
      onSendMessage();
    }
  };

  // Автоматически переключаемся на textarea, если есть переносы строк
  React.useEffect(() => {
    if (message.includes('\n') && !isTextarea) {
      setIsTextarea(true);
    } else if (!message.includes('\n') && isTextarea) {
      setIsTextarea(false);
    }
  }, [message, isTextarea]);

  return (
    <div className="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 p-4">
      <div className="flex items-end space-x-2">
        {isTextarea ? (
          <Textarea
            value={message}
            onChange={(e) => onMessageChange(e.target.value)}
            placeholder="Написать сообщение... (Enter для отправки, Shift+Enter для новой строки)"
            className="flex-1 min-h-[42px] max-h-32 resize-none"
            onKeyDown={handleKeyPress}
            disabled={disabled}
            rows={Math.min(Math.max(message.split('\n').length, 1), 4)}
          />
        ) : (
          <div className="flex-1 relative">
            <input
              type="text"
              value={message}
              onChange={(e) => onMessageChange(e.target.value)}
              placeholder="Написать сообщение..."
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              onKeyDown={handleKeyPress}
              disabled={disabled}
            />
          </div>
        )}
        
        <Button 
          onClick={handleSend} 
          size="sm" 
          className="bg-blue-500 hover:bg-blue-600 h-[42px] px-3"
          disabled={!message.trim() || disabled}
        >
          <Send className="h-4 w-4" />
        </Button>
      </div>
      
      {!disabled && (
        <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
          {isTextarea ? (
            <span>Enter для отправки, Shift+Enter для новой строки</span>
          ) : (
            <span>Enter для отправки, Shift+Enter для многострочного ввода</span>
          )}
        </div>
      )}
    </div>
  );
};
