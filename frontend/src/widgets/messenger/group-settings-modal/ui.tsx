import React, { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/shared/ui/dialog';
import { Button } from '@/shared/ui/button';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Badge } from '@/shared/ui/badge';
import { UserPlus, Users, X, UserMinus, AlertTriangle } from 'lucide-react';
import { UserSearch } from '@/shared/ui/user-search';
import { User as ApiUser } from '@/shared/api/userApi';
import { useToast } from '@/hooks/use-toast';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface User {
  id: number;
  name: string;
  online: boolean;
  role?: string; // 'creator', 'admin', 'member'
}

interface GroupSettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  groupName: string;
  users: User[];
  chatId?: number;
  currentUserId?: number;
  onAddMember?: (userId: number) => Promise<void>;
  onRemoveMember?: (userId: number) => Promise<void>;
  onSetAdmin?: (userId: number) => Promise<void>;
  onRemoveAdmin?: (userId: number) => Promise<void>;
}

export const GroupSettingsModal: React.FC<GroupSettingsModalProps> = ({
  isOpen,
  onClose,
  groupName,
  users,
  chatId,
  currentUserId,
  onAddMember,
  onRemoveMember,
  onSetAdmin,
  onRemoveAdmin
}) => {
  const [showAddUser, setShowAddUser] = useState(false);
  const [isActionInProgress, setIsActionInProgress] = useState(false);
  const [confirmDialogOpen, setConfirmDialogOpen] = useState(false);
  const [userToRemove, setUserToRemove] = useState<{ id: number, name: string } | null>(null);
  const [confirmAdminDialogOpen, setConfirmAdminDialogOpen] = useState(false);
  const [userForAdminAction, setUserForAdminAction] = useState<{ id: number, name: string, action: 'set' | 'remove' } | null>(null);
  const { toast } = useToast();
  
  // Проверка является ли текущий пользователь создателем группы
  const isCreator = users.some(user => user.id === currentUserId && user.role === 'creator');
  
  // Проверка является ли текущий пользователь администратором
  const isAdmin = users.some(user => user.id === currentUserId && (user.role === 'admin' || user.role === 'creator'));

  const onlineUsers = users.filter(user => user.online);
  const offlineUsers = users.filter(user => !user.online);

  const handleAddMember = async (user: ApiUser) => {
    if (!onAddMember || !chatId) return;
    
    setIsActionInProgress(true);
    try {
      await onAddMember(user.id);
      toast({
        title: "Пользователь добавлен",
        description: `${user.username} добавлен в группу`,
      });
      setShowAddUser(false);
    } catch (error) {
      console.error('Ошибка добавления пользователя:', error);
      toast({
        title: "Ошибка",
        description: "Не удалось добавить пользователя в группу",
        variant: "destructive",
      });
    } finally {
      setIsActionInProgress(false);
    }
  };

  const openRemoveConfirmDialog = (userId: number, userName: string) => {
    setUserToRemove({ id: userId, name: userName });
    setConfirmDialogOpen(true);
  };

  const handleRemoveMember = async () => {
    if (!onRemoveMember || !chatId || !userToRemove) return;
    
    setIsActionInProgress(true);
    try {
      await onRemoveMember(userToRemove.id);
      toast({
        title: "Пользователь удален",
        description: `${userToRemove.name} удален из группы`,
      });
      setConfirmDialogOpen(false);
      setUserToRemove(null);
    } catch (error: any) {
      console.error('Ошибка удаления пользователя:', error);
      
      // Определяем сообщение об ошибке в зависимости от ответа сервера
      let errorMessage = "Не удалось удалить пользователя из группы";
      if (error?.response?.status === 403) {
        errorMessage = "У вас недостаточно прав для удаления этого пользователя";
      } else if (error.message) {
        errorMessage = error.message;
      }
      
      toast({
        title: "Ошибка",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setIsActionInProgress(false);
    }
  };
  
  // Открыть диалог подтверждения для назначения/снятия админа
  const openAdminActionDialog = (userId: number, userName: string, action: 'set' | 'remove') => {
    setUserForAdminAction({ id: userId, name: userName, action });
    setConfirmAdminDialogOpen(true);
  };
  
  // Обработчик для выполнения действия с администратором
  const handleAdminAction = async () => {
    if (!userForAdminAction || !chatId) return;
    if (!onSetAdmin && userForAdminAction.action === 'set') return;
    if (!onRemoveAdmin && userForAdminAction.action === 'remove') return;
    
    setIsActionInProgress(true);
    try {
      if (userForAdminAction.action === 'set' && onSetAdmin) {
        await onSetAdmin(userForAdminAction.id);
        toast({
          title: "Права администратора выданы",
          description: `${userForAdminAction.name} теперь администратор`,
        });
      } else if (userForAdminAction.action === 'remove' && onRemoveAdmin) {
        await onRemoveAdmin(userForAdminAction.id);
        toast({
          title: "Права администратора сняты",
          description: `${userForAdminAction.name} больше не администратор`,
        });
      }
      
      setConfirmAdminDialogOpen(false);
      setUserForAdminAction(null);
    } catch (error: any) {
      console.error('Ошибка управления правами администратора:', error);
      
      // Определяем сообщение об ошибке в зависимости от ответа сервера
      let errorMessage = "Не удалось изменить права доступа пользователя";
      if (error?.response?.status === 403) {
        errorMessage = "У вас недостаточно прав для этого действия";
      } else if (error.message) {
        errorMessage = error.message;
      }
      
      toast({
        title: "Ошибка",
        description: errorMessage,
        variant: "destructive",
      });
    } finally {
      setIsActionInProgress(false);
    }
  };

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center space-x-2">
              <Users className="h-5 w-5" />
              <span>{groupName}</span>
            </DialogTitle>
          </DialogHeader>
          
          <div className="space-y-4">
            {showAddUser ? (
              <div>
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                    Добавить участника
                  </h3>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-8 w-8 p-0"
                    onClick={() => setShowAddUser(false)}
                    disabled={isActionInProgress}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
                <UserSearch
                  onUserSelect={handleAddMember}
                  placeholder="Поиск пользователей для добавления..."
                  showEmail={false}
                />
              </div>
            ) : (
              <Button
                className="w-full"
                variant="outline"
                onClick={() => setShowAddUser(true)}
                disabled={!onAddMember || isActionInProgress}
              >
                <UserPlus className="mr-2 h-4 w-4" />
                Добавить пользователя
              </Button>
            )}
            
            {onlineUsers.length > 0 && (
              <div>
                <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-2">
                  В сети ({onlineUsers.length})
                </h3>
                <div className="space-y-2 telegram-scrollbar max-h-40 overflow-y-auto">
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
                      <div className="flex items-center space-x-2">
                        {/* Отображение ролей */}
                        {user.role === 'creator' && (
                          <Badge variant="secondary" className="bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300">
                            создатель
                          </Badge>
                        )}
                        {user.role === 'admin' && (
                          <Badge variant="secondary" className="bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
                            админ
                          </Badge>
                        )}
                        <Badge variant="secondary" className="bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300">
                          в сети
                        </Badge>

                        {/* Кнопки управления пользователями */}
                        <div className="flex space-x-1">
                          {/* Кнопка для назначения администратора (только для создателя) */}
                          {isCreator && user.role !== 'creator' && user.role !== 'admin' && onSetAdmin && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 p-1 text-blue-500"
                              title="Назначить администратором"
                              onClick={() => openAdminActionDialog(user.id, user.name, 'set')}
                              disabled={isActionInProgress}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                                <circle cx="9" cy="7" r="4" />
                                <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
                                <path d="M19 4v3" />
                                <path d="M22 7h-6" />
                              </svg>
                            </Button>
                          )}
                          
                          {/* Кнопка для снятия администратора (только для создателя) */}
                          {isCreator && user.role === 'admin' && onRemoveAdmin && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 p-1 text-orange-500"
                              title="Снять права администратора"
                              onClick={() => openAdminActionDialog(user.id, user.name, 'remove')}
                              disabled={isActionInProgress}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                                <circle cx="9" cy="7" r="4" />
                                <line x1="17" y1="8" x2="22" y2="13" />
                                <line x1="22" y1="8" x2="17" y2="13" />
                              </svg>
                            </Button>
                          )}
                          
                          {/* Кнопка для удаления пользователя (для админов и создателя, но не самого себя и не создателя) */}
                          {onRemoveMember && isAdmin && user.id !== currentUserId && user.role !== 'creator' && (
                            user.role === 'admin' ? (
                              // Только создатель может удалять администраторов
                              isCreator && (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="h-8 p-1 text-red-500"
                                  title="Удалить пользователя"
                                  onClick={() => openRemoveConfirmDialog(user.id, user.name)}
                                  disabled={isActionInProgress}
                                >
                                  <UserMinus className="h-4 w-4" />
                                </Button>
                              )
                            ) : (
                              // Обычных пользователей могут удалять и создатель, и админы
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-8 p-1 text-red-500"
                                title="Удалить пользователя"
                                onClick={() => openRemoveConfirmDialog(user.id, user.name)}
                                disabled={isActionInProgress}
                              >
                                <UserMinus className="h-4 w-4" />
                              </Button>
                            )
                          )}
                        </div>
                      </div>
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
                <div className="space-y-2 telegram-scrollbar max-h-40 overflow-y-auto">
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
                      <div className="flex items-center space-x-2">
                        {/* Отображение ролей */}
                        {user.role === 'creator' && (
                          <Badge variant="secondary" className="bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300">
                            создатель
                          </Badge>
                        )}
                        {user.role === 'admin' && (
                          <Badge variant="secondary" className="bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
                            админ
                          </Badge>
                        )}
                        <Badge variant="secondary" className="bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400">
                          не в сети
                        </Badge>

                        {/* Кнопки управления пользователями */}
                        <div className="flex space-x-1">
                          {/* Кнопка для назначения администратора (только для создателя) */}
                          {isCreator && user.role !== 'creator' && user.role !== 'admin' && onSetAdmin && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 p-1 text-blue-500"
                              title="Назначить администратором"
                              onClick={() => openAdminActionDialog(user.id, user.name, 'set')}
                              disabled={isActionInProgress}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                                <circle cx="9" cy="7" r="4" />
                                <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
                                <path d="M19 4v3" />
                                <path d="M22 7h-6" />
                              </svg>
                            </Button>
                          )}
                          
                          {/* Кнопка для снятия администратора (только для создателя) */}
                          {isCreator && user.role === 'admin' && onRemoveAdmin && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 p-1 text-orange-500"
                              title="Снять права администратора"
                              onClick={() => openAdminActionDialog(user.id, user.name, 'remove')}
                              disabled={isActionInProgress}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                                <circle cx="9" cy="7" r="4" />
                                <line x1="17" y1="8" x2="22" y2="13" />
                                <line x1="22" y1="8" x2="17" y2="13" />
                              </svg>
                            </Button>
                          )}
                          
                          {/* Кнопка для удаления пользователя (для админов и создателя, но не самого себя и не создателя) */}
                          {onRemoveMember && isAdmin && user.id !== currentUserId && user.role !== 'creator' && (
                            user.role === 'admin' ? (
                              // Только создатель может удалять администраторов
                              isCreator && (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="h-8 p-1 text-red-500"
                                  title="Удалить пользователя"
                                  onClick={() => openRemoveConfirmDialog(user.id, user.name)}
                                  disabled={isActionInProgress}
                                >
                                  <UserMinus className="h-4 w-4" />
                                </Button>
                              )
                            ) : (
                              // Обычных пользователей могут удалять и создатель, и админы
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-8 p-1 text-red-500"
                                title="Удалить пользователя"
                                onClick={() => openRemoveConfirmDialog(user.id, user.name)}
                                disabled={isActionInProgress}
                              >
                                <UserMinus className="h-4 w-4" />
                              </Button>
                            )
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
      
      <AlertDialog open={confirmDialogOpen} onOpenChange={setConfirmDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-500" />
              <span>Удаление участника группы</span>
            </AlertDialogTitle>
            <AlertDialogDescription>
              {userToRemove ? (
                <>
                  Вы действительно хотите удалить пользователя <strong>{userToRemove.name}</strong> из группы?
                  <br />
                  <br />
                  Пользователь не будет иметь доступа к сообщениям группы после удаления.
                </>
              ) : (
                'Вы действительно хотите удалить этого пользователя из группы?'
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isActionInProgress}>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                handleRemoveMember();
              }}
              disabled={isActionInProgress}
              className="bg-red-500 hover:bg-red-600 text-white"
            >
              {isActionInProgress ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Удаление...
                </>
              ) : (
                'Удалить'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      
      {/* Диалог подтверждения для назначения/снятия прав администратора */}
      <AlertDialog open={confirmAdminDialogOpen} onOpenChange={setConfirmAdminDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              {userForAdminAction?.action === 'set' ? (
                <>
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-blue-500">
                    <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                    <circle cx="9" cy="7" r="4" />
                    <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
                    <path d="M19 4v3" />
                    <path d="M22 7h-6" />
                  </svg>
                  <span>Назначение администратора</span>
                </>
              ) : (
                <>
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-orange-500">
                    <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
                    <circle cx="9" cy="7" r="4" />
                    <line x1="17" y1="8" x2="22" y2="13" />
                    <line x1="22" y1="8" x2="17" y2="13" />
                  </svg>
                  <span>Снятие прав администратора</span>
                </>
              )}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {userForAdminAction?.action === 'set' ? (
                <>
                  Вы действительно хотите назначить пользователя <strong>{userForAdminAction?.name}</strong> администратором?
                  <br />
                  <br />
                  Администраторы могут удалять пользователей из группы и имеют больше прав.
                </>
              ) : (
                <>
                  Вы действительно хотите снять права администратора у пользователя <strong>{userForAdminAction?.name}</strong>?
                  <br />
                  <br />
                  После этого пользователь не сможет удалять других пользователей и управлять группой.
                </>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isActionInProgress}>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                handleAdminAction();
              }}
              disabled={isActionInProgress}
              className={userForAdminAction?.action === 'set' 
                ? "bg-blue-500 hover:bg-blue-600 text-white" 
                : "bg-orange-500 hover:bg-orange-600 text-white"}
            >
              {isActionInProgress ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  {userForAdminAction?.action === 'set' ? 'Назначение...' : 'Снятие...'}
                </>
              ) : (
                userForAdminAction?.action === 'set' ? 'Назначить' : 'Снять права'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
};
