import React from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/shared/ui/dialog';
import { Button } from '@/shared/ui/button';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Badge } from '@/shared/ui/badge';
import { UserPlus, Users } from 'lucide-react';

interface User {
  id: number;
  name: string;
  online: boolean;
}

interface GroupSettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  groupName: string;
  users: User[];
}

export const GroupSettingsModal: React.FC<GroupSettingsModalProps> = ({
  isOpen,
  onClose,
  groupName,
  users
}) => {
  const onlineUsers = users.filter(user => user.online);
  const offlineUsers = users.filter(user => !user.online);

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <Users className="h-5 w-5" />
            <span>{groupName}</span>
          </DialogTitle>
        </DialogHeader>
        
        <div className="space-y-4">
          <Button className="w-full" variant="outline">
            <UserPlus className="mr-2 h-4 w-4" />
            Добавить пользователя
          </Button>
          
          {onlineUsers.length > 0 && (
            <div>
              <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-2">
                В сети ({onlineUsers.length})
              </h3>
              <div className="space-y-2">
                {onlineUsers.map((user) => (
                  <div key={user.id} className="flex items-center space-x-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700">
                    <div className="relative">
                      <Avatar className="h-8 w-8">
                        <AvatarFallback className="bg-blue-500 text-white text-sm">
                          {user.name.charAt(0)}
                        </AvatarFallback>
                      </Avatar>
                      <div className="absolute bottom-0 right-0 h-2.5 w-2.5 bg-green-500 border border-white dark:border-gray-800 rounded-full"></div>
                    </div>
                    <div className="flex-1">
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {user.name}
                      </p>
                    </div>
                    <Badge variant="secondary" className="bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300">
                      в сети
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          )}

          {offlineUsers.length > 0 && (
            <div>
              <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-2">
                Не в сети ({offlineUsers.length})
              </h3>
              <div className="space-y-2">
                {offlineUsers.map((user) => (
                  <div key={user.id} className="flex items-center space-x-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700">
                    <Avatar className="h-8 w-8">
                      <AvatarFallback className="bg-gray-500 text-white text-sm">
                        {user.name.charAt(0)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1">
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {user.name}
                      </p>
                    </div>
                    <Badge variant="secondary" className="bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400">
                      не в сети
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};
