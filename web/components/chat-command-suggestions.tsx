"use client";

import { ChatCommandSuggestion } from "@/utils/chat-parser";
import useT from "@/hooks/use-translation";

type Props = {
    suggestions: ChatCommandSuggestion[];
    activeIndex: number;
    onPick: (s: ChatCommandSuggestion) => void;
};

export default function ChatCommandSuggestions({
    suggestions,
    activeIndex,
    onPick,
}: Props) {
    const { t } = useT("chat-commands");
    if (suggestions.length === 0) return null;

    return (
        <div className="border-border bg-background absolute right-0 bottom-12 left-0 z-20 max-h-64 overflow-y-auto rounded-md border shadow-md">
            <ul className="text-sm">
                {suggestions.map((s, i) => (
                    <li
                        key={s.name}
                        onMouseDown={(e) => {
                            e.preventDefault();
                            onPick(s);
                        }}
                        className={`flex cursor-pointer items-center justify-between px-3 py-2 ${
                            i === activeIndex ? "bg-muted" : "hover:bg-muted/60"
                        }`}
                    >
                        <div className="min-w-0">
                            <div className="font-mono font-medium">
                                {s.usage}
                            </div>
                            <div className="text-muted-foreground truncate text-xs">
                                {s.description}
                            </div>
                        </div>
                        <span className="text-muted-foreground ml-2 text-[10px] uppercase">
                            {t(`chat-commands:source.${s.source}`)}
                        </span>
                    </li>
                ))}
            </ul>
        </div>
    );
}
