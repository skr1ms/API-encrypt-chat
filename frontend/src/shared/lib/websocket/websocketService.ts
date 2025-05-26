
import { store } from '@/app/store';
import { connectStart, connectSuccess, connectFailure, disconnect, messageReceived } from '@/shared/store/slices/websocketSlice';
import { addMessage, updateChatLastMessage } from '@/shared/store/slices/chatSlice';

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  constructor(url: string = 'ws://localhost:8080/ws') {
    this.url = url;
  }

  connect(token?: string): void {
    store.dispatch(connectStart());
    
    try {
      const wsUrl = token ? `${this.url}?token=${token}` : this.url;
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
        this.connect();
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
      store.dispatch(messageReceived(data));
      
      switch (data.type) {
        case 'message':
          store.dispatch(addMessage(data.payload));
          store.dispatch(updateChatLastMessage({
            chatId: data.payload.chatId,
            message: data.payload
          }));
          break;
        case 'user_status':
          // Handle user status updates
          break;
        case 'typing':
          // Handle typing indicators
          break;
        default:
          console.log('Unknown message type:', data.type);
      }
    } catch (error) {
      console.error('Error parsing WebSocket message:', error);
    }
  }

  sendMessage(type: string, payload: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }));
    } else {
      console.error('WebSocket is not connected');
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    store.dispatch(disconnect());
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

export const websocketService = new WebSocketService();
