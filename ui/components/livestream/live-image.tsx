'use client';

import { useEffect, useState } from 'react';
import Image, { ImageProps } from 'next/image';

interface LiveImageProps extends Omit<ImageProps, 'src'> {
  src: string; // only allow string
  fallbackSrc: string;
  refreshInterval?: number; // in ms
  alwaysRefresh?: boolean;
}

export default function LiveImage({
  src,
  fallbackSrc,
  refreshInterval = 5000,
  alwaysRefresh = true,
  ...props
}: LiveImageProps) {
  const [imgSrc, setImgSrc] = useState<string>(fallbackSrc);
  const [lastFailed, setLastFailed] = useState<boolean>(false);

  const tryLoadImage = () => {
    const testImg = new window.Image();
    testImg.src = `${src}?t=${Date.now()}`; // cache-busting to get new image (live refresh)

    testImg.onload = () => {
      setImgSrc(testImg.src);
      setLastFailed(false);
    };

    testImg.onerror = () => {
      setImgSrc(fallbackSrc);
      setLastFailed(true);
    };
  };

  useEffect(() => {
    tryLoadImage();

    const interval = setInterval(() => {
      if (alwaysRefresh || lastFailed) {
        tryLoadImage();
      }
    }, refreshInterval);

    return () => clearInterval(interval);
  }, []);

  return <Image {...props} src={imgSrc} unoptimized/>;
}