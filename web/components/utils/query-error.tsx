"use client";

import useT from "@/hooks/use-translation";
import { Button } from "@/components/ui/button";
import { cn } from "@/utils/cn";

interface QueryErrorProps {
    /** Re-runs the failed query. Omitted when there is nothing to retry. */
    onRetry?: () => void;
    /** Overrides the default "couldn't load this" message. */
    message?: string;
    className?: string;
}

/**
 * Shown where content would be if the request for it had succeeded.
 *
 * A failed request and an empty result look identical once the data is just
 * missing, so every screen that can render "nothing here" needs a way to say
 * "we could not find out" instead.
 */
export default function QueryError({
    onRetry,
    message,
    className,
}: QueryErrorProps) {
    const { t } = useT(["fetch-error", "common"]);

    return (
        <div
            role="alert"
            className={cn(
                "border-border bg-card flex flex-col items-center gap-3 rounded-lg border p-6 text-center",
                className,
            )}
        >
            <p className="text-muted-foreground text-sm">
                {message ?? t("fetch-error:load_failed")}
            </p>
            {onRetry && (
                <Button variant="outline" size="sm" onClick={onRetry}>
                    {t("common:retry")}
                </Button>
            )}
        </div>
    );
}
