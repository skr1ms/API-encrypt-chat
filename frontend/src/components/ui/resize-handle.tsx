import React from 'react';
import { cn } from '@/lib/utils';

interface ResizeHandleProps {
  direction: 'vertical';
  position: 'left' | 'right';
  onMouseDown: (e: React.MouseEvent) => void;
  isDragging: boolean;
  className?: string;
}

export const ResizeHandle: React.FC<ResizeHandleProps> = ({
  direction,
  position,
  onMouseDown,
  isDragging,
  className,
}) => {
  return (
    <div
      className={cn(
        'relative group cursor-col-resize select-none transition-colors duration-200',
        'hover:bg-blue-100 dark:hover:bg-blue-900/20',
        isDragging && 'bg-blue-200 dark:bg-blue-800/30',
        direction === 'vertical' && 'w-1 h-full',
        className
      )}
      onMouseDown={onMouseDown}
    >
      {/* Видимая полоса */}
      <div
        className={cn(
          'absolute inset-0 bg-gray-200 dark:bg-gray-700 transition-colors duration-200',
          'group-hover:bg-blue-400 dark:group-hover:bg-blue-500',
          isDragging && 'bg-blue-500 dark:bg-blue-400',
        )}
      />
      
      {/* Расширенная область для захвата */}
      <div
        className={cn(
          'absolute inset-0 transition-all duration-200',
          direction === 'vertical' && '-left-2 -right-2',
        )}
      />
      
      {/* Индикатор перетаскивания */}
      <div
        className={cn(
          'absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2',
          'opacity-0 group-hover:opacity-100 transition-opacity duration-200',
          isDragging && 'opacity-100',
          'pointer-events-none'
        )}
      >
        <div className="flex items-center justify-center w-4 h-8 bg-white dark:bg-gray-800 rounded-sm border border-gray-300 dark:border-gray-600 shadow-sm">
          <div className="flex flex-col space-y-0.5">
            <div className="w-0.5 h-0.5 bg-gray-400 dark:bg-gray-500 rounded-full" />
            <div className="w-0.5 h-0.5 bg-gray-400 dark:bg-gray-500 rounded-full" />
            <div className="w-0.5 h-0.5 bg-gray-400 dark:bg-gray-500 rounded-full" />
            <div className="w-0.5 h-0.5 bg-gray-400 dark:bg-gray-500 rounded-full" />
            <div className="w-0.5 h-0.5 bg-gray-400 dark:bg-gray-500 rounded-full" />
          </div>
        </div>
      </div>
    </div>
  );
};
