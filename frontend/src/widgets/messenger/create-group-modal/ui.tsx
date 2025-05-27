import React, { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/shared/ui/dialog';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import { UserSearch } from '@/shared/ui/user-search';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Badge } from '@/shared/ui/badge';
import { X, Users } from 'lucide-react';
import { User } from '@/shared/api/userApi';
import { useToast } from '@/hooks/use-toast';

interface CreateGroupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreateGroup: (groupName: string, participants: number[]) => Promise<void>;
}

export const CreateGroupModal = ({ isOpen, onClose, onCreateGroup }: CreateGroupModalProps) => {
  const [groupName, setGroupName] = useState('');
  const [selectedUsers, setSelectedUsers] = useState<User[]>([]);
  const [isCreating, setIsCreating] = useState(false);
  const { toast } = useToast();

  const handleUserSelect = (user: User) => {
    // Проверяем, не выбран ли уже этот пользователь
    if (selectedUsers.find(u => u.id === user.id)) {
      toast({
        title: "Пользователь уже добавлен",
        description: `${user.username} уже добавлен в группу`,
        variant: "destructive",
      });
      return;
    }

    setSelectedUsers(prev => [...prev, user]);
    toast({
      title: "Пользователь добавлен",
      description: `${user.username} добавлен в группу`,
    });
  };

  const handleRemoveUser = (userId: number) => {
    setSelectedUsers(prev => prev.filter(u => u.id !== userId));
  };

  const handleCreateGroup = async () => {
    if (!groupName.trim()) {
      toast({
        title: "Ошибка",
        description: "Введите название группы",
        variant: "destructive",
      });
      return;
    }

    if (selectedUsers.length === 0) {
      toast({
        title: "Ошибка",
        description: "Добавьте хотя бы одного участника",
        variant: "destructive",
      });
      return;
    }

    setIsCreating(true);
    try {
      const participantIds = selectedUsers.map(user => user.id);
      await onCreateGroup(groupName.trim(), participantIds);
      
      // Сброс состояния после успешного создания
      setGroupName('');
      setSelectedUsers([]);
      onClose();
      
      toast({
        title: "Группа создана",
        description: `Группа "${groupName}" успешно создана`,
      });
    } catch (error) {
      console.error('Ошибка создания группы:', error);
      toast({
        title: "Ошибка создания группы",
        description: "Не удалось создать группу. Попробуйте еще раз",
        variant: "destructive",
      });
    } finally {
      setIsCreating(false);
    }
  };

  const handleClose = () => {
    if (!isCreating) {
      setGroupName('');
      setSelectedUsers([]);
      onClose();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <Users className="h-5 w-5" />
            <span>Создать группу</span>
          </DialogTitle>
        </DialogHeader>
        
        <div className="space-y-4">
          {/* Название группы */}
          <div className="space-y-2">
            <Label htmlFor="groupName">Название группы</Label>
            <Input
              id="groupName"
              placeholder="Введите название группы..."
              value={groupName}
              onChange={(e) => setGroupName(e.target.value)}
              disabled={isCreating}
            />
          </div>

          {/* Выбранные участники */}
          {selectedUsers.length > 0 && (
            <div className="space-y-2">
              <Label>Участники ({selectedUsers.length})</Label>
              <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto telegram-scrollbar">
                {selectedUsers.map((user) => (
                  <Badge
                    key={user.id}
                    variant="secondary"
                    className="flex items-center space-x-1 pr-1"
                  >
                    <Avatar className="h-4 w-4">
                      <AvatarFallback className="bg-blue-100 text-blue-600 text-xs">
                        {user.username.charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <span className="text-xs">{user.username}</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-4 w-4 p-0 hover:bg-red-100"
                      onClick={() => handleRemoveUser(user.id)}
                      disabled={isCreating}
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {/* Поиск пользователей */}
          <div className="space-y-2">
            <Label>Добавить участников</Label>
            <UserSearch
              onUserSelect={handleUserSelect}
              placeholder="Поиск пользователей для добавления в группу..."
              showEmail={false}
            />
          </div>

          {/* Кнопки */}
          <div className="flex justify-end space-x-2 pt-4">
            <Button 
              variant="outline" 
              onClick={handleClose}
              disabled={isCreating}
            >
              Отмена
            </Button>
            <Button 
              onClick={handleCreateGroup}
              disabled={isCreating || !groupName.trim() || selectedUsers.length === 0}
            >
              {isCreating ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Создание...
                </>
              ) : (
                'Создать группу'
              )}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};
