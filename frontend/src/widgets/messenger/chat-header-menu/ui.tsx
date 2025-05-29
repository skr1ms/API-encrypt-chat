import React, { useState } from 'react';
import { Button } from '@/shared/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { MoreVertical, LogOut, Trash2, AlertTriangle } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

interface ChatHeaderMenuProps {
  chatId: number;
  chatName: string;
  isGroup: boolean;
  isCreator?: boolean;
  onLeaveChat?: () => Promise<void>;
  onDeleteChat?: () => Promise<void>;
  onDeleteGroupChat?: () => Promise<void>;
}

export const ChatHeaderMenu: React.FC<ChatHeaderMenuProps> = ({
  chatId,
  chatName,
  isGroup,
  isCreator = false,
  onLeaveChat,
  onDeleteChat,
  onDeleteGroupChat,
}) => {
  const [showLeaveDialog, setShowLeaveDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showDeleteGroupDialog, setShowDeleteGroupDialog] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const { toast } = useToast();

  const handleLeaveChat = async () => {
    if (!onLeaveChat) return;
    
    setIsLoading(true);
    try {
      await onLeaveChat();
      setShowLeaveDialog(false);
      toast({
        title: "Вы покинули группу",
        description: `Группа "${chatName}" удалена из ваших чатов`,
      });
    } catch (error: any) {
      console.error('Ошибка выхода из группы:', error);
      toast({
        title: "Ошибка",
        description: "Не удалось покинуть группу",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteChat = async () => {
    if (!onDeleteChat) return;
    
    setIsLoading(true);
    try {
      await onDeleteChat();
      setShowDeleteDialog(false);
      toast({
        title: "Чат удален",
        description: `Чат "${chatName}" удален из ваших сообщений`,
      });
    } catch (error: any) {
      console.error('Ошибка удаления чата:', error);
      toast({
        title: "Ошибка",
        description: "Не удалось удалить чат",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteGroupChat = async () => {
    if (!onDeleteGroupChat) return;
    
    setIsLoading(true);
    try {
      await onDeleteGroupChat();
      setShowDeleteGroupDialog(false);
      toast({
        title: "Группа удалена",
        description: `Группа "${chatName}" была полностью удалена`,
      });
    } catch (error: any) {
      console.error('Ошибка удаления группы:', error);
      toast({
        title: "Ошибка",
        description: "Не удалось удалить группу",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="h-8 w-8 p-0 hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            <MoreVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>        <DropdownMenuContent align="end" className="w-48">
          {isGroup ? (
            isCreator ? (
              <DropdownMenuItem
                className="text-red-600 hover:text-red-700 focus:text-red-700"
                onClick={() => setShowDeleteGroupDialog(true)}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Удалить группу
              </DropdownMenuItem>
            ) : (
              <DropdownMenuItem
                className="text-red-600 hover:text-red-700 focus:text-red-700"
                onClick={() => setShowLeaveDialog(true)}
              >
                <LogOut className="mr-2 h-4 w-4" />
                Покинуть группу
              </DropdownMenuItem>
            )
          ) : (
            <DropdownMenuItem
              className="text-red-600 hover:text-red-700 focus:text-red-700"
              onClick={() => setShowDeleteDialog(true)}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Удалить чат
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Диалог подтверждения выхода из группы */}
      <AlertDialog open={showLeaveDialog} onOpenChange={setShowLeaveDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-orange-500" />
              <span>Покинуть группу</span>
            </AlertDialogTitle>
            <AlertDialogDescription>
              Вы действительно хотите покинуть группу <strong>"{chatName}"</strong>?
              <br />
              <br />
              После выхода вы не сможете видеть новые сообщения и участвовать в обсуждении.
              Вас можно будет добавить обратно только администратор или создатель группы.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isLoading}>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                handleLeaveChat();
              }}
              disabled={isLoading}
              className="bg-orange-500 hover:bg-orange-600 text-white"
            >
              {isLoading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Выход...
                </>
              ) : (
                'Покинуть группу'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Диалог подтверждения удаления чата */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-500" />
              <span>Удалить чат</span>
            </AlertDialogTitle>
            <AlertDialogDescription>
              Вы действительно хотите удалить чат с <strong>"{chatName}"</strong>?
              <br />
              <br />
              Чат будет удален только у вас. Собеседник сможет продолжать видеть историю сообщений.
              Если собеседник отправит новое сообщение, чат снова появится в вашем списке.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isLoading}>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                handleDeleteChat();
              }}
              disabled={isLoading}
              className="bg-red-500 hover:bg-red-600 text-white"
            >
              {isLoading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Удаление...
                </>
              ) : (
                'Удалить чат'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>      </AlertDialog>

      {/* Диалог подтверждения удаления группы */}
      <AlertDialog open={showDeleteGroupDialog} onOpenChange={setShowDeleteGroupDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-500" />
              <span>Удалить группу</span>
            </AlertDialogTitle>
            <AlertDialogDescription>
              Вы действительно хотите полностью удалить группу <strong>"{chatName}"</strong>?
              <br />
              <br />
              <span className="text-red-600 font-medium">
                Внимание! Это действие необратимо.
              </span>
              <br />
              Группа будет удалена для всех участников, и вся история сообщений будет потеряна навсегда.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isLoading}>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                handleDeleteGroupChat();
              }}
              disabled={isLoading}
              className="bg-red-500 hover:bg-red-600 text-white"
            >
              {isLoading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Удаление...
                </>
              ) : (
                'Удалить группу навсегда'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
};
