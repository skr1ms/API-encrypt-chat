import { useState, useCallback } from 'react';
import { userAPI, User, SearchUsersResponse } from '@/shared/api/userApi';
import { useToast } from '@/hooks/use-toast';

interface UseUserSearchReturn {
  users: User[];
  isLoading: boolean;
  error: string | null;
  searchUsers: (query: string) => Promise<void>;
  clearResults: () => void;
}

export const useUserSearch = (): UseUserSearchReturn => {
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { toast } = useToast();
  const searchUsers = useCallback(async (query: string) => {
    if (!query.trim()) {
      setUsers([]);
      return;
    }

    setIsLoading(true);
    setError(null);    try {
      const response: SearchUsersResponse = await userAPI.searchUsers(query.trim(), 10);
      setUsers(response.users);
    } catch (err) {
      console.error('Search error:', err);
      const errorMessage = err instanceof Error ? err.message : 'Ошибка при поиске пользователей';
      setError(errorMessage);
      toast({
        title: 'Ошибка поиска',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setIsLoading(false);
    }
  }, [toast]);

  const clearResults = useCallback(() => {
    setUsers([]);
    setError(null);
  }, []);

  return {
    users,
    isLoading,
    error,
    searchUsers,
    clearResults,
  };
};
