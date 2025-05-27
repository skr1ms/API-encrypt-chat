import { useState, useEffect, useRef } from 'react';

export const usePanelWidth = () => {
  const [width, setWidth] = useState<number>(0);
  const [isNarrow, setIsNarrow] = useState<boolean>(false);
  const [isVeryNarrow, setIsVeryNarrow] = useState<boolean>(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const updateWidth = () => {
      if (ref.current) {
        const newWidth = ref.current.offsetWidth;
        setWidth(newWidth);
        setIsNarrow(newWidth < 250);
        setIsVeryNarrow(newWidth < 200);
      }
    };

    const resizeObserver = new ResizeObserver(() => {
      updateWidth();
    });

    if (ref.current) {
      resizeObserver.observe(ref.current);
      updateWidth(); // Initial measurement
    }

    return () => {
      resizeObserver.disconnect();
    };
  }, []);

  return {
    ref,
    width,
    isNarrow,
    isVeryNarrow
  };
};
