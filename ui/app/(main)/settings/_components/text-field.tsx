import { Input } from "@/components/ui/input";
import React, { ComponentProps } from "react";
import Description from "./description";
import { cn } from "@/utils/cn";

type TextProps = ComponentProps<typeof Input>;
type Props = {
  label: string;
  description?: string;
} & TextProps;

export default function TextField({
  label,
  description,
  className,
  ...props
}: Props) {
  return (
    <div>
      <label className="block text-sm font-medium mb-2" htmlFor={label}>
        {label}
      </label>
      <Input
        id={label}
        className={cn("text-foreground border border-border focus:outline-none", className)}
        {...props}
      />
      {description && <Description content={description} className="mt-1" />}
    </div>
  );
}
