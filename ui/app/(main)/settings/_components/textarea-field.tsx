import { Textarea } from "@/components/ui/textarea";
import { ComponentProps } from "react";
import Description from "./description";
import { cn } from "@/utils/cn";

type TextAreaProps = ComponentProps<typeof Textarea>;
type Props = {
  label: string;
  description?: string;
} & TextAreaProps;

export default function TextAreaField({
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
      <Textarea
        id={label}
        className={cn("text-foreground border border-border", className)}
        {...props}
      />
      {description && <Description content={description} className="mt-1" />}
    </div>
  );
}
