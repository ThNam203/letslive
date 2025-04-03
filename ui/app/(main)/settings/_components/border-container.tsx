import { cn } from "@/utils/cn";
import React, { ReactNode } from "react";

interface Props {
  children: ReactNode | ReactNode[];
  className?: string;
}
export default function BorderContainer({ children, className }: Props) {
  return (
    <div
      className={cn(
        "relative w-full rounded-lg border-1 border-gray-900 p-6",
        className
      )}
    >
      {children}
    </div>
  );
}
