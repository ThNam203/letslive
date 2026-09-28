"use client";

import useT from "@/hooks/use-translation";

export default function TypingIndicator({
    usernames,
}: {
    usernames: string[];
}) {
    const { t } = useT("messages");

    if (usernames.length === 0) return null;

    let text: string;
    if (usernames.length === 1) {
        text = t("typing_one", { name: usernames[0] });
    } else if (usernames.length === 2) {
        text = t("typing_two", {
            name1: usernames[0],
            name2: usernames[1],
        });
    } else {
        text = t("typing_many", {
            name: usernames[0],
            count: usernames.length - 1,
        });
    }

    return (
        <div className="px-4 py-1">
            <p className="text-muted-foreground animate-pulse text-xs">
                {text}
            </p>
        </div>
    );
}
