
import { ClassValue } from "clsx";
import { cn } from "../utils/cn";
import AllChannelsView from "./leftbar/AllChannelsView";

const View = ({
  className,
  viewers,
}: {
  className?: ClassValue;
  viewers: number;
}) => {
  return (
    <div
      className={cn(
        "flex flex-row items-center justify-center gap-2",
        className
      )}
    >
      <span className="w-2 h-2 rounded-full bg-red-600"></span>
      <span className="text-xs">{viewers.toString()}</span>
    </div>
  );
};

export const LeftBar = () => {
    
  return (
    <div className="fixed ml-2 h-full xl:w-64 max-xl:w-fit bg-leftBarColor flex flex-col max-xl:items-center justify-start gap-2 text-primaryWord border-r-1">
      <div className="flex flex-row justify-between items-center px-2 pt-4">
        <AllChannelsView/>
      </div>
    </div>
  );
};