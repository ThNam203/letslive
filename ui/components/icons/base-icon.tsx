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

    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width={width}
            height={height}
            className={className}
            style={{ color }}
            viewBox={viewBox}
            {...rest}
        >
            {children}
        </svg>
    );
}
