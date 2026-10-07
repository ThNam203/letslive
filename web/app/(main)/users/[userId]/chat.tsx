"use client";

import type React from "react";
import { useState, useRef, useEffect, useMemo, useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "@/components/utils/toast";
import useUser from "@/hooks/user";
import { ChatMemberEvent, ReceivedMessage } from "@/types/message";
import { useRealtime } from "@/contexts/realtime-context";
import { REALTIME_EVENT, realtimeRoomTopic } from "@/constant/realtime";
import { SendChatMessage } from "@/lib/api/chat";
import { usePublicUser } from "@/hooks/queries/use-users";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import IconClose from "@/components/icons/close";
import IconSend from "@/components/icons/send";
import EmotePicker from "@/components/emote-picker";
import ChatCommandSuggestions from "@/components/chat-command-suggestions";
import { uuidToReadableHexColor } from "@/utils/uuid-color";
import {
    buildChatCommandHelpText,
    buildChatCommandIndex,
    ChatCommandSuggestion,
    chatCommandName,
    filterChatCommandSuggestions,
    parseChatCommand,
    parseEmotes,
} from "@/utils/chat-parser";
import useT from "@/hooks/use-translation";
import { CHAT_MESSAGE_MAX_LENGTH } from "@/constant/field-limits";
import {
    roomMessagesQueryKey,
    useRoomChatCommands,
    useRoomMessages,
} from "@/hooks/queries/use-chat";
import { formatLocaleDate } from "@/utils/timeFormats";
import UserAvatar from "@/components/ui/user-avatar";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";

type LocalMessage = {
    kind: "system";
    text: string;
};

type MemberLine = {
    userId: string;
    joined: boolean;
    timestamp: number;
};

type ChatLine =
    | { kind: "remote"; data: ReceivedMessage }
    | { kind: "member"; data: MemberLine }
    | { kind: "local"; data: LocalMessage };

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === "object" && value !== null;
}

function isReceivedMessage(value: unknown): value is ReceivedMessage {
    return (
        isRecord(value) &&
        typeof value.userId === "string" &&
        typeof value.username === "string" &&
        typeof value.text === "string"
    );
}

function isMemberEvent(value: unknown): value is ChatMemberEvent {
    return isRecord(value) && typeof value.userId === "string";
}

function ChatMessageRow({ message }: { message: ReceivedMessage }) {
    const { t, i18n } = useT("chat");
    const sentAt = new Date(message.timestamp);

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <div className="mb-3">
                    <UserAvatar
                        src={message.profilePicture}
                        name={message.username}
                        size="sm"
                        className="mr-2 inline-flex h-6 w-6 align-middle"
                        fallbackClassName="text-xs"
                    />
                    <span
                        style={{
                            color: `${uuidToReadableHexColor(message.userId)}`,
                        }}
                        className="mr-2 font-semibold"
                    >
                        {message.username}:
                    </span>
                    <span className="text-foreground">
                        {parseEmotes(message.text)}
                    </span>
                </div>
            </TooltipTrigger>
            <TooltipContent side="left">
                <time dateTime={sentAt.toISOString()}>
                    {formatLocaleDate(sentAt, i18n.resolvedLanguage, {
                        year: "numeric",
                        month: "short",
                        day: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                    })}
                </time>
            </TooltipContent>
        </Tooltip>
    );
}

// Join/leave events carry only the user id; the name and avatar come from the
// cached public profile.
function MemberEventRow({ member }: { member: MemberLine }) {
    const { t } = useT("chat");
    const { data: profile } = usePublicUser(member.userId);
    if (!profile) return null;

    return (
        <ChatMessageRow
            message={{
                userId: member.userId,
                username: profile.username,
                profilePicture: profile.profilePicture ?? null,
                text: member.joined ? t("chat:joined") : t("chat:left"),
                timestamp: member.timestamp,
            }}
        />
    );
}

