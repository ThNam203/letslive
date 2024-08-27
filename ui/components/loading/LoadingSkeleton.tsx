import React from 'react';
import styles from './LoadingSkeleton.module.css';

const LoadingSkeleton = () => {
  return (
    <div className={styles.scene}>
      <div className={styles.cubeWrapper}>
        <div className={styles.cube}>
          <div className={styles.cubeFaces}>
            <div className={`${styles.cubeFace} ${styles.shadow}`}></div>
            <div className={`${styles.cubeFace} ${styles.bottom}`}></div>
            <div className={`${styles.cubeFace} ${styles.top}`}></div>
            <div className={`${styles.cubeFace} ${styles.left}`}></div>
            <div className={`${styles.cubeFace} ${styles.right}`}></div>
            <div className={`${styles.cubeFace} ${styles.back}`}></div>
            <div className={`${styles.cubeFace} ${styles.front}`}></div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LoadingSkeleton;
