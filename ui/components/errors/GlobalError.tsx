import React from 'react';
import styles from './GlobalError.module.css'; // Adjust the path as necessary

type GlobalErrorProps = {
    error: Error & { digest?: string },
    reset: () => void
}

const GlobalError = (props: GlobalErrorProps) => {
    return (
        <section className={styles.wrapper}>
            <div className={styles.container}>
                <div id="scene" className={styles.scene} data-hover-only="false">
                    <div className={styles.circle} data-depth="1.2"></div>

                    <div className={styles.one} data-depth="0.9">
                        <div className={styles.content}>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                        </div>
                    </div>

                    <div className={styles.two} data-depth="0.60">
                        <div className={styles.content}>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                        </div>
                    </div>

                    <div className={styles.three} data-depth="0.40">
                        <div className={styles.content}>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                        </div>
                    </div>

                    <p className={styles.p404} data-depth="0.50">500</p>
                    <p className={styles.p404} data-depth="0.10">500</p>
                </div>

                <div className={styles.text}>
                    <article className={styles.article}>
                        <p>
                            Something went wrong!<br />
                            {props.error.digest ? props.error.digest : props.error.message}
                        </p>
                        <button onClick={props.reset}>Try again</button>
                    </article>
                </div>
            </div>
        </section>
    );
};

export default GlobalError;
