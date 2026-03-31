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
import useT from "@/hooks/use-translation";
import { CHAT_MESSAGE_MAX_LENGTH } from "@/constant/field-limits";

export default function ChatPanel({
    roomId,
    onClose,
}: {
    roomId: string;
    onClose: () => any;
}) {
    const user = useUser((state) => state.user);
    const [messages, setMessages] = useState<ReceivedMessage[]>([]);
    const [inputMessage, setInputMessage] = useState("");
    const wsRef = useRef<WebSocket | null>(null);
    const [atBottom, setAtBottom] = useState(false);
    const messageContainerRef = useRef<HTMLDivElement | null>(null);
    const { t } = useT("users");

    const handleSendMessage = (e: React.FormEvent) => {
        e.preventDefault();
        if (inputMessage.trim()) {
            const newMessage: SendMessage = {
                userId: user!.id,
                roomId: roomId,
                type: "message",
                username: user!.displayName ?? user!.username,
                text: inputMessage.trim(),
            };

            setInputMessage("");
            wsRef.current?.send(JSON.stringify(newMessage));
        }
    };

    useEffect(() => {
        const fetchMessages = async () => {
            const res = await GetMessages(roomId);
            if (res.messages) {
                setMessages((prev) => [...res.messages, ...prev]);
            }
        };

        fetchMessages();
    }, [roomId]);

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
                setMessages((prev) =>
                    prev.length >= 100
                        ? [...prev.slice(1), data]
                        : [...prev, data],
                );
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
                className="border-border mb-14 flex-1 overflow-y-auto rounded-md rounded-t-none border border-t-0 px-4 py-2"
            >
                {messages.map((message, idx) => (
                    <div key={idx} className="mb-3">
                        <span
                            style={{
                                color: `${uuidToReadableHexColor(
                                    message.userId,
                                )}`,
                            }}
                            className="mr-2 font-semibold"
                        >
                            {message.username}:
                        </span>
                        <span className="text-foreground">
                            {message.type === "join"
                                ? t("users:chat.joined")
                                : message.type === "leave"
                                  ? t("users:chat.left")
                                  : message.text}
                        </span>
                    </div>
                ))}
            </div>
            {/* Message input form */}
            <form
                onSubmit={handleSendMessage}
                className="absolute right-0 bottom-2 left-0 flex gap-2"
            >
                <div className="relative flex-1">
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
                        onChange={(e) => setInputMessage(e.target.value)}
                    />
                </div>
                <Button type="submit" disabled={!user} className="h-9 w-12 p-0">
                    <IconSend className="!h-6 !w-6" />
                </Button>
            </form>
        </div>
    );
}

function uuidToReadableHexColor(uuid: string): string {
    // Remove dashes
    const hex: string = uuid.replace(/-/g, "");

    // Take the first 6 valid hexadecimal characters
    let color: string = `#${hex.substring(0, 6)}`;

    // Convert to RGB
    let r: number = parseInt(color.substring(1, 3), 16);
    let g: number = parseInt(color.substring(3, 5), 16);
    let b: number = parseInt(color.substring(5, 7), 16);

    // Convert RGB to HSL
    let hsl: { h: number; s: number; l: number } = rgbToHsl(r, g, b);

    // Ensure the color is dark enough by capping lightness
    if (hsl.l > 0.7) hsl.l = 0.7; // Limit max lightness to 70%

    // Convert back to RGB
    let adjustedRgb: { r: number; g: number; b: number } = hslToRgb(
        hsl.h,
        hsl.s,
        hsl.l,
    );

    // Convert RGB back to HEX
    return rgbToHex(adjustedRgb.r, adjustedRgb.g, adjustedRgb.b);
}

// Convert RGB to HSL
function rgbToHsl(
    r: number,
    g: number,
    b: number,
): { h: number; s: number; l: number } {
    ((r /= 255), (g /= 255), (b /= 255));
    let max: number = Math.max(r, g, b),
        min: number = Math.min(r, g, b);
    let h: number = 0,
        s: number = 0,
        l: number = (max + min) / 2;

    if (max !== min) {
        let d: number = max - min;
        s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
        switch (max) {
            case r:
                h = (g - b) / d + (g < b ? 6 : 0);
                break;
            case g:
                h = (b - r) / d + 2;
                break;
            case b:
                h = (r - g) / d + 4;
                break;
        }
        h /= 6;
    }

    return { h, s, l };
}

// Convert HSL to RGB
function hslToRgb(
    h: number,
    s: number,
    l: number,
): { r: number; g: number; b: number } {
    let r: number, g: number, b: number;

    if (s === 0) {
        r = g = b = l; // Achromatic
    } else {
        let q: number = l < 0.5 ? l * (1 + s) : l + s - l * s;
        let p: number = 2 * l - q;
        r = hue2rgb(p, q, h + 1 / 3);
        g = hue2rgb(p, q, h);
        b = hue2rgb(p, q, h - 1 / 3);
    }

    return {
        r: Math.round(r * 255),
        g: Math.round(g * 255),
        b: Math.round(b * 255),
    };
}

// Convert RGB to HEX
function rgbToHex(r: number, g: number, b: number): string {
    return `#${((1 << 24) | (r << 16) | (g << 8) | b).toString(16).slice(1)}`;
}

function hue2rgb(p: number, q: number, t: number): number {
    if (t < 0) t += 1;
    if (t > 1) t -= 1;
    if (t < 1 / 6) return p + (q - p) * 6 * t;
    if (t < 1 / 2) return q;
    if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
    return p;
}
