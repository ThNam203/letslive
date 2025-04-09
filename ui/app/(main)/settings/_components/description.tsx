import { cn } from "@/utils/cn";
import React from "react";

interface Props {
  content: string;
  className?: string;
}
export default function Description({ content, className }: Props) {
  return (
    <p className={cn("text-sm text-slate-500 p-0 m-0", className)}>{content}</p>
  );
}
