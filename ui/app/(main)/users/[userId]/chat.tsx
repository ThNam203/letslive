"use client";

import type React from "react";
import { useState, useRef, useEffect } from "react";
import { Send, X } from "lucide-react";
import { toast } from "react-toastify";
import useUser from "../../../../hooks/user";
import { ReceivedMessage, SendMessage } from "../../../../types/message";
import { GetMessages } from "../../../../lib/api/chat";
import GLOBAL from "../../../../global";
import { Input } from "../../../../components/ui/input";
import { Button } from "../../../../components/ui/button";

export default function ChatPanel({ roomId, onClose }: { roomId: string, onClose: () => any }) {
    const user = useUser((state) => state.user);
    const [messages, setMessages] = useState<ReceivedMessage[]>([]);
    const [inputMessage, setInputMessage] = useState("");
    const wsRef = useRef<WebSocket | null>(null);

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
            const { messages, fetchError } = await GetMessages(roomId);
            if (fetchError) {
                toast.error(fetchError.message, {
                    toastId: "message-fetch-error",
                });
            } else {
                setMessages(prev => [...messages, ...prev]);
            }
        };

        fetchMessages();
    }, []);

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
                        })
                    );
                }
            };

            ws.onmessage = (event) => {
                const data: ReceivedMessage = JSON.parse(event.data);
                if (messages.length >= 100)
                    setMessages((prev) => [...prev.slice(1), data]);
                else setMessages((prev) => [...prev, data]);
            };

            ws.onclose = (ev) => {

            };

            ws.onerror = (error) => {

            };
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
                        })
                    );
                } else wsRef.current.close();
            }
        };
    }, [user]);

    return (
        <div className="w-full h-full flex flex-col">
            <div className="border border-y-0 px-4 py-3 flex justify-between items-center">
                <h2 className="font-semibold">Chat</h2>
                <Button variant="ghost" size="icon" onClick={onClose} className="md:hidden">
                    <X className="h-4 w-4" />
                </Button>
            </div>
            <div className="flex-1 overflow-y-auto mb-4 border border-t-0 border-gray-200 rounded-md rounded-t-none px-4 py-2 bg-gray-50">
                {messages.map((message, idx) => (
                    <div key={idx} className="mb-3">
                        <span
                            style={{
                                color: `${uuidToReadableHexColor(
                                    message.userId
                                )}`,
                            }}
                            className="font-semibold mr-2"
                        >
                            {message.username}:
                        </span>
                        <span className="text-gray-800">
                            {message.type === "join"
                                ? "joined the chat"
                                : message.type === "leave"
                                    ? "left the chat"
                                    : message.text}
                        </span>
                    </div>
                ))}
            </div>

            {/* Message input form */}
            <form onSubmit={handleSendMessage} className="flex gap-2 mb-2">
                <Input
                    type="text"
                    placeholder={!user ? "Login to start messaging" : "Type a message..."}
                    disabled={!user}
                    value={inputMessage}
                    onChange={(e) => setInputMessage(e.target.value)}
                    className="flex-1"
                />
                <Button
                    type="submit"
                    className="bg-purple-600 hover:bg-purple-700 text-white"
                    disabled={!user}
                >
                    <Send className="w-4 h-4" />
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
        hsl.l
    );

    // Convert RGB back to HEX
    return rgbToHex(adjustedRgb.r, adjustedRgb.g, adjustedRgb.b);
}

// Convert RGB to HSL
function rgbToHsl(
    r: number,
    g: number,
    b: number
): { h: number; s: number; l: number } {
    (r /= 255), (g /= 255), (b /= 255);
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
    l: number
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

function getCookie(name: string) {
    var nameEQ = name + "=";
    var ca = document.cookie.split(";");
    for (var i = 0; i < ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) == " ") c = c.substring(1, c.length);
        if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length, c.length);
    }
    return null;
}
