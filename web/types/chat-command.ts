export type ChatCommandScope = "user" | "channel";

export type ChatCommand = {
    id: string;
    scope: ChatCommandScope;
    ownerId: string;
    name: string;
    response: string;
    description: string;
    createdAt: string;
};

export type MyChatCommands = {
    user: ChatCommand[];
    channel: ChatCommand[];
};
