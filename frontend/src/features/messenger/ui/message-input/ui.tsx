
import React, { useState, useRef, useEffect } from 'react';
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
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (message.trim() && !disabled) {
        onSendMessage();
      }
    } else if (e.key === 'Enter' && e.shiftKey) {
      // Shift+Enter для новой строки
      e.preventDefault();
      
      if (!isTextarea) {
        // Переключаемся на textarea и добавляем перенос строки
        setIsTextarea(true);
        const cursorPosition = (e.target as HTMLInputElement).selectionStart || message.length;
        const newMessage = message.slice(0, cursorPosition) + '\n' + message.slice(cursorPosition);
        onMessageChange(newMessage);
        
        // Устанавливаем курсор на следующую строку после переключения
        setTimeout(() => {
          if (textareaRef.current) {
            textareaRef.current.focus();
            textareaRef.current.setSelectionRange(cursorPosition + 1, cursorPosition + 1);
          }
        }, 0);
      } else {
        // Уже в textarea - просто добавляем перенос строки
        const target = e.target as HTMLTextAreaElement;
        const cursorPosition = target.selectionStart || message.length;
        const newMessage = message.slice(0, cursorPosition) + '\n' + message.slice(cursorPosition);
        onMessageChange(newMessage);
        
        // Устанавливаем курсор после новой строки
        setTimeout(() => {
          if (textareaRef.current) {
            textareaRef.current.setSelectionRange(cursorPosition + 1, cursorPosition + 1);
          }
        }, 0);
      }
    }
  };

  const handleSend = () => {
    if (message.trim() && !disabled) {
      onSendMessage();
    }
  };

  // Автоматически переключаемся на textarea, если есть переносы строк
  useEffect(() => {
    if (message.includes('\n') && !isTextarea) {
      setIsTextarea(true);
    } else if (!message.includes('\n') && isTextarea && message.trim() === '') {
      // Переключаемся обратно на input только если сообщение пустое
      setIsTextarea(false);
    }
  }, [message, isTextarea]);

  // Автоматическое изменение высоты textarea
  useEffect(() => {
    if (isTextarea && textareaRef.current) {
      const textarea = textareaRef.current;
      // Сбрасываем высоту для правильного вычисления
      textarea.style.height = 'auto';
      // Устанавливаем высоту по содержимому с ограничениями
      const scrollHeight = textarea.scrollHeight;
      const minHeight = 42; // минимальная высота (как у input)
      const maxHeight = 150; // максимальная высота (примерно 6 строк)
      const newHeight = Math.min(Math.max(scrollHeight, minHeight), maxHeight);
      textarea.style.height = `${newHeight}px`;
    }
  }, [message, isTextarea]);

  // Функция для автофокуса на правильный элемент
  useEffect(() => {
    if (isTextarea && textareaRef.current) {
      textareaRef.current.focus();
    } else if (!isTextarea && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isTextarea]);

  return (
    <div className="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 p-4">
      <div className="flex items-end space-x-2">
        {isTextarea ? (
          <Textarea
            ref={textareaRef}
            value={message}
            onChange={(e) => onMessageChange(e.target.value)}
            placeholder="Написать сообщение... (Enter для отправки, Shift+Enter для новой строки)"
            className="flex-1 min-h-[42px] resize-none overflow-y-auto transition-all duration-200 ease-in-out"
            onKeyDown={handleKeyPress}
            disabled={disabled}
            style={{ height: '42px' }} // начальная высота
          />
        ) : (
          <div className="flex-1 relative">
            <input
              ref={inputRef}
              type="text"
              value={message}
              onChange={(e) => onMessageChange(e.target.value)}
              placeholder="Написать сообщение..."
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent h-[42px] transition-all duration-200 ease-in-out"
              onKeyDown={handleKeyPress}
              disabled={disabled}
            />
          </div>
        )}
        
        <Button 
          onClick={handleSend} 
          size="sm" 
          className="bg-blue-500 hover:bg-blue-600 h-[42px] px-3 flex-shrink-0"
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
