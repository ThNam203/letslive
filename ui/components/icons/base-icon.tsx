import { defaultIconProps, IconProp } from "@/types/icon-prop";

type BaseIconComponentProps = IconProp & { children: React.ReactNode };

export function BaseIcon({
    color = defaultIconProps.color,
    width = defaultIconProps.width,
    height = defaultIconProps.height,
    className = defaultIconProps.className,
    viewBox = defaultIconProps.viewBox,
    children,
    ...props
}: BaseIconComponentProps) {
    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width={width}
            height={height}
            className={className}
            fill={color}
            viewBox={viewBox}
            {...props}
        >
            {children}
        </svg>
    );
}
