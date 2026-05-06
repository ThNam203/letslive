import twemoji from "twemoji";

export type Emote = {
    code: string;
    codepoint: string;
    emoji: string;
    name: string;
    category: "smileys" | "gestures" | "hype" | "misc";
};

type EmoteDefinition = Omit<Emote, "emoji">;

const EMOTE_DEFINITIONS: EmoteDefinition[] = [
    // Smileys
    {
        code: "smile",
        codepoint: "1f60a",
        name: "Smile",
        category: "smileys",
    },
    {
        code: "laugh",
        codepoint: "1f602",
        name: "Laugh",
        category: "smileys",
    },
    { code: "love", codepoint: "1f60d", name: "Love", category: "smileys" },
    { code: "cool", codepoint: "1f60e", name: "Cool", category: "smileys" },
    {
        code: "thinking",
        codepoint: "1f914",
        name: "Thinking",
        category: "smileys",
    },
    { code: "grin", codepoint: "1f601", name: "Grin", category: "smileys" },
    { code: "joy", codepoint: "1f923", name: "ROFL", category: "smileys" },
    { code: "sweat", codepoint: "1f605", name: "Sweat Smile", category: "smileys" },
    { code: "plead", codepoint: "1f97a", name: "Pleading", category: "smileys" },
    { code: "mindblown", codepoint: "1f92f", name: "Mind Blown", category: "smileys" },
    { code: "melting", codepoint: "1fae0", name: "Melting", category: "smileys" },
    { code: "partyface", codepoint: "1f973", name: "Party Face", category: "smileys" },
    { code: "monocle", codepoint: "1f9d0", name: "Monocle", category: "smileys" },
    { code: "cry", codepoint: "1f622", name: "Cry", category: "smileys" },
    { code: "angry", codepoint: "1f621", name: "Angry", category: "smileys" },
    { code: "wink", codepoint: "1f609", name: "Wink", category: "smileys" },
    { code: "sus", codepoint: "1f928", name: "Sus", category: "smileys" },
    { code: "skull", codepoint: "1f480", name: "Skull", category: "smileys" },

    // Gestures
    {
        code: "thumbsup",
        codepoint: "1f44d",
        name: "Thumbs Up",
        category: "gestures",
    },
    {
        code: "thumbsdown",
        codepoint: "1f44e",
        name: "Thumbs Down",
        category: "gestures",
    },
    { code: "clap", codepoint: "1f44f", name: "Clap", category: "gestures" },
    { code: "wave", codepoint: "1f44b", name: "Wave", category: "gestures" },
    { code: "ok", codepoint: "1f44c", name: "OK", category: "gestures" },
    { code: "peace", codepoint: "270c-fe0f", name: "Peace", category: "gestures" },
    { code: "fist", codepoint: "270a", name: "Fist", category: "gestures" },
    { code: "muscle", codepoint: "1f4aa", name: "Flex", category: "gestures" },
    { code: "pointup", codepoint: "1f446", name: "Point Up", category: "gestures" },
    { code: "pray", codepoint: "1f64f", name: "Pray", category: "gestures" },
    {
        code: "salute",
        codepoint: "1fae1",
        name: "Salute",
        category: "gestures",
    },

    // Hype
    { code: "fire", codepoint: "1f525", name: "Fire", category: "hype" },
    { code: "heart", codepoint: "2764-fe0f", name: "Heart", category: "hype" },
    { code: "star", codepoint: "2b50", name: "Star", category: "hype" },
    { code: "hype", codepoint: "1f389", name: "Hype", category: "hype" },
    { code: "gg", codepoint: "1f3c6", name: "GG", category: "hype" },
    { code: "sparkles", codepoint: "2728", name: "Sparkles", category: "hype" },
    { code: "boom", codepoint: "1f4a5", name: "Boom", category: "hype" },
    { code: "zap", codepoint: "26a1", name: "Zap", category: "hype" },
    { code: "target", codepoint: "1f3af", name: "Target", category: "hype" },
    { code: "medal", codepoint: "1f3c5", name: "Medal", category: "hype" },
    { code: "tada", codepoint: "1f38a", name: "Tada", category: "hype" },
    { code: "pog", codepoint: "1f62e", name: "Pog", category: "hype" },
    { code: "ez", codepoint: "1f60f", name: "EZ", category: "hype" },
    { code: "goat", codepoint: "1f410", name: "GOAT", category: "hype" },

    // Misc
    { code: "eyes", codepoint: "1f440", name: "Eyes", category: "misc" },
    { code: "100", codepoint: "1f4af", name: "100", category: "misc" },
    { code: "money", codepoint: "1f4b0", name: "Money", category: "misc" },
    { code: "ghost", codepoint: "1f47b", name: "Ghost", category: "misc" },
    { code: "rocket", codepoint: "1f680", name: "Rocket", category: "misc" },
    { code: "crown", codepoint: "1f451", name: "Crown", category: "misc" },
    { code: "gem", codepoint: "1f48e", name: "Gem", category: "misc" },
    { code: "cookie", codepoint: "1f36a", name: "Cookie", category: "misc" },
    { code: "coffee", codepoint: "2615", name: "Coffee", category: "misc" },
    { code: "pizza", codepoint: "1f355", name: "Pizza", category: "misc" },
    { code: "joystick", codepoint: "1f579-fe0f", name: "Joystick", category: "misc" },
    { code: "keyboard", codepoint: "2328-fe0f", name: "Keyboard", category: "misc" },
    { code: "headphones", codepoint: "1f3a7", name: "Headphones", category: "misc" },
    { code: "moon", codepoint: "1f319", name: "Moon", category: "misc" },
];

export const EMOTES: Emote[] = EMOTE_DEFINITIONS.map((emote) => ({
    ...emote,
    emoji: twemoji.convert.fromCodePoint(emote.codepoint),
}));

export const EMOTE_MAP = new Map<string, Emote>(EMOTES.map((e) => [e.code, e]));

export const EMOTE_CATEGORIES = [
    "smileys",
    "gestures",
    "hype",
    "misc",
] as const;
