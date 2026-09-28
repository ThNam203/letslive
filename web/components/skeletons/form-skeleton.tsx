import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/utils/cn";

export function FieldSkeleton({ multiline = false }: { multiline?: boolean }) {
    return (
        <div className="space-y-2">
            <Skeleton className="h-4 w-32" />
            <Skeleton className={cn("w-full", multiline ? "h-24" : "h-9")} />
        </div>
    );
}

/** Labelled inputs followed by a submit button. */
export function FormSkeleton({
    fields = 3,
    multiline = [],
    className,
}: {
    fields?: number;
    /** indexes of fields that render as a textarea */
    multiline?: number[];
    className?: string;
}) {
    return (
        <div className={cn("space-y-6", className)} aria-busy="true">
            {Array.from({ length: fields }, (_, i) => (
                <FieldSkeleton key={i} multiline={multiline.includes(i)} />
            ))}
            <Skeleton className="h-9 w-28" />
        </div>
    );
}
