export interface BaseIconProp {
    color?: string;
    width?: string;
    height?: string;
    viewBox?: string;
    className?: string;
}

export const getDefaultIconProps = (resolvedTheme: string): BaseIconProp => {
    return {
        color: resolvedTheme === "light" ? "#000000" : "#FFFFFF",
        width: "1.5rem",
        height: "1.5rem",
        viewBox: "0 0 24 24",
        className: ""
    }
}
  
export type IconProp = React.SVGProps<SVGSVGElement> & BaseIconProp;