"use client";

import useT from "@/hooks/use-translation";
import useCompactNumber from "@/hooks/use-compact-number";

/** "1.2K followers", pluralised and formatted for the active language. */
export default function FollowerCount({
    count,
    className,
}: {
    count: number;
    className?: string;
}) {
    const { t } = useT("common");
    const formatCompact = useCompactNumber();

    return (
        <span className={className}>
            {t("common:followers_count", {
                count,
                formatted: formatCompact(count),
            })}
        </span>
    );
}
