"use client";

export default function TypingIndicator({
    usernames,
}: {
    usernames: string[];
}) {
    if (usernames.length === 0) return null;

    let text: string;
    if (usernames.length === 1) {
        text = `${usernames[0]} is typing...`;
    } else if (usernames.length === 2) {
        text = `${usernames[0]} and ${usernames[1]} are typing...`;
    } else {
        text = `${usernames[0]} and ${usernames.length - 1} others are typing...`;
    }

    return (
        <div className="px-4 py-1">
            <p className="text-muted-foreground animate-pulse text-xs">
                {text}
            </p>
        </div>
    );
}
