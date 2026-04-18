import { http } from "msw";
import { API_BASE, ok, noContent, created, notFound } from "../utils";
import { chatCommands, chatMessages, ME_USER_ID, uid, now } from "../db";
import { ChatCommand, MyChatCommands } from "@/types/chat-command";
import { ChatMessage } from "../db";

export const chatHandlers = [
    // GET /messages?roomId=
    http.get(`${API_BASE}/messages`, ({ request }) => {
        const url = new URL(request.url);
        const roomId = url.searchParams.get("roomId") ?? "";
        const messages = chatMessages[roomId] ?? [];
        return ok<ChatMessage[]>(messages);
    }),

    // GET /chat-commands?roomId=  — channel-scoped commands for a room
    http.get(`${API_BASE}/chat-commands`, ({ request }) => {
        const url = new URL(request.url);
        const roomId = url.searchParams.get("roomId") ?? "";
        const cmds = chatCommands.filter(
            (c) => c.scope === "channel" && c.ownerId === roomId,
        );
        return ok<ChatCommand[]>(cmds);
    }),

    // GET /chat-commands/mine — split by scope
    http.get(`${API_BASE}/chat-commands/mine`, () => {
        const mine: MyChatCommands = {
            user: chatCommands.filter(
                (c) => c.scope === "user" && c.ownerId === ME_USER_ID,
            ),
            channel: chatCommands.filter(
                (c) => c.scope === "channel" && c.ownerId === ME_USER_ID,
            ),
        };
        return ok<MyChatCommands>(mine);
    }),

    // POST /chat-commands
    http.post(`${API_BASE}/chat-commands`, async ({ request }) => {
        const body = (await request.json()) as Omit<
            ChatCommand,
            "id" | "ownerId" | "createdAt"
        >;
        const newCmd: ChatCommand = {
            id: uid(),
            ownerId: ME_USER_ID,
            scope: body.scope,
            name: body.name,
            response: body.response,
            description: body.description,
            createdAt: now(),
        };
        chatCommands.push(newCmd);
        return created<ChatCommand>(newCmd);
    }),

    // PATCH /chat-commands/:id
    http.patch(`${API_BASE}/chat-commands/:id`, async ({ params, request }) => {
        const { id } = params as { id: string };
        const cmd = chatCommands.find((c) => c.id === id);
        if (!cmd) return notFound("res_err_not_found", "Command not found");
        const body = (await request.json()) as Partial<ChatCommand>;
        Object.assign(cmd, body);
        return ok<ChatCommand>(cmd);
    }),

    // DELETE /chat-commands/:id
    http.delete(`${API_BASE}/chat-commands/:id`, ({ params }) => {
        const { id } = params as { id: string };
        const idx = chatCommands.findIndex((c) => c.id === id);
        if (idx !== -1) chatCommands.splice(idx, 1);
        return noContent();
    }),
];
