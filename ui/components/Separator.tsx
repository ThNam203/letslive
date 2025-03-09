
import { ClassValue } from "clsx";
import { cn } from "../utils/cn";

const Separator = ({
  color = "bg-gray-200",
  classname,
}: {
  classname?: ClassValue;
  color?: string;
  height?: string;
}) => {
  return <div className={cn("h-[0.5px] w-full", color, classname)}></div>;
};

export default Separator;