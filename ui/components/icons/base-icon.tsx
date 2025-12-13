"use client";

import { getDefaultIconProps, IconProp } from "@/types/icon-prop";
import { useTheme } from "next-themes";

type BaseIconComponentProps = IconProp & { children: React.ReactNode };

export function BaseIcon(rawProps: BaseIconComponentProps) {
    const { resolvedTheme } = useTheme();

    if (!resolvedTheme) return null;

    const props = { ...getDefaultIconProps(resolvedTheme), ...rawProps };
    const { color, width, height, className, viewBox, children, ...rest } =
        props;

    // apply inline color style if color prop is explicitly provided
    // this allows Tailwind classes to work
    const style = rawProps.color !== undefined ? { color } : undefined;

    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width={width}
            height={height}
            style={style}
            className={className}
            viewBox={viewBox}
            {...rest}
        >
            {children}
        </svg>
    );
}
