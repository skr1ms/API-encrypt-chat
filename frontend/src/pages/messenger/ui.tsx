import * as React from 'react';
import { useState, useEffect } from 'react';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Button } from '@/shared/ui/button';
import { MessengerSidebar } from '@/widgets/messenger/sidebar';
import { MessageList } from '@/features/messenger/ui/message-list';
import { MessageInput } from '@/features/messenger/ui/message-input';
import { GroupSettingsModal } from '@/widgets/messenger/group-settings-modal';
import { MessageSquare, Users } from 'lucide-react';
import { chatAPI } from '@/shared/api/chatApi';
import { useToast } from '@/hooks/use-toast';
import { websocketService } from '@/shared/lib/websocket/websocketService';

interface Chat {
  id: number;
  name: string;
  lastMessage: string;
  time: string;
  unread: number;
  online: boolean;
  isGroup?: boolean;
}

export const MessengerPage = () => {
  const [selectedChat, setSelectedChat] = useState<number | null>(null);
  const [message, setMessage] = useState('');
  const [showProfile, setShowProfile] = useState(false);
  const [showGroupSettings, setShowGroupSettings] = useState(false);
  const [chats, setChats] = useState<Chat[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();

  const groupUsers: any[] = [];
  const [currentUserId, setCurrentUserId] = useState<number | null>(null);

  const [messages, setMessages] = useState<any[]>([]);
  const [isMessagesLoading, setIsMessagesLoading] = useState(false);

  // Общая функция для форматирования ошибок
  const formatError = (error: any, defaultTitle: string, defaultDescription: string) => {
    let title = defaultTitle;
    let description = defaultDescription;
    
    if (error?.response?.status) {
      const status = error.response.status;
      switch (status) {
        case 401:
          title = "Ошибка авторизации";
          description = "Требуется авторизация";
          break;
        case 403:
          title = "Доступ запрещен";
          description = "У вас нет прав для выполнения этого действия";
          break;
        case 404:
          title = "Не найдено";
          description = "Запрашиваемый ресурс не найден";
          break;
        case 409:
          title = "Конфликт";
          description = "Ресурс уже существует";
          break;
        case 413:
          title = "Данные слишком большие";
          description = "Размер данных превышает допустимый лимит";
          break;
        case 422:
          title = "Некорректные данные";
          description = "Проверьте правильность заполнения полей";
          break;
        case 429:
          title = "Превышен лимит запросов";
          description = "Слишком много запросов. Попробуйте позже";
          break;
        case 500:
          title = "Ошибка сервера";
          description = "Внутренняя ошибка сервера. Попробуйте позже";
          break;
        default:
          description = `Ошибка сервера (код ${status})`;
      }
    } else if (error?.code === 'NETWORK_ERROR' || error?.message?.includes('Network Error')) {
      title = "Ошибка сети";
      description = "Проверьте подключение к интернету";
    } else if (error instanceof TypeError || error?.message?.includes('fetch')) {
      title = "Ошибка подключения";
      description = "Не удается подключиться к серверу";
    } else if (error instanceof Error) {
      description = error.message;
    } else if (typeof error === 'string') {
      description = error;
    }
    
    return { title, description };
  };

  // Загрузка чатов при монтировании компонента
  useEffect(() => {
    loadChats();
    
    // Получение ID текущего пользователя из localStorage
    const userDataString = localStorage.getItem('user');
    if (userDataString) {
      try {
        const userData = JSON.parse(userDataString);
        if (userData && userData.id) {
          const userId = typeof userData.id === 'string' ? parseInt(userData.id, 10) : userData.id;
          setCurrentUserId(userId);
        }
      } catch (error) {
        console.error('Ошибка при получении данных пользователя:', error);
      }
    }

    // Подключение к WebSocket
    const token = localStorage.getItem('token');
    if (token) {
      websocketService.connect(token);
    }

    // Слушатель для новых сообщений
    const handleNewMessage = (event: CustomEvent) => {
      const message = event.detail;
      console.log('Received new message event:', message);
      
      // Обновляем сообщения если текущий чат совпадает
      if (selectedChat && message.chatId === selectedChat.toString()) {
        setMessages(prevMessages => {
          // Проверяем, что сообщение еще не добавлено
          const messageExists = prevMessages.some(msg => msg.id === message.id);
          if (!messageExists) {
            const formattedMessage = {
              id: message.id,
              text: message.content,
              time: new Date(message.timestamp).toLocaleTimeString('ru-RU', { 
                hour: '2-digit', 
                minute: '2-digit' 
              }),
              isOwn: currentUserId !== null && parseInt(message.senderId) === currentUserId,
              user: {
                name: message.senderUsername
              }
            };
            return [...prevMessages, formattedMessage];
          }
          return prevMessages;
        });
      }
    };

    window.addEventListener('newMessage', handleNewMessage as EventListener);

    return () => {
      window.removeEventListener('newMessage', handleNewMessage as EventListener);
      websocketService.disconnect();
    };
  }, [selectedChat, currentUserId]);

  const loadChats = async () => {
    try {
      setIsLoading(true);
      const response = await chatAPI.getChats();
      console.log('Loaded chats:', response);
      
      // Проверка на ответ с ошибкой
      if (response && (response as any).error) {
        throw new Error((response as any).error);
      }
      
      // Преобразуем данные с бэкенда в формат для фронтенда
      const chatsData = Array.isArray(response) ? response : (response as any)?.data || [];
      const formattedChats: Chat[] = chatsData.map((chat: any) => ({
        id: chat.id,
        name: chat.name,
        lastMessage: chat.last_message || 'Нет сообщений',
        time: chat.updated_at ? new Date(chat.updated_at).toLocaleTimeString('ru-RU', { 
          hour: '2-digit', 
          minute: '2-digit' 
        }) : '',
        unread: 0, // TODO: реализовать подсчет непрочитанных
        online: false, // TODO: реализовать статус онлайн для приватных чатов
        isGroup: chat.is_group
      }));
      
      setChats(formattedChats);
      
      // Восстанавливаем последний выбранный чат
      const lastSelectedChatId = localStorage.getItem('lastSelectedChat');
      if (lastSelectedChatId && formattedChats.length > 0) {
        const chatId = parseInt(lastSelectedChatId, 10);
        const chatExists = formattedChats.some(chat => chat.id === chatId);
        if (chatExists) {
          setSelectedChat(chatId);
        } else if (formattedChats.length > 0) {
          // Если последний чат не найден, выбираем первый доступный
          setSelectedChat(formattedChats[0].id);
          localStorage.setItem('lastSelectedChat', formattedChats[0].id.toString());
        }
      }
    } catch (error: any) {
      console.error('Ошибка загрузки чатов:', error);
      
      // Используем общую функцию форматирования ошибок с кастомизацией
      let { title, description } = formatError(error, "Ошибка загрузки чатов", "Не удалось загрузить список чатов");
      
      // Специфичная обработка для ошибок авторизации при загрузке чатов
      if (error?.response?.status === 401) {
        description = "Необходимо авторизоваться для доступа к чатам";
      }
      
      toast({
        title,
        description,
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateChat = async (chatName: string, isGroup: boolean, participants: number[]) => {
    try {
      const response = await chatAPI.createChat(chatName, isGroup, participants);
      console.log('Создан чат:', response);
      
      const chatData = response?.data || response;
      
      // Обновляем список чатов
      await loadChats();
      
      // Автоматически выбираем созданный чат
      if (chatData?.id) {
        setSelectedChat(chatData.id);
        
        // Сохраняем ID последнего выбранного чата
        localStorage.setItem('lastSelectedChat', chatData.id.toString());
        
        toast({
          title: "Чат создан",
          description: `Чат "${chatName}" успешно создан`,
        });
      }
      
      return chatData;
    } catch (error: any) {
      console.error('Ошибка создания чата:', error);
      
      // Специальная обработка для некоторых ошибок создания чата
      let { title, description } = formatError(error, "Ошибка создания чата", "Не удалось создать чат");
      
      // Дополнительные специфичные для создания чата ошибки
      if (error?.response?.status === 409) {
        title = "Чат уже существует";
        description = "Чат с таким именем уже существует";
      } else if (error?.response?.status === 401) {
        description = "Требуется авторизация для создания чата";
      } else if (error?.response?.status === 403) {
        description = "У вас нет прав для создания чатов";
      }
      
      toast({
        title,
        description,
        variant: "destructive",
      });
      
      throw error;
    }
  };
  
  const loadMessages = async (chatId: number) => {
    if (!chatId) return;
    
    try {
      setIsMessagesLoading(true);
      const response = await chatAPI.getMessages(chatId.toString());
      console.log('Loaded messages:', response);
      
      // Проверка на ответ с ошибкой
      if (response && (response as any).error) {
        throw new Error((response as any).error);
      }
      
      const messagesData = Array.isArray(response) ? response : (response as any)?.data || [];
      const formattedMessages = messagesData.map((msg: any) => ({
        id: msg.id,
        text: msg.content || msg.encrypted_content || "Зашифрованное сообщение",
        time: new Date(msg.created_at).toLocaleTimeString('ru-RU', { 
          hour: '2-digit', 
          minute: '2-digit' 
        }),
        isOwn: currentUserId !== null && msg.sender_id === currentUserId,
        user: {
          name: msg.sender?.username || "Пользователь"
        }
      }));
      
      setMessages(formattedMessages);
    } catch (error: any) {
      console.error('Ошибка загрузки сообщений:', error);
      setMessages([]);
      
      // Используем общую функцию форматирования ошибок с кастомизацией
      let { title, description } = formatError(error, "Ошибка загрузки сообщений", "Не удалось загрузить сообщения чата");
      
      // Специфичные для загрузки сообщений ошибки
      if (error?.response?.status === 401) {
        description = "Требуется авторизация для получения сообщений";
      } else if (error?.response?.status === 403) {
        description = "У вас нет доступа к этому чату";
      } else if (error?.response?.status === 404) {
        title = "Чат не найден";
        description = "Запрашиваемый чат не существует или был удален";
      }
      
      toast({
        title,
        description,
        variant: "destructive",
      });
    } finally {
      setIsMessagesLoading(false);
    }
  };
  
  // Эффект для загрузки сообщений при смене чата
  useEffect(() => {
    if (selectedChat) {
      loadMessages(selectedChat);
    } else {
      setMessages([]);
    }
  }, [selectedChat]);

  const handleSendMessage = async () => {
    if (message.trim() && selectedChat) {
      try {
        console.log('Отправка сообщения:', message);
        const sentMessage = await chatAPI.sendMessage(selectedChat.toString(), message.trim());
        console.log('Сообщение отправлено:', sentMessage);
        
        // Проверка на ответ с ошибкой
        if (sentMessage && (sentMessage as any).error) {
          throw new Error((sentMessage as any).error);
        }
        
        // Обновляем список сообщений
        await loadMessages(selectedChat);
        
        // Очищаем поле ввода
        setMessage('');
        
        // Показываем успешное уведомление (опционально)
        // toast({
        //   title: "Сообщение отправлено",
        //   description: "Ваше сообщение успешно доставлено",
        // });
        
      } catch (error: any) {
        console.error('Ошибка отправки сообщения:', error);
        
        // Используем общую функцию форматирования ошибок с кастомизацией
        let { title, description } = formatError(error, "Ошибка отправки", "Не удалось отправить сообщение");
        
        // Специфичные для отправки сообщений ошибки
        if (error?.response?.status === 401) {
          description = "Требуется повторная авторизация для отправки сообщений";
        } else if (error?.response?.status === 403) {
          description = "У вас нет прав для отправки сообщений в этот чат";
        } else if (error?.response?.status === 404) {
          title = "Чат не найден";
          description = "Выбранный чат больше не существует";
        } else if (error?.response?.status === 413) {
          title = "Сообщение слишком большое";
          description = "Размер сообщения превышает допустимый лимит";
        } else if (error?.response?.status === 429) {
          description = "Подождите немного перед отправкой следующего сообщения";
        }
        
        toast({
          title,
          description,
          variant: "destructive",
        });
      }
    }
  };

  const handleChatHeaderClick = () => {
    const currentChat = chats.find(c => c.id === selectedChat);
    if (currentChat && currentChat.isGroup) {
      setShowGroupSettings(true);
    }
  };

  // Функция для выбора чата с сохранением в localStorage
  const handleSelectChat = (chatId: number) => {
    setSelectedChat(chatId);
    localStorage.setItem('lastSelectedChat', chatId.toString());
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex">
      {/* Боковая панель */}
      <MessengerSidebar 
        chats={chats}
        selectedChat={selectedChat}
        onSelectChat={handleSelectChat}
        onShowProfile={() => setShowProfile(!showProfile)}
        onCreateChat={handleCreateChat}
        onRefreshChats={loadChats}
        isLoading={isLoading}
      />

      {/* Основная область чата */}
      <div className="flex-1 flex flex-col">
        {selectedChat ? (
          <>
            {/* Заголовок чата */}
            <div 
              className={`bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-4 ${
                chats.find(c => c.id === selectedChat)?.isGroup ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700' : ''
              }`}
              onClick={handleChatHeaderClick}
            >
              <div className="flex items-center space-x-3">
                <Avatar className="h-8 w-8">
                  <AvatarFallback className="bg-blue-500 text-white text-sm">
                    {chats.find(c => c.id === selectedChat)?.isGroup ? 
                      <Users className="h-4 w-4" /> : 
                      chats.find(c => c.id === selectedChat)?.name.charAt(0)
                    }
                  </AvatarFallback>
                </Avatar>
                <div>
                  <h2 className="font-medium text-gray-900 dark:text-white">
                    {chats.find(c => c.id === selectedChat)?.name}
                  </h2>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {chats.find(c => c.id === selectedChat)?.isGroup 
                      ? `${groupUsers.filter(u => u.online).length} участников в сети`
                      : chats.find(c => c.id === selectedChat)?.online ? 'в сети' : 'был в сети недавно'
                    }
                  </p>
                </div>
              </div>
            </div>

            {/* Сообщения */}
            {isMessagesLoading ? (
              <div className="flex-1 flex items-center justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : (
              <MessageList messages={messages} />
            )}

            {/* Поле ввода */}
            <MessageInput 
              message={message}
              onMessageChange={setMessage}
              onSendMessage={handleSendMessage}
            />
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center bg-gray-50 dark:bg-gray-900">
            {isLoading ? (
              <div className="text-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
                  Загрузка чатов...
                </h3>
              </div>
            ) : (
              <div className="text-center">
                <MessageSquare className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
                  Добро пожаловать в мессенджер!
                </h3>
                <p className="text-gray-500 dark:text-gray-400 mb-4">
                  У вас пока нет чатов. Найдите пользователей через поиск в боковой панели.
                </p>
                <div className="text-sm text-gray-400 space-y-1">
                  <p>• Используйте поиск по логину или email</p>
                  <p>• Создавайте групповые чаты</p>
                  <p>• Все сообщения зашифрованы</p>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Панель профиля */}
      {showProfile && (
        <div className="w-80 bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 p-4">
          <div className="text-center">
            <Avatar className="h-20 w-20 mx-auto mb-4">
              <AvatarFallback className="bg-blue-500 text-white text-xl">И</AvatarFallback>
            </Avatar>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white">Иван Иванов</h3>
            <p className="text-sm text-gray-500 dark:text-gray-400">@ivanov</p>
          </div>
          
          <div className="mt-6 space-y-4">
            <Button variant="outline" className="w-full">
              Редактировать профиль
            </Button>
            <Button variant="outline" className="w-full">
              Настройки
            </Button>
            <Button variant="outline" className="w-full text-red-600 hover:text-red-700">
              Выйти
            </Button>
          </div>
        </div>
      )}

      {/* Модальное окно настроек группы */}
      <GroupSettingsModal
        isOpen={showGroupSettings}
        onClose={() => setShowGroupSettings(false)}
        groupName={chats.find(c => c.id === selectedChat)?.name || ''}
        users={groupUsers}
      />
    </div>
  );
};
