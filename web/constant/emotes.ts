export type Emote = {
    code: string;
    emoji: string;
    name: string;
    category: "smileys" | "gestures" | "hype" | "misc";
};

export const EMOTES: Emote[] = [
    // Smileys
    { code: "smile", emoji: "😊", name: "Smile", category: "smileys" },
    { code: "laugh", emoji: "😂", name: "Laugh", category: "smileys" },
    { code: "love", emoji: "😍", name: "Love", category: "smileys" },
    { code: "cool", emoji: "😎", name: "Cool", category: "smileys" },
    { code: "thinking", emoji: "🤔", name: "Thinking", category: "smileys" },
    { code: "cry", emoji: "😢", name: "Cry", category: "smileys" },
    { code: "angry", emoji: "😡", name: "Angry", category: "smileys" },
    { code: "wink", emoji: "😉", name: "Wink", category: "smileys" },
    { code: "sus", emoji: "🤨", name: "Sus", category: "smileys" },
    { code: "skull", emoji: "💀", name: "Skull", category: "smileys" },

    // Gestures
    { code: "thumbsup", emoji: "👍", name: "Thumbs Up", category: "gestures" },
    {
        code: "thumbsdown",
        emoji: "👎",
        name: "Thumbs Down",
        category: "gestures",
    },
    { code: "clap", emoji: "👏", name: "Clap", category: "gestures" },
    { code: "wave", emoji: "👋", name: "Wave", category: "gestures" },
    { code: "pray", emoji: "🙏", name: "Pray", category: "gestures" },
    { code: "salute", emoji: "🫡", name: "Salute", category: "gestures" },

    // Hype
    { code: "fire", emoji: "🔥", name: "Fire", category: "hype" },
    { code: "heart", emoji: "❤️", name: "Heart", category: "hype" },
    { code: "star", emoji: "⭐", name: "Star", category: "hype" },
    { code: "hype", emoji: "🎉", name: "Hype", category: "hype" },
    { code: "gg", emoji: "🏆", name: "GG", category: "hype" },
    { code: "pog", emoji: "😮", name: "Pog", category: "hype" },
    { code: "ez", emoji: "😏", name: "EZ", category: "hype" },
    { code: "goat", emoji: "🐐", name: "GOAT", category: "hype" },

    // Misc
    { code: "eyes", emoji: "👀", name: "Eyes", category: "misc" },
    { code: "100", emoji: "💯", name: "100", category: "misc" },
    { code: "money", emoji: "💰", name: "Money", category: "misc" },
    { code: "ghost", emoji: "👻", name: "Ghost", category: "misc" },
    { code: "rocket", emoji: "🚀", name: "Rocket", category: "misc" },
    { code: "crown", emoji: "👑", name: "Crown", category: "misc" },
];

export const EMOTE_MAP = new Map<string, Emote>(EMOTES.map((e) => [e.code, e]));

export const EMOTE_CATEGORIES = [
    "smileys",
    "gestures",
    "hype",
    "misc",
] as const;
