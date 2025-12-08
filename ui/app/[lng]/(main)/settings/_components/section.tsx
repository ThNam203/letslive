import React, { ReactNode } from "react";
import BorderContainer from "./border-container";
import Description from "./description";

interface Props {
    title: string;
    description?: string;
    children: ReactNode | ReactNode[];
    className?: string;
    contentClassName?: string;
    hasBorder?: boolean;
}
export default function Section({
    title,
    description,
    children,
    className,
    contentClassName,
}: Props) {
    return (
        <section className={className}>
            <div className="mb-4">
                <h2 className="text-xl font-semibold text-foreground">
                    {title}
                </h2>
                {description && <Description content={description} />}
            </div>
            <BorderContainer className={contentClassName}>
                {children}
            </BorderContainer>
        </section>
    );
}
