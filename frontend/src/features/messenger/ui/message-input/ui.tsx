
import React from 'react';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Send } from 'lucide-react';

interface MessageInputProps {
  message: string;
  onMessageChange: (message: string) => void;
  onSendMessage: () => void;
}

export const MessageInput = ({ message, onMessageChange, onSendMessage }: MessageInputProps) => {
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      onSendMessage();
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 p-4">
      <div className="flex items-center space-x-2">
        <Input
          value={message}
          onChange={(e) => onMessageChange(e.target.value)}
          placeholder="Написать сообщение..."
          className="flex-1"
          onKeyPress={handleKeyPress}
        />
        <Button onClick={onSendMessage} size="sm" className="bg-blue-500 hover:bg-blue-600">
          <Send className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
};
