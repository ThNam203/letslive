"use client";
import React from "react";
import styles from "./GlobalError.module.css";

type GlobalErrorProps = {
    error?: Error & { digest?: string };
    reset?: () => void;
    type: "404" | "500";
};

const GlobalErrorComponent = (props: GlobalErrorProps) => {
    return (
        <section className={styles.wrapper}>
            <div className={styles.container}>
                <div id="scene" className={styles.scene}>
                    <div className={styles.circle}></div>

                    <div className={styles.one}>
                        <div className={styles.content}>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                        </div>
                    </div>

                    <div className={styles.two}>
                        <div className={styles.content}>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                        </div>
                    </div>

                    <div className={styles.three}>
                        <div className={styles.content}>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                            <span className={styles.piece}></span>
                        </div>
                    </div>
                    <p className={styles.p404}>{props.type}</p>
                </div>

                <div className={styles.text}>
                    <article className={styles.article}>
                        <p>
                            {props.type === "500"
                                ? "Something went wrong!"
                                : "Page not found"}
                            <br />
                            {props.type === "500" && props.error?.digest
                                ? props.error.digest
                                : props.type === "500" && props.error?.message
                                ? props.error.message
                                : null}
                        </p>
                        {props.type === "500" && (
                            <button onClick={props.reset}>Try again</button>
                        )}
                        {props.type === "404" && (
                            <button onClick={() => window.history.back()}>
                                Go back
                            </button>
                        )}
                    </article>
                </div>
            </div>
        </section>
    );
};

export default GlobalErrorComponent;
