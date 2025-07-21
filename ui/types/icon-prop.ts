export interface BaseIconProp {
    color?: string;
    width?: string;
    height?: string;
    viewBox?: string;
    className?: string;
}

export const defaultIconProps: BaseIconProp = {
    color: "#000000",
    width: "1.5rem",
    height: "1.5rem",
    viewBox: "0 0 24 24",
    className: "",
};
  
export type IconProp = React.SVGProps<SVGSVGElement> & BaseIconProp;