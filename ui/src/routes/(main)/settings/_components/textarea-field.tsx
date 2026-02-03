import { Textarea } from "@/components/ui/textarea";
import { ComponentProps } from "react";
import Description from "@/routes/(main)/settings/_components/description";
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
            <label className="mb-2 block text-sm font-medium" htmlFor={label}>
                {label}
            </label>
            <Textarea
                id={label}
                className={cn(
                    "border-border text-foreground border",
                    className,
                )}
                {...props}
            />
            {description && (
                <Description content={description} className="mt-1" />
            )}
        </div>
    );
}
