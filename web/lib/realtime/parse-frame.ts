import type { RealtimeServerFrame } from "@/types/realtime";

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === "object" && value !== null;
}

export function parseRealtimeFrame(raw: unknown): RealtimeServerFrame | null {
    if (typeof raw !== "string") return null;

    let parsed: unknown;
    try {
        parsed = JSON.parse(raw);
    } catch {
        return null;
    }
    if (!isRecord(parsed)) return null;

    switch (parsed.op) {
        case "event":
            return typeof parsed.topic === "string" &&
                typeof parsed.type === "string"
                ? {
                      op: "event",
                      topic: parsed.topic,
                      type: parsed.type,
                      data: parsed.data,
                  }
                : null;
        case "error":
            return typeof parsed.code === "string"
                ? { op: "error", code: parsed.code }
                : null;
        case "pong":
            return { op: "pong" };
        default:
            return null;
    }
}