export default function ChatPanel({
    roomId,
    onClose,
}: {
    roomId: string;
    onClose: () => void;
}) {
    const activeEmotePattern = /(^|\s):([a-z0-9_]*)$/i;
    const user = useUser((state) => state.user);
    // Lines that arrived over the socket, or were produced locally by a
    // command. The backlog lives in the query cache and is prepended below.
    const [liveLines, setLiveLines] = useState<ChatLine[]>([]);
    const [inputMessage, setInputMessage] = useState("");
    const [atBottom, setAtBottom] = useState(true);
    const messageContainerRef = useRef<HTMLDivElement | null>(null);
    const { t } = useT(["chat", "chat-commands"]);
    const { t: tApi } = useT("api-response");
    const { subscribe, onEvent, onReconnect } = useRealtime();
    const queryClient = useQueryClient();
    const { data: backlog } = useRoomMessages(roomId);
    const { data: customChatCommands = [] } = useRoomChatCommands(roomId);

    const messages: ChatLine[] = useMemo(
        () => [
            ...(backlog ?? []).map((m): ChatLine => ({
                kind: "remote",
                data: m,
            })),
            ...liveLines,
        ],
        [backlog, liveLines],
    );
    const [suggestions, setSuggestions] = useState<ChatCommandSuggestion[]>([]);
    const [activeSuggestion, setActiveSuggestion] = useState(0);
    const [pickedCommand, setPickedCommand] =
        useState<ChatCommandSuggestion | null>(null);
    const [emotePickerOpen, setEmotePickerOpen] = useState(false);
    const [emoteSearch, setEmoteSearch] = useState("");

    const chatCommandIndex = useMemo(
        () => buildChatCommandIndex(customChatCommands, t),
        [customChatCommands, t],
    );

    const appendLine = useCallback(
        (line: ChatLine) =>
            setLiveLines((prev) =>
                prev.length >= 100 ? [...prev.slice(1), line] : [...prev, line],
            ),
        [],
    );

    // the line shows up when its chat.message push arrives, so only
    // failures are handled here
    const sendText = (text: string) => {
        if (!user) {
            toast(t("chat:login_required"), { type: "error" });
            return;
        }
        SendChatMessage(roomId, text)
            .then((res) => {
                if (!res.success) {
                    toast(tApi(res.key) || res.message, { type: "error" });
                }
            })
            .catch(() => {
                toast(tApi("fetch-error:client_fetch_error"), {
                    type: "error",
                });
            });
    };

    const handleSendMessage = (e: React.FormEvent) => {
        e.preventDefault();
        const raw = inputMessage.trim();
        if (!raw || !user) return;

        if (raw.startsWith("/")) {
            const result = parseChatCommand(
                raw,
                customChatCommands,
                t,
                pickedCommand?.id,
            );
            setInputMessage("");
            setSuggestions([]);
            setPickedCommand(null);
            if (!result) return;
            if (result.kind === "error") {
                appendLine({
                    kind: "local",
                    data: {
                        kind: "system",
                        text: t(result.messageKey, result.params),
                    },
                });
                return;
            }
            if (result.kind === "help") {
                appendLine({
                    kind: "local",
                    data: {
                        kind: "system",
                        text: buildChatCommandHelpText(customChatCommands, t),
                    },
                });
                return;
            }
            sendText(result.text.slice(0, CHAT_MESSAGE_MAX_LENGTH));
            return;
        }

        setInputMessage("");
        setSuggestions([]);
        sendText(raw);
    };

    const applySuggestion = (s: ChatCommandSuggestion) => {
        setInputMessage(`/${s.name} `);
        setPickedCommand(s);
        setSuggestions([]);
        setActiveSuggestion(0);
    };

    const handleInputChange = (value: string) => {
        setInputMessage(value);
        if (pickedCommand && chatCommandName(value) !== pickedCommand.name) {
            setPickedCommand(null);
        }
        const next = filterChatCommandSuggestions(chatCommandIndex, value);
        setSuggestions(next);
        setActiveSuggestion(0);

        const emoteMatch = value.match(activeEmotePattern);
        if (emoteMatch) {
            setEmoteSearch(emoteMatch[2] ?? "");
            setEmotePickerOpen(true);
        } else {
            setEmotePickerOpen(false);
            setEmoteSearch("");
        }
    };

    const handleEmoteSelect = (shortcode: string) => {
        const emoteCode = shortcode.slice(1, -1);
        setInputMessage((prev) => {
            const next = prev.match(activeEmotePattern)
                ? prev.replace(activeEmotePattern, (_, prefix) => {
                      return `${prefix}:${emoteCode}: `;
                  })
                : `${prev}${shortcode}`;
            const nextSuggestions = filterChatCommandSuggestions(
                chatCommandIndex,
                next,
            );
            setSuggestions(nextSuggestions);
            setActiveSuggestion(0);
            return next;
        });
        setEmotePickerOpen(false);
        setEmoteSearch("");
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (suggestions.length === 0) return;
        if (e.key === "ArrowDown") {
            e.preventDefault();
            setActiveSuggestion((i) => (i + 1) % suggestions.length);
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            setActiveSuggestion(
                (i) => (i - 1 + suggestions.length) % suggestions.length,
            );
        } else if (e.key === "Tab") {
            e.preventDefault();
            applySuggestion(suggestions[activeSuggestion]);
        } else if (e.key === "Escape") {
            setSuggestions([]);
        }
    };

    useEffect(() => {
        const container = messageContainerRef.current;
        if (!container) return;

        const handleScroll = () => {
            const distanceFromBottom =
                container.scrollHeight -
                container.scrollTop -
                container.clientHeight;
            setAtBottom(distanceFromBottom < 3); // 3px tolerance
        };

        container.addEventListener("scroll", handleScroll);
        return () => container.removeEventListener("scroll", handleScroll);
    }, []);

    useEffect(() => {
        if (atBottom) {
            const container = messageContainerRef.current;
            if (container) {
                container.scrollTop = container.scrollHeight;
            }
        }
    }, [messages, atBottom]);

    // Viewers, signed in or not, receive the room's lines over the shared
    // realtime socket; join/leave come from the gateway's membership events.
    useEffect(() => {
        const topic = realtimeRoomTopic(roomId);
        const memberHandler = (joined: boolean) =>
            onEvent(
                joined
                    ? REALTIME_EVENT.MEMBER_JOINED
                    : REALTIME_EVENT.MEMBER_LEFT,
                (frame) => {
                    if (frame.topic !== topic || !isMemberEvent(frame.data))
                        return;
                    appendLine({
                        kind: "member",
                        data: {
                            userId: frame.data.userId,
                            joined,
                            timestamp: Date.now(),
                        },
                    });
                },
            );

        const cleanups = [
            subscribe(topic),
            onEvent(REALTIME_EVENT.CHAT_MESSAGE, (frame) => {
                if (frame.topic !== topic || !isReceivedMessage(frame.data))
                    return;
                appendLine({ kind: "remote", data: frame.data });
            }),
            memberHandler(true),
            memberHandler(false),
            // lines sent while the socket was down are only in the backlog
            onReconnect(() => {
                setLiveLines([]);
                queryClient.invalidateQueries({
                    queryKey: roomMessagesQueryKey(roomId),
                });
            }),
        ];
        return () => cleanups.forEach((cleanup) => cleanup());
    }, [roomId, subscribe, onEvent, onReconnect, queryClient, appendLine]);

    return (
        <div className="relative flex h-full w-full flex-col">
            <div className="border-border flex items-center justify-between border border-y-0 px-4 py-3">
                <h2 className="font-semibold">{t("chat:title")}</h2>
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={onClose}
                    className="md:hidden"
                >
                    <IconClose className="h-4 w-4" />
                </Button>
            </div>
            <div
                ref={messageContainerRef}
                className="border-border mb-18 flex-1 overflow-y-auto rounded-md rounded-t-none border border-t-0 px-4 py-2"
            >
                <TooltipProvider delayDuration={300}>
                    {messages.map((line, idx) =>
                        line.kind === "local" ? (
                            <div
                                key={idx}
                                className="text-muted-foreground mb-3 text-sm whitespace-pre-wrap italic"
                            >
                                {line.data.text}
                            </div>
                        ) : line.kind === "member" ? (
                            <MemberEventRow key={idx} member={line.data} />
                        ) : (
                            <ChatMessageRow key={idx} message={line.data} />
                        ),
                    )}
                </TooltipProvider>
            </div>
            {/* Message input form */}
            <form
                onSubmit={handleSendMessage}
                className="absolute right-0 bottom-2 left-0 flex gap-2"
            >
                <div className="relative flex-1">
                    <ChatCommandSuggestions
                        suggestions={suggestions}
                        activeIndex={activeSuggestion}
                        onPick={applySuggestion}
                    />
                    <Input
                        type="text"
                        placeholder={
                            !user
                                ? t("chat:placeholder_login")
                                : t("chat:placeholder_typing")
                        }
                        disabled={!user}
                        maxLength={CHAT_MESSAGE_MAX_LENGTH}
                        showCount
                        value={inputMessage}
                        onChange={(e) => handleInputChange(e.target.value)}
                        onKeyDown={handleKeyDown}
                    />
                </div>
                <EmotePicker
                    disabled={!user}
                    searchPlaceholder={t("chat:emote_search_placeholder")}
                    emptyStateText={t("chat:emote_empty_state")}
                    getCategoryLabel={(category) =>
                        t(`chat:emote_category_${category}`)
                    }
                    open={emotePickerOpen}
                    onOpenChange={setEmotePickerOpen}
                    searchValue={emoteSearch}
                    onSearchChange={setEmoteSearch}
                    onSelect={handleEmoteSelect}
                />
                <Button type="submit" disabled={!user} className="h-9 w-12 p-0">
                    <IconSend className="!h-6 !w-6" />
                </Button>
            </form>
        </div>
    );
}
