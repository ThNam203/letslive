export type RealtimeEventFrame = {
    op: "event";
    topic: string;
    type: string;
    data: unknown;
};

export type RealtimeErrorFrame = {
    op: "error";
    code: string;
};

export type RealtimePongFrame = {
    op: "pong";
};

export type RealtimeServerFrame =
    RealtimeEventFrame | RealtimeErrorFrame | RealtimePongFrame;
