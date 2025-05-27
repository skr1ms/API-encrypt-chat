import React from 'react';
import { cn } from '@/lib/utils';
import { ResizeHandle } from './resize-handle';
import { useResizablePanels, PanelSizes } from '@/hooks/use-resizable-panels';

interface ResizablePanelsProps {
  children: {
    sidebar: React.ReactNode;
    main: React.ReactNode;
    profile?: React.ReactNode;
  };
  className?: string;
  onPanelSizeChange?: (sizes: PanelSizes) => void;
  showProfile?: boolean;
  onToggleProfile?: (visible: boolean) => void;
}

export const ResizablePanels: React.FC<ResizablePanelsProps> = ({
  children,
  className,
  onPanelSizeChange,
  showProfile: externalShowProfile,
  onToggleProfile,
}) => {
  const {
    panelSizes,
    isDragging,
    showProfile: internalShowProfile,
    handleMouseDown,
    toggleProfile,
  } = useResizablePanels();

  // Используем внешнее управление видимостью профиля, если оно предоставлено
  const showProfile = externalShowProfile !== undefined ? externalShowProfile : internalShowProfile;

  // Если предоставлен внешний обработчик, используем его
  const handleToggleProfile = (visible: boolean) => {
    if (onToggleProfile) {
      onToggleProfile(visible);
    } else {
      toggleProfile(visible);
    }
  };

  // Вызываем callback при изменении размеров
  React.useEffect(() => {
    if (onPanelSizeChange) {
      onPanelSizeChange(panelSizes);
    }
  }, [panelSizes, onPanelSizeChange]);

  return (
    <div className={cn('h-screen bg-gray-50 dark:bg-gray-900 flex overflow-hidden', className)}>
      {/* Боковая панель (sidebar) */}
      <div
        className="bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex-shrink-0 relative"
        style={{ width: `${panelSizes.sidebar}px` }}
      >
        {children.sidebar}
        
        {/* Ручка изменения размера для sidebar */}
        <div className="absolute top-0 right-0 h-full">
          <ResizeHandle
            direction="vertical"
            position="right"
            onMouseDown={handleMouseDown('sidebar')}
            isDragging={isDragging === 'sidebar'}
          />
        </div>
      </div>

      {/* Основная область */}
      <div className="flex-1 flex flex-col min-w-0">
        {children.main}
      </div>

      {/* Панель профиля */}
      {showProfile && children.profile && (
        <div
          className="bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 flex-shrink-0 relative"
          style={{ width: `${panelSizes.profile}px` }}
        >
          {/* Ручка изменения размера для profile */}
          <div className="absolute top-0 left-0 h-full">
            <ResizeHandle
              direction="vertical"
              position="left"
              onMouseDown={handleMouseDown('profile')}
              isDragging={isDragging === 'profile'}
            />
          </div>
          
          {children.profile}
        </div>
      )}

      {/* Overlay для предотвращения выбора текста во время перетаскивания */}
      {isDragging && (
        <div className="fixed inset-0 z-50 cursor-col-resize select-none bg-transparent" />
      )}
    </div>
  );
};
