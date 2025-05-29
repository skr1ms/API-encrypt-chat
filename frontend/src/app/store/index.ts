
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '@/shared/store/slices/authSlice';
import chatReducer from '@/shared/store/slices/chatSlice';
import websocketReducer from '@/shared/store/slices/websocketSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    chat: chatReducer,
    websocket: websocketReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: ['websocket/connect', 'websocket/disconnect'],
      },
    }),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
