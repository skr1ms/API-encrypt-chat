import React, { useState, useCallback, useEffect } from 'react';
import { Search, User as UserIcon, Mail, Circle } from 'lucide-react';
import { Input } from '@/shared/ui/input';
import { Button } from '@/shared/ui/button';
import { Card, CardContent } from '@/shared/ui/card';
import { Avatar, AvatarFallback } from '@/shared/ui/avatar';
import { Badge } from '@/shared/ui/badge';
import { useUserSearch } from '@/shared/hooks/useUserSearch';
import { User } from '@/shared/api/userApi';

interface UserSearchProps {
  onUserSelect?: (user: User) => void;
  placeholder?: string;
  showEmail?: boolean;
  className?: string;
}

export const UserSearch: React.FC<UserSearchProps> = ({
  onUserSelect,
  placeholder = "Поиск пользователей по имени или email...",
  showEmail = true,
  className = ""
}) => {  const [query, setQuery] = useState('');
  const { users, isLoading, error, searchUsers, clearResults } = useUserSearch();

  // Custom debounce implementation
  useEffect(() => {
    const timer = setTimeout(() => {
      if (query.length >= 2) {
        searchUsers(query);
      } else {
        clearResults();
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [query, searchUsers, clearResults]);

  const handleInputChange = (value: string) => {
    setQuery(value);
  };

  const handleUserClick = (user: User) => {
    if (onUserSelect) {
      onUserSelect(user);
    }
    setQuery('');
    clearResults();
  };

  const getUserInitials = (username: string) => {
    return username.substring(0, 2).toUpperCase();
  };

  const formatLastSeen = (lastSeen?: string) => {
    if (!lastSeen) return '';
    
    const date = new Date(lastSeen);
    const now = new Date();
    const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);
    
    if (diffInHours < 1) {
      return 'недавно';
    } else if (diffInHours < 24) {
      return `${Math.floor(diffInHours)} ч. назад`;
    } else {
      return date.toLocaleDateString('ru-RU');
    }
  };

  return (
    <div className={`relative ${className}`}>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
        <Input
          type="text"
          placeholder={placeholder}
          value={query}
          onChange={(e) => handleInputChange(e.target.value)}
          className="pl-10"
        />
        {isLoading && (
          <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
          </div>
        )}
      </div>

      {/* Results dropdown */}
      {(users.length > 0 || error) && (
        <Card className="absolute top-full mt-1 w-full z-50 shadow-lg border">
          <CardContent className="p-0 max-h-64 overflow-y-auto">
            {error && (
              <div className="p-4 text-red-600 text-sm">
                {error}
              </div>
            )}
            
            {users.length > 0 && (
              <div className="py-2">
                {users.map((user) => (
                  <Button
                    key={user.id}
                    variant="ghost"
                    className="w-full justify-start p-3 h-auto hover:bg-gray-50 dark:hover:bg-gray-800"
                    onClick={() => handleUserClick(user)}
                  >
                    <div className="flex items-center space-x-3 w-full">
                      <Avatar className="h-10 w-10">
                        <AvatarFallback className="bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-300">
                          {getUserInitials(user.username)}
                        </AvatarFallback>
                      </Avatar>
                      
                      <div className="flex-1 text-left">
                        <div className="flex items-center space-x-2">
                          <span className="font-medium text-gray-900 dark:text-gray-100">
                            {user.username}
                          </span>
                          
                          <div className="flex items-center space-x-1">
                            <Circle 
                              className={`h-2 w-2 ${
                                user.is_online 
                                  ? 'text-green-500 fill-current' 
                                  : 'text-gray-400 fill-current'
                              }`}
                            />
                            <Badge 
                              variant={user.is_online ? "default" : "secondary"}
                              className="text-xs"
                            >
                              {user.is_online ? 'Онлайн' : 'Оффлайн'}
                            </Badge>
                          </div>
                        </div>
                        
                        {showEmail && (
                          <div className="flex items-center space-x-1 mt-1">
                            <Mail className="h-3 w-3 text-gray-400" />
                            <span className="text-sm text-gray-500 dark:text-gray-400">
                              {user.email}
                            </span>
                          </div>
                        )}
                        
                        {!user.is_online && user.last_seen && (
                          <div className="text-xs text-gray-400 mt-1">
                            {formatLastSeen(user.last_seen)}
                          </div>
                        )}
                      </div>
                    </div>
                  </Button>
                ))}
              </div>
            )}
            
            {query.length >= 2 && users.length === 0 && !isLoading && !error && (
              <div className="p-4 text-gray-500 text-sm text-center">
                Пользователи не найдены
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
};
