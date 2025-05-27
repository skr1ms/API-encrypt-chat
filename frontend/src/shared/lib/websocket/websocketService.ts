
import { store } from '@/app/store';
import { connectStart, connectSuccess, connectFailure, disconnect, messageReceived } from '@/shared/store/slices/websocketSlice';
import { addMessage, updateChatLastMessage } from '@/shared/store/slices/chatSlice';

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private currentToken: string | null = null;

  constructor(url: string = '/ws') {
    this.url = url;
  }

  connect(token?: string): void {
    store.dispatch(connectStart());
    
    // Сохраняем токен для переподключений
    if (token) {
      this.currentToken = token;
    }
    
    // Используем сохраненный токен если новый не передан
    const useToken = token || this.currentToken;
    
    try {
      // Создаем полный WebSocket URL
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      
      // Используем правильный WebSocket URL (nginx проксирует /ws на backend)
      const baseWsUrl = `${protocol}//${host}/ws`;
      const wsUrl = useToken ? `${baseWsUrl}?token=${encodeURIComponent(useToken)}` : baseWsUrl;
      
      console.log('Connecting to WebSocket:', wsUrl.substring(0, wsUrl.indexOf('token=') + 10) + '...');
      
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = this.handleOpen.bind(this);
      this.ws.onclose = this.handleClose.bind(this);
      this.ws.onerror = this.handleError.bind(this);
      this.ws.onmessage = this.handleMessage.bind(this);
    } catch (error) {
      store.dispatch(connectFailure(error instanceof Error ? error.message : 'Connection failed'));
    }
  }

  private handleOpen(): void {
    console.log('WebSocket connected');
    store.dispatch(connectSuccess());
    this.reconnectAttempts = 0;
  }

  private handleClose(): void {
    console.log('WebSocket disconnected');
    store.dispatch(disconnect());
    
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      setTimeout(() => {
        this.reconnectAttempts++;
        console.log(`Attempting reconnect ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
        this.connect(); // Теперь будет использовать сохраненный токен
      }, this.reconnectDelay * this.reconnectAttempts);
    }
  }

  private handleError(error: Event): void {
    console.error('WebSocket error:', error);
    store.dispatch(connectFailure('WebSocket connection error'));
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const data = JSON.parse(event.data);
      console.log('WebSocket message received:', data);
      
      store.dispatch(messageReceived(data));
      
      switch (data.type) {
        case 'chat':
          // Обрабатываем сообщения чата
          if (data.data && data.chat_id) {
            const messageData = data.data;
            
            // Диспетчим в Redux store
            const message = {
              id: messageData.id?.toString() || Date.now().toString(),
              chatId: data.chat_id.toString(),
              senderId: messageData.sender_id?.toString() || '',
              senderUsername: messageData.sender?.username || 'Пользователь',
              content: messageData.content || '',
              encrypted: false,
              timestamp: messageData.timestamp ? new Date(messageData.timestamp * 1000).toISOString() : new Date().toISOString(),
              isOwn: false // будет определено в компоненте
            };
            
            store.dispatch(addMessage(message));
            store.dispatch(updateChatLastMessage({ 
              chatId: data.chat_id.toString(), 
              message 
            }));
            
            // Также отправляем кастомное событие для компонентов, которые его слушают
            const customEvent = new CustomEvent('newMessage', {
              detail: {
                id: messageData.id,
                chatId: data.chat_id.toString(),
                content: messageData.content,
                senderId: messageData.sender_id?.toString(),
                senderUsername: messageData.sender?.username || 'Пользователь',
                timestamp: messageData.timestamp ? new Date(messageData.timestamp * 1000).toISOString() : new Date().toISOString(),
                messageType: messageData.message_type || 'text'
              }
            });
            window.dispatchEvent(customEvent);
          }
          break;
        case 'user_status':
          // Handle user status updates
          console.log('User status update:', data.data);
          break;
        case 'notification':
          console.log('Notification received:', data);
          if (data.notification) {
            // Отправляем уведомление как событие
            const notificationEvent = new CustomEvent('chatNotification', {
              detail: {
                type: data.notification.type,
                chatId: data.chatId || data.chat_id,
                message: data.notification.message,
                data: data.notification.data
              }
            });
            window.dispatchEvent(notificationEvent);
            
            // Для некоторых типов уведомлений показываем системное сообщение
            if (data.chatId || data.chat_id) {
              const chatId = data.chatId || data.chat_id;
              const messageText = data.notification.message;
              const notificationType = data.notification.type;
              
              // Создаем системное сообщение
              const systemMessage = {
                id: `system_${Date.now()}`,
                chatId: chatId.toString(),
                senderId: '0', // Системное сообщение
                senderUsername: 'Система',
                content: messageText,
                encrypted: false,
                timestamp: new Date().toISOString(),
                isSystem: true,
                isOwn: false, // Системные сообщения не принадлежат пользователю
                type: notificationType
              };
              
              store.dispatch(addMessage(systemMessage));
              
              // Отправляем также событие о необходимости обновить список чатов
              if (notificationType === 'user_left' || notificationType === 'group_deleted') {
                const refreshEvent = new CustomEvent('refreshChats');
                window.dispatchEvent(refreshEvent);
              }
              
              // Дополнительно обновляем список чатов при добавлении участников или создании группы
              if (notificationType === 'user_joined' || notificationType === 'group_created' || notificationType === 'user_left' || notificationType === 'group_deleted') {
                const refreshEvent = new CustomEvent('refreshChats');
                window.dispatchEvent(refreshEvent);
              }
            }
          }
          break;
        case 'error':
          console.error('WebSocket error message:', data.data);
          break;
        default:
          console.log('Unknown message type:', data.type, data);
      }
    } catch (error) {
      console.error('Error parsing WebSocket message:', error);
    }
  }

  sendMessage(type: string, data: any, chatId?: number, to?: number): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message = {
        type,
        data,
        chat_id: chatId,
        to,
        timestamp: Date.now()
      };
      this.ws.send(JSON.stringify(message));
    } else {
      console.error('WebSocket is not connected');
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    // Очищаем сохраненный токен
    this.currentToken = null;
    store.dispatch(disconnect());
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

export const websocketService = new WebSocketService();
