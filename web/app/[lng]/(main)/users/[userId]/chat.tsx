"use client";

import type React from "react";
import { useState, useRef, useEffect } from "react";
import { toast } from "@/components/utils/toast";
import useUser from "@/hooks/user";
import { ReceivedMessage, SendMessage } from "@/types/message";
import { GetMessages } from "@/lib/api/chat";
import GLOBAL from "@/global";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import IconClose from "@/components/icons/close";
import IconSend from "@/components/icons/send";
import EmotePicker from "@/components/emote-picker";
import ChatCommandSuggestions from "@/components/chat-command-suggestions";
import { uuidToReadableHexColor } from "@/utils/uuid-color";
import {
    buildChatCommandHelpText,
    buildChatCommandIndex,
    ChatCommandSuggestion,
    filterChatCommandSuggestions,
    parseChatCommand,
    parseEmotes,
} from "@/utils/chat-parser";
import { GetRoomChatCommands } from "@/lib/api/chat-command";
import { ChatCommand } from "@/types/chat-command";
import useT from "@/hooks/use-translation";
import { CHAT_MESSAGE_MAX_LENGTH } from "@/constant/field-limits";

type LocalMessage = {
    kind: "system";
    text: string;
};

type ChatLine =
    | { kind: "remote"; data: ReceivedMessage }
    | { kind: "local"; data: LocalMessage };

export default function ChatPanel({
    roomId,
    onClose,
}: {
    roomId: string;
    onClose: () => any;
}) {
    const user = useUser((state) => state.user);
    const [messages, setMessages] = useState<ChatLine[]>([]);
    const [inputMessage, setInputMessage] = useState("");
    const wsRef = useRef<WebSocket | null>(null);
    const [atBottom, setAtBottom] = useState(false);
    const messageContainerRef = useRef<HTMLDivElement | null>(null);
    const { t } = useT(["users", "chat-commands"]);
    const [customChatCommands, setCustomChatCommands] = useState<ChatCommand[]>(
        [],
    );
    const [suggestions, setSuggestions] = useState<ChatCommandSuggestion[]>([]);
    const [activeSuggestion, setActiveSuggestion] = useState(0);

    const chatCommandIndex = buildChatCommandIndex(customChatCommands, t);

    const appendLine = (line: ChatLine) =>
        setMessages((prev) =>
            prev.length >= 100 ? [...prev.slice(1), line] : [...prev, line],
        );

    const sendText = (text: string) => {
        const newMessage: SendMessage = {
            userId: user!.id,
            roomId: roomId,
            type: "message",
            username: user!.displayName ?? user!.username,
            text,
        };
        wsRef.current?.send(JSON.stringify(newMessage));
    };

    const handleSendMessage = (e: React.FormEvent) => {
        e.preventDefault();
        const raw = inputMessage.trim();
        if (!raw || !user) return;

        if (raw.startsWith("/")) {
            const result = parseChatCommand(raw, customChatCommands, t);
            setInputMessage("");
            setSuggestions([]);
            if (!result || result.kind === "noop") return;
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
        setSuggestions([]);
        setActiveSuggestion(0);
    };

    const handleInputChange = (value: string) => {
        setInputMessage(value);
        const next = filterChatCommandSuggestions(chatCommandIndex, value);
        setSuggestions(next);
        setActiveSuggestion(0);
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
        const fetchMessages = async () => {
            const res = await GetMessages(roomId);
            if (res.messages) {
                const lines: ChatLine[] = res.messages.map((m) => ({
                    kind: "remote",
                    data: m,
                }));
                setMessages((prev) => [...lines, ...prev]);
            }
        };

        fetchMessages();
    }, [roomId]);

    useEffect(() => {
        let cancelled = false;
        GetRoomChatCommands(roomId)
            .then((res) => {
                if (!cancelled && res.success && res.data) {
                    setCustomChatCommands(res.data);
                }
            })
            .catch(() => {});
        return () => {
            cancelled = true;
        };
    }, [roomId, user?.id]);

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

    useEffect(() => {
        const connectWebSocket = async () => {
            const ws = new WebSocket(GLOBAL.WS_SERVER_URL);

            ws.onopen = () => {
                wsRef.current = ws;
                if (user) {
                    ws.send(
                        JSON.stringify({
                            type: "join",
                            roomId: roomId,
                            userId: user.id,
                            username: user.displayName ?? user.username,
                        }),
                    );
                }
            };

            ws.onmessage = (event) => {
                const data: ReceivedMessage = JSON.parse(event.data);
                appendLine({ kind: "remote", data });
            };

            ws.onclose = (ev) => {};

            ws.onerror = (error) => {};
        };

        connectWebSocket();
        return () => {
            if (wsRef.current) {
                if (user) {
                    wsRef.current.send(
                        JSON.stringify({
                            type: "leave",
                            roomId: roomId,
                            userId: user.id,
                            username: user!.displayName ?? user!.username,
                        }),
                    );
                } else wsRef.current.close();
            }
        };
    }, [user, roomId]);

    return (
        <div className="relative flex h-full w-full flex-col">
            <div className="border-border flex items-center justify-between border border-y-0 px-4 py-3">
                <h2 className="font-semibold">{t("users:chat.title")}</h2>
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
                {messages.map((line, idx) => {
                    if (line.kind === "local") {
                        return (
                            <div
                                key={idx}
                                className="text-muted-foreground mb-3 text-sm whitespace-pre-wrap italic"
                            >
                                {line.data.text}
                            </div>
                        );
                    }
                    const message = line.data;
                    const isAction =
                        message.type === "message" &&
                        message.text.startsWith("_") &&
                        message.text.endsWith("_") &&
                        message.text.length >= 2;
                    const displayText = isAction
                        ? message.text.slice(1, -1)
                        : message.text;
                    return (
                        <div key={idx} className="mb-3">
                            <span
                                style={{
                                    color: `${uuidToReadableHexColor(
                                        message.userId,
                                    )}`,
                                }}
                                className="mr-2 font-semibold"
                            >
                                {message.username}
                                {isAction ? "" : ":"}
                            </span>
                            <span
                                className={`text-foreground ${isAction ? "italic" : ""}`}
                            >
                                {message.type === "join"
                                    ? t("users:chat.joined")
                                    : message.type === "leave"
                                      ? t("users:chat.left")
                                      : parseEmotes(displayText)}
                            </span>
                        </div>
                    );
                })}
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
                                ? t("users:chat.placeholder_login")
                                : t("users:chat.placeholder_typing")
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
                    onSelect={(code) => handleInputChange(inputMessage + code)}
                />
                <Button type="submit" disabled={!user} className="h-9 w-12 p-0">
                    <IconSend className="!h-6 !w-6" />
                </Button>
            </form>
        </div>
    );
}
