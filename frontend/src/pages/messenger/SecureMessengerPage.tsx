import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Button } from '@/shared/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { ResizablePanels } from '@/components/ui/resizable-panels';
import { MessengerSidebar } from '@/widgets/messenger/sidebar';
import { MessageList } from '@/features/messenger/ui/message-list';
import { MessageInput } from '@/features/messenger/ui/message-input';
import EncryptionManager from '@/components/ui/encryption-manager';
import { EncryptionProvider, useEncryption, useEncryptionReady } from '@/shared/providers/encryption-provider';
import { useSecureChatAPI } from '@/shared/api/secure-chat-api';
import { useToast } from '@/hooks/use-toast';
import { Shield, ShieldCheck, AlertTriangle, MessageSquare, Users, X } from 'lucide-react';
import type { Chat, Message, User } from '@/shared/api/secure-chat-api';

// Внутренний компонент мессенджера, который использует контекст шифрования
const SecureMessengerContent: React.FC = () => {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { isInitialized, error: encryptionError } = useEncryption();
  const isEncryptionReady = useEncryptionReady();
  const chatAPI = useSecureChatAPI();

  // Состояние приложения
  const [selectedChat, setSelectedChat] = useState<number | null>(null);
  const [message, setMessage] = useState('');
  const [showProfile, setShowProfile] = useState(true);
  const [chats, setChats] = useState<Chat[]>([]);
  const [messages, setMessages] = useState<Message[]>([]);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isMessagesLoading, setIsMessagesLoading] = useState(false);

  // Загрузка профиля пользователя
  useEffect(() => {
    const loadUserProfile = async () => {
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          navigate('/login');
          return;
        }

        const user = await chatAPI.getProfile();
        setCurrentUser(user);
      } catch (error) {
        console.error('Failed to load user profile:', error);
        toast({
          title: 'Authentication Error',
          description: 'Failed to load user profile. Please login again.',
          variant: 'destructive',
        });
        navigate('/login');
      }
    };

    loadUserProfile();
  }, [chatAPI, navigate, toast]);

  // Загрузка чатов после успешной инициализации шифрования
  useEffect(() => {
    if (!isEncryptionReady || !currentUser) return;

    const loadChats = async () => {
      try {
        setIsLoading(true);
        const userChats = await chatAPI.getUserChats();
        setChats(userChats);
      } catch (error) {
        console.error('Failed to load chats:', error);
        toast({
          title: 'Error',
          description: 'Failed to load chats',
          variant: 'destructive',
        });
      } finally {
        setIsLoading(false);
      }
    };

    loadChats();
  }, [isEncryptionReady, currentUser, chatAPI, toast]);

  // Загрузка сообщений при выборе чата
  useEffect(() => {
    if (!selectedChat || !isEncryptionReady) return;

    const loadMessages = async () => {
      try {
        setIsMessagesLoading(true);
        const chatMessages = await chatAPI.getChatMessages(selectedChat);
        setMessages(chatMessages);
      } catch (error) {
        console.error('Failed to load messages:', error);
        toast({
          title: 'Error',
          description: 'Failed to load messages',
          variant: 'destructive',
        });
      } finally {
        setIsMessagesLoading(false);
      }
    };

    loadMessages();
  }, [selectedChat, isEncryptionReady, chatAPI, toast]);

  const handleSendMessage = async () => {
    if (!message.trim() || !selectedChat || !isEncryptionReady) return;

    try {
      const newMessage = await chatAPI.sendMessage(selectedChat, {
        content: message.trim(),
        message_type: 'text',
      });

      setMessages(prev => [...prev, newMessage]);
      setMessage('');

      toast({
        title: 'Message Sent',
        description: 'Your encrypted message has been sent successfully.',
      });
    } catch (error) {
      console.error('Failed to send message:', error);
      toast({
        title: 'Error',
        description: 'Failed to send message',
        variant: 'destructive',
      });
    }
  };

  const handleCreatePrivateChat = async (userId: number) => {
    if (!isEncryptionReady) {
      toast({
        title: 'Encryption Required',
        description: 'Please initialize encryption before creating chats.',
        variant: 'destructive',
      });
      return;
    }

    try {
      const chat = await chatAPI.createOrGetPrivateChat({ user_id: userId });
      setChats(prev => {
        const existing = prev.find(c => c.id === chat.id);
        if (existing) return prev;
        return [...prev, chat];
      });
      setSelectedChat(chat.id);

      toast({
        title: 'Chat Created',
        description: 'Private chat has been created with end-to-end encryption.',
      });
    } catch (error) {
      console.error('Failed to create private chat:', error);
      toast({
        title: 'Error',
        description: 'Failed to create private chat',
        variant: 'destructive',
      });
    }
  };

  const handleLogout = async () => {
    try {
      await chatAPI.logout();
      localStorage.removeItem('token');
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
      // Очищаем токен даже если запрос не удался
      localStorage.removeItem('token');
      navigate('/login');
    }
  };

  const selectedChatData = chats.find(chat => chat.id === selectedChat);

  if (!currentUser) {
    return <div className="flex items-center justify-center h-screen">Loading...</div>;
  }

  return (
    <div className="h-screen flex flex-col">
      {/* Статус шифрования */}
      <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="flex items-center justify-between p-4">
          <div className="flex items-center gap-3">
            <Shield className="h-5 w-5" />
            <span className="font-semibold">SleekChat</span>
            {isInitialized ? (
              <Badge variant="default" className="flex items-center gap-1">
                <ShieldCheck className="h-3 w-3" />
                Encrypted
              </Badge>
            ) : (
              <Badge variant="destructive" className="flex items-center gap-1">
                <AlertTriangle className="h-3 w-3" />
                Not Encrypted
              </Badge>
            )}
          </div>
          
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={handleLogout}>
              Logout
            </Button>
          </div>
        </div>
      </div>

      {/* Основное содержимое */}
      <div className="flex-1 flex">
        {/* Панель инициализации шифрования */}
        {!isEncryptionReady && (
          <div className="w-96 border-r p-4">
            <EncryptionManager
              userId={currentUser.id}
              onEncryptionReady={() => {
                toast({
                  title: 'Encryption Ready',
                  description: 'You can now safely send encrypted messages.',
                });
              }}
              onEncryptionError={(error) => {
                toast({
                  title: 'Encryption Error',
                  description: error,
                  variant: 'destructive',
                });
              }}
            />
          </div>
        )}

        <ResizablePanels>
          {/* Боковая панель чатов */}
          <div className="min-w-80 border-r">
            <MessengerSidebar
              chats={chats}
              selectedChat={selectedChat}
              onChatSelect={setSelectedChat}
              onCreatePrivateChat={handleCreatePrivateChat}
              currentUser={currentUser}
              isLoading={isLoading}
              encryptionReady={isEncryptionReady}
            />
          </div>

          {/* Область чата */}
          <div className="flex-1 flex flex-col">
            {selectedChat ? (
              <>
                {/* Заголовок чата */}
                <div className="border-b p-4 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Avatar className="h-8 w-8">
                      <AvatarFallback>
                        {selectedChatData?.is_group ? (
                          <Users className="h-4 w-4" />
                        ) : (
                          selectedChatData?.name.slice(0, 2).toUpperCase()
                        )}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <h3 className="font-semibold">{selectedChatData?.name}</h3>
                      {isEncryptionReady && (
                        <p className="text-xs text-muted-foreground flex items-center gap-1">
                          <ShieldCheck className="h-3 w-3" />
                          End-to-end encrypted
                        </p>
                      )}
                    </div>
                  </div>
                </div>

                {/* Список сообщений */}
                <div className="flex-1 overflow-hidden">
                  <MessageList
                    messages={messages}
                    currentUserId={currentUser.id}
                    isLoading={isMessagesLoading}
                  />
                </div>

                {/* Поле ввода сообщения */}
                <div className="border-t p-4">
                  {isEncryptionReady ? (
                    <MessageInput
                      message={message}
                      onMessageChange={setMessage}
                      onSendMessage={handleSendMessage}
                      placeholder="Type an encrypted message..."
                    />
                  ) : (
                    <Alert>
                      <AlertTriangle className="h-4 w-4" />
                      <AlertDescription>
                        Initialize encryption to send secure messages.
                      </AlertDescription>
                    </Alert>
                  )}
                </div>
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center">
                  <MessageSquare className="h-12 w-12 mx-auto text-muted-foreground" />
                  <h3 className="mt-4 text-lg font-semibold">Select a chat</h3>
                  <p className="text-muted-foreground">
                    Choose a conversation to start messaging securely
                  </p>
                </div>
              </div>
            )}
          </div>
        </ResizablePanels>
      </div>
    </div>
  );
};

// Главный компонент с провайдером шифрования
export const SecureMessengerPage: React.FC = () => {
  return (
    <EncryptionProvider
      baseURL={process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1'}
      showToasts={true}
    >
      <SecureMessengerContent />
    </EncryptionProvider>
  );
};

export default SecureMessengerPage;
