import { cn } from "@/utils/cn";

type NotificationEmptyProps = {
    message: string;
    className?: string;
};

export function NotificationEmpty({
    message,
    className,
}: NotificationEmptyProps) {
    return (
        <div
            className={cn(
                "text-muted-foreground flex items-center justify-center py-8 text-sm",
                className,
            )}
        >
            {message}
        </div>
    );
}
