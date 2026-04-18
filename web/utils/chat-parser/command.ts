import type { TFunction } from "next-i18next";
import { ChatCommand } from "@/types/chat-command";

export type ChatCommandResult =
    | { kind: "send"; text: string }
    | { kind: "help" }
    | {
          kind: "error";
          messageKey: string;
          params?: Record<string, string | number>;
      }
    | { kind: "noop" };

export type BuiltinChatCommand = {
    name: string;
    usage: string;
    descriptionKey: string;
    run: (args: string, t: TFunction) => ChatCommandResult;
};

const SHRUG = "¯\\_(ツ)_/¯";
const TABLEFLIP = "(╯°□°)╯︵ ┻━┻";
const UNFLIP = "┬─┬ ノ( ゜-゜ノ)";

export const BUILTIN_CHAT_COMMANDS: BuiltinChatCommand[] = [
    {
        name: "me",
        usage: "/me <text>",
        descriptionKey: "chat-commands:builtin.me",
        run: (args) => {
            const trimmed = args.trim();
            if (!trimmed) return { kind: "noop" };
            return { kind: "send", text: `_${trimmed}_` };
        },
    },
    {
        name: "shrug",
        usage: "/shrug [text]",
        descriptionKey: "chat-commands:builtin.shrug",
        run: (args) => {
            const trimmed = args.trim();
            return {
                kind: "send",
                text: trimmed ? `${trimmed} ${SHRUG}` : SHRUG,
            };
        },
    },
    {
        name: "tableflip",
        usage: "/tableflip [text]",
        descriptionKey: "chat-commands:builtin.tableflip",
        run: (args) => {
            const trimmed = args.trim();
            return {
                kind: "send",
                text: trimmed ? `${trimmed} ${TABLEFLIP}` : TABLEFLIP,
            };
        },
    },
    {
        name: "unflip",
        usage: "/unflip [text]",
        descriptionKey: "chat-commands:builtin.unflip",
        run: (args) => {
            const trimmed = args.trim();
            return {
                kind: "send",
                text: trimmed ? `${trimmed} ${UNFLIP}` : UNFLIP,
            };
        },
    },
    {
        name: "roll",
        usage: "/roll [max]",
        descriptionKey: "chat-commands:builtin.roll",
        run: (args, t) => {
            const trimmed = args.trim();
            let max = 100;
            if (trimmed) {
                const n = parseInt(trimmed, 10);
                if (!Number.isFinite(n) || n < 1 || n > 1_000_000) {
                    return {
                        kind: "error",
                        messageKey: "chat-commands:errors.roll_usage",
                    };
                }
                max = n;
            }
            const value = Math.floor(Math.random() * max) + 1;
            return {
                kind: "send",
                text: t("chat-commands:roll_message", { value, max }),
            };
        },
    },
    {
        name: "help",
        usage: "/help",
        descriptionKey: "chat-commands:builtin.help",
        run: () => ({ kind: "help" }),
    },
];

const BUILTIN_MAP = new Map(BUILTIN_CHAT_COMMANDS.map((c) => [c.name, c]));

export type ChatCommandSuggestion = {
    name: string;
    description: string;
    usage: string;
    source: "builtin" | "user" | "channel";
};

export function buildChatCommandIndex(
    custom: ChatCommand[],
    t: TFunction,
): ChatCommandSuggestion[] {
    const seen = new Set<string>();
    const out: ChatCommandSuggestion[] = [];

    for (const c of BUILTIN_CHAT_COMMANDS) {
        if (seen.has(c.name)) continue;
        seen.add(c.name);
        out.push({
            name: c.name,
            description: t(c.descriptionKey),
            usage: c.usage,
            source: "builtin",
        });
    }
    for (const c of custom) {
        if (seen.has(c.name)) continue;
        seen.add(c.name);
        out.push({
            name: c.name,
            description: c.description || c.response,
            usage: `/${c.name}`,
            source: c.scope,
        });
    }
    return out;
}

export function filterChatCommandSuggestions(
    index: ChatCommandSuggestion[],
    input: string,
): ChatCommandSuggestion[] {
    if (!input.startsWith("/")) return [];
    const after = input.slice(1);
    if (after.includes(" ")) return [];
    const q = after.toLowerCase();
    return index.filter((c) => c.name.startsWith(q)).slice(0, 8);
}

export function parseChatCommand(
    input: string,
    custom: ChatCommand[],
    t: TFunction,
): ChatCommandResult | null {
    if (!input.startsWith("/")) return null;
    const space = input.indexOf(" ");
    const name = (space === -1 ? input.slice(1) : input.slice(1, space))
        .trim()
        .toLowerCase();
    const args = space === -1 ? "" : input.slice(space + 1);
    if (!name) return null;

    const builtin = BUILTIN_MAP.get(name);
    if (builtin) return builtin.run(args, t);

    const customMatch = custom.find((c) => c.name === name);
    if (customMatch) return { kind: "send", text: customMatch.response };

    return {
        kind: "error",
        messageKey: "chat-commands:errors.unknown",
        params: { name },
    };
}

export function buildChatCommandHelpText(
    custom: ChatCommand[],
    t: TFunction,
): string {
    const lines = [
        t("chat-commands:help.header"),
        ...BUILTIN_CHAT_COMMANDS.map(
            (c) => `${c.usage} — ${t(c.descriptionKey)}`,
        ),
    ];
    if (custom.length > 0) {
        lines.push(t("chat-commands:help.custom_header"));
        for (const c of custom) {
            lines.push(`/${c.name} — ${c.description || c.response}`);
        }
    }
    return lines.join("\n");
}
