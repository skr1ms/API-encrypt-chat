import { useState, useCallback, useEffect } from 'react';

export interface PanelSizes {
  sidebar: number;
  main: number;
  profile: number;
}

export interface ResizablePanelsConfig {
  minSidebar: number;
  maxSidebar: number;
  minMain: number;
  minProfile: number;
  maxProfile: number;
  defaultSizes: PanelSizes;
}

const DEFAULT_CONFIG: ResizablePanelsConfig = {
  minSidebar: 180,
  maxSidebar: 450,
  minMain: 300,
  minProfile: 250,
  maxProfile: 450,
  defaultSizes: {
    sidebar: 320,
    main: 0, // будет вычислено динамически
    profile: 320,
  },
};

export const useResizablePanels = (config: Partial<ResizablePanelsConfig> = {}) => {
  const finalConfig = { ...DEFAULT_CONFIG, ...config };
  
  // Загружаем сохраненные размеры из localStorage
  const loadSavedSizes = (): PanelSizes => {
    try {
      const saved = localStorage.getItem('messenger-panel-sizes');
      if (saved) {
        const parsed = JSON.parse(saved);
        return {
          sidebar: Math.max(finalConfig.minSidebar, Math.min(finalConfig.maxSidebar, parsed.sidebar || finalConfig.defaultSizes.sidebar)),
          main: parsed.main || 0,
          profile: Math.max(finalConfig.minProfile, Math.min(finalConfig.maxProfile, parsed.profile || finalConfig.defaultSizes.profile)),
        };
      }
    } catch (error) {
      console.error('Error loading panel sizes from localStorage:', error);
    }
    return finalConfig.defaultSizes;
  };

  const [panelSizes, setPanelSizes] = useState<PanelSizes>(loadSavedSizes);
  const [isDragging, setIsDragging] = useState<'sidebar' | 'profile' | null>(null);
  const [showProfile, setShowProfile] = useState(true);

  // Сохраняем размеры в localStorage
  const saveSizes = useCallback((sizes: PanelSizes) => {
    try {
      localStorage.setItem('messenger-panel-sizes', JSON.stringify(sizes));
    } catch (error) {
      console.error('Error saving panel sizes to localStorage:', error);
    }
  }, []);

  // Обновляем размеры панелей
  const updatePanelSizes = useCallback((newSizes: Partial<PanelSizes>) => {
    setPanelSizes(prev => {
      const updated = { ...prev, ...newSizes };
      saveSizes(updated);
      return updated;
    });
  }, [saveSizes]);

  // Обработчик начала перетаскивания
  const handleMouseDown = useCallback((panelType: 'sidebar' | 'profile') => (e: React.MouseEvent) => {
    e.preventDefault();
    setIsDragging(panelType);
  }, []);

  // Обработчик перетаскивания
  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isDragging) return;

    const containerWidth = window.innerWidth;
    
    if (isDragging === 'sidebar') {
      const newSidebarWidth = Math.max(
        finalConfig.minSidebar,
        Math.min(finalConfig.maxSidebar, e.clientX)
      );
      
      const remainingWidth = containerWidth - newSidebarWidth - (showProfile ? panelSizes.profile : 0);
      
      if (remainingWidth >= finalConfig.minMain) {
        updatePanelSizes({
          sidebar: newSidebarWidth,
          main: remainingWidth,
        });
      }
    } else if (isDragging === 'profile') {
      const newProfileWidth = Math.max(
        finalConfig.minProfile,
        Math.min(finalConfig.maxProfile, containerWidth - e.clientX)
      );
      
      const remainingWidth = containerWidth - panelSizes.sidebar - newProfileWidth;
      
      if (remainingWidth >= finalConfig.minMain) {
        updatePanelSizes({
          profile: newProfileWidth,
          main: remainingWidth,
        });
      }
    }
  }, [isDragging, showProfile, panelSizes.sidebar, panelSizes.profile, finalConfig, updatePanelSizes]);

  // Обработчик завершения перетаскивания
  const handleMouseUp = useCallback(() => {
    setIsDragging(null);
  }, []);

  // Добавляем и удаляем обработчики событий
  useEffect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
      
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
      };
    }
  }, [isDragging, handleMouseMove, handleMouseUp]);

  // Пересчитываем размеры при изменении видимости профиля
  useEffect(() => {
    const containerWidth = window.innerWidth;
    const mainWidth = containerWidth - panelSizes.sidebar - (showProfile ? panelSizes.profile : 0);
    
    updatePanelSizes({
      main: Math.max(finalConfig.minMain, mainWidth),
    });
  }, [showProfile, panelSizes.sidebar, panelSizes.profile, finalConfig.minMain, updatePanelSizes]);

  // Функция для сброса размеров к значениям по умолчанию
  const resetSizes = useCallback(() => {
    const containerWidth = window.innerWidth;
    const mainWidth = containerWidth - finalConfig.defaultSizes.sidebar - (showProfile ? finalConfig.defaultSizes.profile : 0);
    
    const newSizes = {
      ...finalConfig.defaultSizes,
      main: Math.max(finalConfig.minMain, mainWidth),
    };
    
    setPanelSizes(newSizes);
    saveSizes(newSizes);
  }, [finalConfig, showProfile, saveSizes]);

  // Функция для переключения видимости профиля
  const toggleProfile = useCallback((visible?: boolean) => {
    setShowProfile(prev => visible !== undefined ? visible : !prev);
  }, []);

  return {
    panelSizes,
    isDragging,
    showProfile,
    handleMouseDown,
    resetSizes,
    toggleProfile,
    updatePanelSizes,
    config: finalConfig,
  };
};
