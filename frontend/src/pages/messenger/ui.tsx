import * as React from 'react';
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Button } from '@/shared/ui/button';
import { ResizablePanels } from '@/components/ui/resizable-panels';
import { MessengerSidebar } from '@/widgets/messenger/sidebar';
import { MessageList } from '@/features/messenger/ui/message-list';
import { MessageInput } from '@/features/messenger/ui/message-input';
import { GroupSettingsModal } from '@/widgets/messenger/group-settings-modal';
import { CreateGroupModal } from '@/widgets/messenger/create-group-modal';
import { ChatHeaderMenu } from '@/widgets/messenger/chat-header-menu';
import { MessageSquare, Users, X } from 'lucide-react';
import { chatAPI } from '@/shared/api/chatApi';
import { useToast } from '@/hooks/use-toast';
import { ChangePasswordModal } from '../../widgets/messenger/change-password-modal';
import { websocketService } from '@/shared/lib/websocket/websocketService';

interface Chat {
  id: number;
  name: string;
  lastMessage: string;
  time: string;
  unread: number;
  online: boolean;
  isGroup?: boolean;
  isCreator?: boolean;
}

export const MessengerPage = () => {
  const navigate = useNavigate();
  const [selectedChat, setSelectedChat] = useState<number | null>(null);
  const [message, setMessage] = useState('');
  const [showProfile, setShowProfile] = useState(true);
  const [showGroupSettings, setShowGroupSettings] = useState(false);
  const [showChangePassword, setShowChangePassword] = useState(false);
  const [showCreateGroup, setShowCreateGroup] = useState(false);
  const [chats, setChats] = useState<Chat[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { toast } = useToast();

  const [groupUsers, setGroupUsers] = useState<any[]>([]);
  const [currentUserId, setCurrentUserId] = useState<number | null>(null);
  const [currentUser, setCurrentUser] = useState<any>(null);

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

  // Загрузка профиля пользователя
  const loadUserProfile = async () => {
    try {
      // Всегда делаем запрос к API, чтобы получить актуальные данные пользователя
      const profile = await chatAPI.getProfile();
      
      if (profile?.data) {
        // Обновляем данные пользователя в localStorage
        localStorage.setItem('user', JSON.stringify(profile.data));
        setCurrentUser(profile.data);
      } else {
        // Если API не вернул данные, попробуем использовать данные из localStorage
        const userDataString = localStorage.getItem('user');
        if (userDataString) {
          const userData = JSON.parse(userDataString);
          setCurrentUser(userData);
        }
      }
    } catch (error) {
      console.error('Ошибка загрузки профиля:', error);
      
      // Если запрос не удался, используем данные из localStorage
      const userDataString = localStorage.getItem('user');
      if (userDataString) {
        try {
          const userData = JSON.parse(userDataString);
          setCurrentUser(userData);
        } catch (e) {
          console.error('Ошибка при разборе данных пользователя из localStorage:', e);
        }
      }
    }
  };

  // Функция выхода из системы
  const handleLogout = async () => {
    try {
      await chatAPI.logout();
      
      // Очищаем все данные из localStorage
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('ecdhPrivateKey');
      localStorage.removeItem('ecdsaPrivateKey');
      localStorage.removeItem('rsaPrivateKey');
      localStorage.removeItem('lastSelectedChat');
      
      // Отключаемся от WebSocket
      websocketService.disconnect();
      
      toast({
        title: "Выход выполнен",
        description: "Вы успешно вышли из системы",
      });
      
      // Перенаправляем на страницу входа
      navigate('/login');
    } catch (error) {
      console.error('Ошибка при выходе:', error);
      
      // Даже если сервер недоступен, очищаем локальные данные
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('ecdhPrivateKey');
      localStorage.removeItem('ecdsaPrivateKey');
      localStorage.removeItem('rsaPrivateKey');
      localStorage.removeItem('lastSelectedChat');
      
      websocketService.disconnect();
      
      toast({
        title: "Выход выполнен",
        description: "Локальные данные очищены",
      });
      
      navigate('/login');
    }
  };

  // Инициализация пользователя при монтировании компонента
  useEffect(() => {
    loadUserProfile();
    
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

    // Обработчик для системных уведомлений из WebSocket
    const handleChatNotification = (e: any) => {
      const notification = e.detail;
      
      const chatId = notification.chatId ? Number(notification.chatId) : undefined;
      if (!chatId) return;
      
      if (notification.type === 'user_left') {
        // Пользователь покинул группу - обновляем список чатов
        loadChats();
        
        // Добавляем системное сообщение в текущий чат, если это текущий открытый чат
        if (selectedChat === chatId) {
          addSystemMessage(chatId, notification.message || 'Пользователь покинул группу', 'user_left');
        }
        
        // Показываем уведомление в любом случае
        toast({
          title: "Пользователь покинул группу",
          description: notification.message,
        });
      }
      else if (notification.type === 'user_removed') {
        // Пользователь был удален из группы администратором/создателем
        loadChats();
        
        // Добавляем системное сообщение в текущий чат, если это текущий открытый чат
        if (selectedChat === chatId) {
          addSystemMessage(chatId, notification.message || 'Пользователь был удален из группы', 'user_removed');
          
          // Обновляем список участников группы
          loadGroupMembers(chatId);
        }
        
        // Показываем уведомление в любом случае
        toast({
          title: "Пользователь удален из группы",
          description: notification.message,
          variant: "destructive",
        });
      }
      else if (notification.type === 'user_joined') {
        
        if (selectedChat === chatId) {
          addSystemMessage(chatId, notification.message || 'Пользователь присоединился к группе', 'user_joined');
          
          loadGroupMembers(chatId);
        }
        
        loadChats();
        
        toast({
          title: "Новый участник",
          description: notification.message,
        });
      }
      else if (notification.type === 'group_created') {
        loadChats();
        
        toast({
          title: "Создана новая группа",
          description: notification.message,
        });
      }
      else if (notification.type === 'group_deleted') {
        // Группа была удалена создателем
        
        // Показываем всплывающее уведомление
        toast({
          title: "Группа удалена",
          description: notification.message,
          variant: "destructive",
        });
        
        // Если это текущий выбранный чат, добавляем сообщение и сбрасываем выбор
        if (selectedChat === chatId) {
          // Добавляем сообщение перед сбросом чата
          addSystemMessage(chatId, notification.message || 'Группа была удалена', 'group_deleted');
          
          // Делаем небольшую задержку перед сбросом, чтобы пользователь увидел сообщение
          setTimeout(() => {
            setSelectedChat(null);
            localStorage.removeItem('lastSelectedChat');
            // Обновляем список чатов
            loadChats();
          }, 2000);
        } else {
          // Если это не текущий чат, просто обновляем список
          loadChats();
        }
      }
    };
    
    // Обработчик для обновления списка чатов
    const handleRefreshChats = () => {
      console.log('Refresh chats event received');
      // Добавляем небольшую задержку перед обновлением чатов для обеспечения 
      // завершения всех операций на сервере
      setTimeout(() => {
        loadChats();
      }, 300);
    };
    
    // Добавляем обработчики
    window.addEventListener('chatNotification', handleChatNotification);
    window.addEventListener('refreshChats', handleRefreshChats);
    
    return () => {
      websocketService.disconnect();
      window.removeEventListener('chatNotification', handleChatNotification);
      window.removeEventListener('refreshChats', handleRefreshChats);
    };
  }, []);

  // Загрузка чатов после установки currentUserId
  useEffect(() => {
    if (currentUserId !== null) {
      loadChats();
    }
  }, [currentUserId]);

  const loadChats = async () => {
    try {
      setIsLoading(true);
      console.log('Starting to load chats...');
      const response = await chatAPI.getChats();
      console.log('Raw response from API:', response);
      
      // Проверка на ответ с ошибкой
      if (response && (response as any).error) {
        throw new Error((response as any).error);
      }
      
      // Преобразуем данные с бэкенда в формат для фронтенда
      const chatsData = Array.isArray(response) ? response : (response as any)?.data || [];
      console.log('Processed chats data:', chatsData);
      
      const formattedChats: Chat[] = await Promise.all(chatsData.map(async (chat: any) => {
        console.log('Processing chat:', chat);
        
        let isCreator = false;
        
        // Для групповых чатов загружаем информацию о участниках, чтобы определить создателя
        if (chat.is_group && currentUserId) {
          try {
            const membersResponse = await chatAPI.getChatMembers(chat.id.toString());
            const membersData = Array.isArray(membersResponse) ? membersResponse : (membersResponse as any)?.data || [];
            
            console.log(`Members data for chat ${chat.id}:`, membersData);
            
            // Проверяем, является ли текущий пользователь создателем группы
            // Способ 1: Проверка через поле role
            const currentUserMember = membersData.find((member: any) => 
              (member.user_id || member.id) === currentUserId
            );
            isCreator = currentUserMember?.role === 'creator';
            
            // Способ 2: Сравниваем с created_by в чате
            if (chat.created_by === currentUserId) {
              isCreator = true;
            }
            
            console.log(`Chat ${chat.id} - Current user is creator: ${isCreator}`);
          } catch (error) {
            console.error(`Ошибка загрузки участников для чата ${chat.id}:`, error);
            // В случае ошибки проверяем поле created_by
            if (chat.created_by === currentUserId) {
              isCreator = true;
              console.log(`Fallback: Chat ${chat.id} - Current user is creator based on created_by`);
            } else {
              isCreator = false;
            }
          }
        }
        
        return {
          id: chat.id,
          name: chat.name,
          lastMessage: chat.last_message || 'Нет сообщений',
          time: chat.updated_at ? new Date(chat.updated_at).toLocaleTimeString('ru-RU', { 
            hour: '2-digit', 
            minute: '2-digit' 
          }) : '',
          unread: 0, // TODO: реализовать подсчет непрочитанных
          online: false, // TODO: реализовать статус онлайн для приватных чатов
          isGroup: chat.is_group,
          isCreator: isCreator
        };
      }));
      
      console.log('Final formatted chats:', formattedChats);
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
      
      // Добавляем новый чат прямо в список без полной перезагрузки
      if (chatData) {
        const newChat = {
          id: chatData.id,
          name: chatName,
          lastMessage: 'Чат создан',
          time: new Date().toLocaleTimeString('ru-RU', { 
            hour: '2-digit', 
            minute: '2-digit' 
          }),
          unread: 0,
          online: false,
          isGroup: isGroup,
          isCreator: isGroup // Пользователь всегда является создателем группы, которую он только что создал
        };
        
        setChats(prevChats => [newChat, ...prevChats]);
        
        // Автоматически выбираем созданный чат
        setSelectedChat(chatData.id);
        
        // Сохраняем ID последнего выбранного чата
        localStorage.setItem('lastSelectedChat', chatData.id.toString());
        
        // Вызываем событие обновления чатов для синхронизации данных
        const refreshEvent = new CustomEvent('refreshChats');
        window.dispatchEvent(refreshEvent);
        
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
  
  const handleCreateGroup = async (groupName: string, participants: number[]) => {
    try {
      await handleCreateChat(groupName, true, participants);
      setShowCreateGroup(false);
    } catch (error) {
      console.error('Ошибка создания группы:', error);
      // Ошибка уже обработана in handleCreateChat
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
      
      // Сортируем сообщения по времени создания (старые -> новые)
      const sortedMessages = messagesData.sort((a: any, b: any) => {
        return new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      });
      
      const formattedMessages = sortedMessages.map((msg: any) => {
        // Проверяем, является ли сообщение системным
        const isSystem = msg.sender_id === 0 || msg.is_system;
        // Если не системное, проверяем принадлежность текущему пользователю
        const isOwn = !isSystem && currentUserId !== null && Number(msg.sender_id) === Number(currentUserId);

        // Определяем тип системного сообщения по тексту или message_type
        let systemType: 'user_left' | 'group_deleted' | 'user_joined' | 'user_removed' | 'system' = 'system';
        if (isSystem) {
          const text = (msg.decrypted_content || msg.content || '').toLowerCase();
          if (msg.message_type && msg.message_type !== 'system') {
            systemType = msg.message_type;
          } else if (text.includes('покинул')) {
            systemType = 'user_left';
          } else if (text.includes('удален') && text.includes('создател')) {
            systemType = 'group_deleted';
          } else if (text.includes('удален')) {
            systemType = 'user_removed';
          } else if (text.includes('присоединился') || text.includes('присоединилась')) {
            systemType = 'user_joined';
          }
        }

        return {
          id: msg.id,
          text: msg.decrypted_content || msg.content || "Зашифрованное сообщение",
          time: new Date(msg.created_at).toLocaleTimeString('ru-RU', { 
            hour: '2-digit', 
            minute: '2-digit' 
          }),
          isOwn: isOwn,
          isSystem: isSystem,
          type: isSystem ? systemType : undefined,
          user: {
            name: isSystem ? "Система" : (msg.sender?.username || "Пользователь")
          }
        };
      });
      
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
      const messageToSend = message.trim();
      
      // Добавляем сообщение локально сразу для лучшего UX
      const optimisticMessage = {
        id: Date.now(),
        text: messageToSend,
        time: new Date().toLocaleTimeString('ru-RU', { 
          hour: '2-digit', 
          minute: '2-digit' 
        }),
        isOwn: true,
        user: {
          name: 'Вы'
        }
      };
      
      try {
        console.log('Отправка сообщения:', messageToSend);
        
        setMessages(prev => [...prev, optimisticMessage]);
        
        // Очищаем поле ввода сразу
        setMessage('');
        
        const sentMessage = await chatAPI.sendMessage(selectedChat.toString(), messageToSend);
        console.log('Сообщение отправлено:', sentMessage);
        
        // Проверка на ответ с ошибкой
        if (sentMessage && (sentMessage as any).error) {
          throw new Error((sentMessage as any).error);
        }
        
        // Обновляем список сообщений с сервера для получения правильного ID
        await loadMessages(selectedChat);
        
      } catch (error: any) {
        console.error('Ошибка отправки сообщения:', error);
        
        // Удаляем оптимистичное сообщение в случае ошибки
        setMessages(prev => prev.filter(msg => msg.id !== optimisticMessage.id));
        
        // Восстанавливаем текст в поле ввода
        setMessage(messageToSend);
        
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
      loadGroupMembers(selectedChat);
      setShowGroupSettings(true);
    }
  };

  // Функция для загрузки участников группы
  const loadGroupMembers = async (chatId?: number) => {
    const currentChatId = chatId || selectedChat;
    if (!currentChatId) return;
    
    try {
      const response = await chatAPI.getChatMembers(currentChatId.toString());
      console.log('Loaded group members:', response);
      
      const membersData = Array.isArray(response) ? response : (response as any)?.data || [];
      
      // Форматируем участников для компонента
      const formattedMembers = membersData.map((member: any) => ({
        id: member.user_id || member.id,
        name: member.username || member.name,
        online: member.is_online || false,
        // Получаем роль пользователя (creator, admin, member)
        role: member.role || 'member'
      }));
      
      setGroupUsers(formattedMembers);
    } catch (error: any) {
      console.error('Ошибка загрузки участников группы:', error);
      
      const { title, description } = formatError(error, "Ошибка загрузки", "Не удалось загрузить список участников группы");
      
      toast({
        title,
        description,
        variant: "destructive",
      });
    }
  };

  // Функция для добавления участника в группу
  const handleAddMember = async (userId: number) => {
    if (!selectedChat) return;
    
    try {
      const response = await chatAPI.addMember(selectedChat.toString(), userId);
      
      // Получаем данные о добавленном пользователе
      const userData = response?.data?.user || response?.user;
      
      if (userData) {
        // Добавляем пользователя локально в список участников группы
        const newUser = {
          id: userData.id,
          name: userData.username || userData.name,
          online: userData.is_online || false
        };
        
        setGroupUsers(prevUsers => {
          // Проверяем, не существует ли уже такой пользователь
          const userExists = prevUsers.some(user => user.id === newUser.id);
          if (!userExists) {
            return [...prevUsers, newUser];
          }
          return prevUsers;
        });
      } else {
        // Если не получили данные о пользователе, обновляем весь список
        await loadGroupMembers();
      }
      
      toast({
        title: "Участник добавлен",
        description: "Пользователь успешно добавлен в группу",
      });
    } catch (error: any) {
      console.error('Ошибка добавления участника:', error);
      
      const { title, description } = formatError(error, "Ошибка добавления", "Не удалось добавить участника в группу");
      
      toast({
        title,
        description,
        variant: "destructive",
      });
      
      throw error;
    }
  };

  // Функция для удаления участника из группы
  const handleRemoveMember = async (userId: number) => {
    if (!selectedChat) return;
    
    try {
      await chatAPI.removeMember(selectedChat.toString(), userId);
      
      // После успешного удаления, перезагружаем данные о группе
      await loadGroupMembers(selectedChat);
      
      toast({
        title: "Участник удален",
        description: "Пользователь удален из группы",
      });
    } catch (error: any) {
      console.error('Ошибка удаления участника:', error);
      
      const { title, description } = formatError(error, "Ошибка удаления", "Не удалось удалить участника из группы");
      
      toast({
        title,
        description,
        variant: "destructive",
      });
      
      throw error;
    }
  };

  // Функция для назначения администратора
  const handleSetAdmin = async (userId: number) => {
    if (!selectedChat) return;
    
    try {
      await chatAPI.setAdmin(selectedChat.toString(), userId);
      
      // После успешного назначения, перезагружаем данные о группе
      await loadGroupMembers(selectedChat);
      
      toast({
        title: "Администратор назначен",
        description: "Пользователь назначен администратором группы",
      });
    } catch (error: any) {
      console.error('Ошибка назначения администратора:', error);
      
      const { title, description } = formatError(error, "Ошибка назначения", "Не удалось назначить пользователя администратором");
      
      toast({
        title,
        description,
        variant: "destructive",
      });
      
      throw error;
    }
  };
  
  // Функция для снятия прав администратора
  const handleRemoveAdmin = async (userId: number) => {
    if (!selectedChat) return;
    
    try {
      await chatAPI.removeAdmin(selectedChat.toString(), userId);
      
      // После успешного снятия прав, перезагружаем данные о группе
      await loadGroupMembers(selectedChat);
      
      toast({
        title: "Права администратора сняты",
        description: "Пользователь больше не является администратором",
      });
    } catch (error: any) {
      console.error('Ошибка снятия прав администратора:', error);
      
      const { title, description } = formatError(error, "Ошибка управления правами", "Не удалось снять права администратора");
      
      toast({
        title,
        description,
        variant: "destructive",
      });
      
      throw error;
    }
  };

  // Функция для выбора чата с сохранением в localStorage
  const handleSelectChat = (chatId: number) => {
    setSelectedChat(chatId);
    localStorage.setItem('lastSelectedChat', chatId.toString());
    
    // Если это групповой чат, загружаем информацию о его участниках
    const chat = chats.find(c => c.id === chatId);
    if (chat?.isGroup) {
      loadGroupMembers(chatId);
    }
  };

  // Helper function to check if current user is creator of the currently selected chat
  const isCurrentUserCreator = () => {
    if (!currentUserId || !selectedChat || !groupUsers.length) return false;
    const currentUser = groupUsers.find(user => user.id === currentUserId);
    return currentUser?.role === 'creator';
  };

  // Функция для выхода из группового чата
  const handleLeaveChat = async (chatId: number) => {
    try {
      const chatName = chats.find(c => c.id === chatId)?.name || '';
      
      // ИМИТАЦИЯ: Добавляем системное сообщение для демонстрации WebSocket уведомлений
      const currentUsername = currentUser?.username || 'Пользователь';
      const messageText = `${currentUsername} покинул(а) группу "${chatName}"`;
      
      // Делаем фактический выход на backend
      await chatAPI.leaveChat(chatId.toString());
      
      // Имитируем получение WebSocket уведомления (для демонстрации)
      const notificationEvent = new CustomEvent('chatNotification', {
        detail: {
          type: 'user_left',
          chatId: chatId,
          message: messageText
        }
      });
      window.dispatchEvent(notificationEvent);
      
      // Обновляем список чатов
      await loadChats();
      
      // Если покидаем текущий выбранный чат, сбрасываем выбор
      if (selectedChat === chatId) {
        setSelectedChat(null);
        localStorage.removeItem('lastSelectedChat');
      }
      
      toast({
        title: "Вы покинули группу",
        description: "Группа удалена из ваших чатов",
      });
    } catch (error: any) {
      console.error('Ошибка выхода из группы:', error);
      
      const { title, description } = formatError(error, "Ошибка выхода", "Не удалось покинуть группу");
      
      toast({
        title,
        description,
        variant: "destructive",
      });
      
      throw error;
    }
  };

  // Функция для удаления приватного чата
  const handleDeleteChat = async (chatId: number) => {
    try {
      await chatAPI.deleteChat(chatId.toString());
      
      // Обновляем список чатов
      await loadChats();
      
      // Если удаляем текущий выбранный чат, сбрасываем выбор
      if (selectedChat === chatId) {
        setSelectedChat(null);
        localStorage.removeItem('lastSelectedChat');
      }
      
      toast({
        title: "Чат удален",
        description: "Чат удален из ваших сообщений",
      });
    } catch (error: any) {
      console.error('Ошибка удаления чата:', error);
      
      const { title, description } = formatError(error, "Ошибка удаления", "Не удалось удалить чат");
      
      toast({
        title,
        description,
        variant: "destructive",
      });
      
      throw error;
    }
  };

  // Функция для удаления группового чата (только для создателей)
  const handleDeleteGroupChat = async (chatId: number) => {
    try {
      const chatName = chats.find(c => c.id === chatId)?.name || '';
      
      // ИМИТАЦИЯ: Добавляем системное сообщение для демонстрации WebSocket уведомлений
      const currentUsername = currentUser?.username || 'Пользователь';
      const messageText = `Группа "${chatName}" была удалена создателем ${currentUsername}`;
      
      // Делаем фактическое удаление на backend
      await chatAPI.deleteGroupChat(chatId.toString());
      
      // Имитируем получение WebSocket уведомления (для демонстрации)
      const notificationEvent = new CustomEvent('chatNotification', {
        detail: {
          type: 'group_deleted',
          chatId: chatId,
          message: messageText
        }
      });
      window.dispatchEvent(notificationEvent);
      
      // Обновляем список чатов
      await loadChats();
      
      // Если удаляем текущий выбранный чат, сбрасываем выбор
      if (selectedChat === chatId) {
        setSelectedChat(null);
        localStorage.removeItem('lastSelectedChat');
      }
      
      toast({
        title: "Группа удалена",
        description: "Группа была полностью удалена для всех участников",
      });
    } catch (error: any) {
      console.error('Ошибка удаления группы:', error);
      
      const { title, description } = formatError(error, "Ошибка удаления группы", "Не удалось удалить группу");
      
      toast({
        title,
        description,
        variant: "destructive",
      });
      
      throw error;
    }
  };

  // Функция для добавления системного сообщения в чат
  const addSystemMessage = (chatId: number, text: string, type: 'user_left' | 'group_deleted' | 'user_joined' | 'user_removed' | 'system' = 'system') => {
    if (!chatId) return;
    
    const systemMessage = {
      id: `system_${Date.now()}`,
      text,
      time: new Date().toLocaleTimeString('ru-RU', { 
        hour: '2-digit', 
        minute: '2-digit' 
      }),
      isOwn: false,
      isSystem: true,
      type,
      user: {
        name: 'Система'
      }
    };
    
    // Если это текущий открытый чат, добавляем сообщение в список
    if (selectedChat === chatId) {
      setMessages(prev => [...prev, systemMessage]);
    }
  };

  return (
    <>
      <ResizablePanels
        showProfile={showProfile}
        onToggleProfile={setShowProfile}
      >
        {{
          sidebar: (
            <MessengerSidebar 
              chats={chats}
              selectedChat={selectedChat}
              onSelectChat={handleSelectChat}
              onShowProfile={() => setShowProfile(!showProfile)}
              onCreateChat={handleCreateChat}
              onRefreshChats={loadChats}
              onShowCreateGroup={() => setShowCreateGroup(true)}
              onLeaveChat={handleLeaveChat}
              onDeleteChat={handleDeleteChat}
              onDeleteGroupChat={handleDeleteGroupChat}
              currentUserId={currentUserId}
              groupUsers={groupUsers}
              isLoading={isLoading}
            />
          ),
          main: (
            <div className="flex flex-col h-full">
              {selectedChat ? (
                <>
                  {/* Заголовок чата */}
                  <div 
                    className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-4"
                  >
                    <div className="flex items-center justify-between">
                      <div 
                        className={`flex items-center space-x-3 ${
                          chats.find(c => c.id === selectedChat)?.isGroup ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg p-2 -m-2' : ''
                        }`}
                        onClick={handleChatHeaderClick}
                      >
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
                      
                      {/* Кнопка меню с тремя точками */}
                      <ChatHeaderMenu
                        chatId={selectedChat}
                        chatName={chats.find(c => c.id === selectedChat)?.name || ''}
                        isGroup={chats.find(c => c.id === selectedChat)?.isGroup || false}
                        isCreator={chats.find(c => c.id === selectedChat)?.isGroup ? isCurrentUserCreator() : false}
                        onLeaveChat={() => handleLeaveChat(selectedChat)}
                        onDeleteChat={() => handleDeleteChat(selectedChat)}
                        onDeleteGroupChat={() => handleDeleteGroupChat(selectedChat)}
                      />
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
          ),
          profile: (
            <div className="p-4 relative h-full">
              {/* Кнопка закрытия */}
              <Button
                variant="ghost"
                size="sm"
                className="absolute top-2 right-2 h-8 w-8 p-0"
                onClick={() => setShowProfile(false)}
              >
                <X className="h-4 w-4" />
              </Button>
              
              <div className="text-center mt-6">
                <Avatar className="h-20 w-20 mx-auto mb-4">
                  <AvatarFallback className="bg-blue-500 text-white text-xl">
                    {currentUser?.username ? currentUser.username.charAt(0).toUpperCase() : 'У'}
                  </AvatarFallback>
                </Avatar>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                  {currentUser?.username || 'Пользователь'}
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {currentUser?.email || 'email@example.com'}
                </p>
              </div>
              
              <div className="mt-6 space-y-4">
                <Button 
                  variant="outline" 
                  className="w-full"
                  onClick={() => setShowChangePassword(true)}
                >
                  Изменить пароль
                </Button>
                <Button 
                  variant="outline" 
                  className="w-full text-red-600 hover:text-red-700"
                  onClick={handleLogout}
                >
                  Выйти
                </Button>
              </div>
            </div>
          ),
        }}
      </ResizablePanels>

      {/* Модальные компоненты */}
      <CreateGroupModal 
        isOpen={showCreateGroup} 
        onClose={() => setShowCreateGroup(false)}
        onCreateGroup={handleCreateGroup}
      />
      
      <ChangePasswordModal 
        isOpen={showChangePassword} 
        onClose={() => setShowChangePassword(false)}
      />
      
      <GroupSettingsModal 
        isOpen={showGroupSettings} 
        onClose={() => setShowGroupSettings(false)}
        groupName={selectedChat ? chats.find(c => c.id === selectedChat)?.name || '' : ''}
        users={groupUsers}
        chatId={selectedChat}
        currentUserId={currentUserId}
        onAddMember={handleAddMember}
        onRemoveMember={handleRemoveMember}
        onSetAdmin={handleSetAdmin}
        onRemoveAdmin={handleRemoveAdmin}
      />

    </>
  );
};
