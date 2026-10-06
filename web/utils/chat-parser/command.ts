import type { TFunction } from "i18next";
import { ChatCommand } from "@/types/chat-command";

export type ChatCommandResult =
    | { kind: "send"; text: string }
    | { kind: "help" }
    | {
          kind: "error";
          messageKey: string;
          params?: Record<string, string | number>;
      };

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
    id: string;
    name: string;
    description: string;
    usage: string;
    source: "builtin" | "user" | "channel";
};

const builtinId = (name: string) => `builtin:${name}`;

// a typed name that several commands share resolves in this order
const SCOPE_ORDER: ChatCommand["scope"][] = ["channel", "user"];

function orderedCustom(custom: ChatCommand[]): ChatCommand[] {
    return SCOPE_ORDER.flatMap((scope) =>
        custom.filter((c) => c.scope === scope),
    );
}

export function buildChatCommandIndex(
    custom: ChatCommand[],
    t: TFunction,
): ChatCommandSuggestion[] {
    return [
        ...BUILTIN_CHAT_COMMANDS.map((c): ChatCommandSuggestion => ({
            id: builtinId(c.name),
            name: c.name,
            description: t(c.descriptionKey),
            usage: c.usage,
            source: "builtin",
        })),
        ...orderedCustom(custom).map((c): ChatCommandSuggestion => ({
            id: c.id,
            name: c.name,
            description: c.description || c.response,
            usage: `/${c.name}`,
            source: c.scope,
        })),
    ];
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

export function chatCommandName(input: string): string {
    if (!input.startsWith("/")) return "";
    const space = input.indexOf(" ");
    return (space === -1 ? input.slice(1) : input.slice(1, space))
        .trim()
        .toLowerCase();
}

// pickedId is the suggestion the user chose; it decides between commands
// that share a name, otherwise the built-in > channel > user order applies
export function parseChatCommand(
    input: string,
    custom: ChatCommand[],
    t: TFunction,
    pickedId: string | null = null,
): ChatCommandResult | null {
    const name = chatCommandName(input);
    if (!name) return null;
    const space = input.indexOf(" ");
    const args = space === -1 ? "" : input.slice(space + 1);

    const builtin = BUILTIN_MAP.get(name);
    const customMatches = orderedCustom(custom).filter((c) => c.name === name);
    const picked = customMatches.find((c) => c.id === pickedId);

    if (picked) return { kind: "send", text: picked.response };
    if (builtin) return builtin.run(args, t);
    if (customMatches.length > 0) {
        return { kind: "send", text: customMatches[0].response };
    }

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
        for (const c of orderedCustom(custom)) {
            lines.push(
                `/${c.name} — ${c.description || c.response} (${t(`chat-commands:source.${c.scope}`)})`,
            );
        }
    }
    return lines.join("\n");
}
