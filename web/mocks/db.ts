/**
 * In-memory mock database.
 * All handlers read/write from this singleton so mutations persist
 * within a browser session (page reload resets to seed data).
 */

import { AuthProvider, MeUser, PublicUser, UserStatus } from "@/types/user";
import { Notification } from "@/types/notification";
import { Livestream } from "@/types/livestream";
import { VOD } from "@/types/vod";
import { VODComment } from "@/types/vod-comment";
import { ChatCommand } from "@/types/chat-command";
import {
    Conversation,
    ConversationType,
    DmMessage,
    DmMessageType,
    ParticipantRole,
} from "@/types/dm";
import {
    Account,
    AccountBalance,
    AccountStatus,
    AccountType,
    Currency,
    CurrencyCode,
    LedgerEntry,
    Payment,
    PaymentProvider,
    PaymentStatus,
    Transaction,
    TransactionStatus,
    TransactionType,
} from "@/types/wallet";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

let _seq = 1;
export const uid = () => `mock-${_seq++}`;
export const now = () => new Date().toISOString();
export const daysAgo = (d: number) =>
    new Date(Date.now() - d * 86400_000).toISOString();

// ---------------------------------------------------------------------------
// Seed: Users
// ---------------------------------------------------------------------------

export const ME_USER_ID = "user-me-001";

export const meUser: MeUser = {
    id: ME_USER_ID,
    username: "mockuser",
    email: "mockuser@letslive.dev",
    status: UserStatus.NORMAL,
    authProvider: AuthProvider.LOCAL,
    createdAt: daysAgo(120),
    displayName: "Mock User",
    bio: "Just a mock user for local UI testing.",
    profilePicture: "https://api.dicebear.com/9.x/avataaars/svg?seed=mockuser",
    backgroundPicture:
        "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=1200&q=80",
    followerCount: 42,
    streamAPIKey: "mock-stream-key-abc123",
    livestreamInformation: {
        title: "My Awesome Stream",
        description: "Welcome to my channel!",
        thumbnailUrl:
            "https://images.unsplash.com/photo-1511512578047-dfb367046420?w=640&q=80",
    },
    socialMediaLinks: {
        github: "https://github.com/mockuser",
        twitter: "https://twitter.com/mockuser",
    },
};

export const otherUsers: PublicUser[] = [
    {
        id: "user-002",
        username: "streamer_jane",
        email: "jane@letslive.dev",
        status: UserStatus.NORMAL,
        authProvider: AuthProvider.LOCAL,
        createdAt: daysAgo(200),
        displayName: "Jane Streams",
        bio: "Gaming & Tech streamer",
        profilePicture: "https://api.dicebear.com/9.x/avataaars/svg?seed=jane",
        followerCount: 1340,
        livestreamInformation: {
            title: "Jane's Gaming Zone",
            description: "Best games, best vibes",
            thumbnailUrl:
                "https://images.unsplash.com/photo-1542751371-adc38448a05e?w=640&q=80",
        },
        isFollowing: true,
    },
    {
        id: "user-003",
        username: "coder_alex",
        email: "alex@letslive.dev",
        status: UserStatus.NORMAL,
        authProvider: AuthProvider.LOCAL,
        createdAt: daysAgo(90),
        displayName: "Alex Codes",
        bio: "Live coding sessions every weekday",
        profilePicture: "https://api.dicebear.com/9.x/avataaars/svg?seed=alex",
        followerCount: 882,
        livestreamInformation: {
            title: "Coding with Alex",
            description: "Building real apps from scratch",
            thumbnailUrl:
                "https://images.unsplash.com/photo-1461749280684-dccba630e2f6?w=640&q=80",
        },
        isFollowing: false,
    },
    {
        id: "user-004",
        username: "music_sam",
        email: "sam@letslive.dev",
        status: UserStatus.NORMAL,
        authProvider: AuthProvider.LOCAL,
        createdAt: daysAgo(60),
        displayName: "Sam Beats",
        bio: "Producer & DJ streaming live sessions",
        profilePicture: "https://api.dicebear.com/9.x/avataaars/svg?seed=sam",
        followerCount: 3201,
        livestreamInformation: {
            title: "Sam's Studio Sessions",
            description: "Lo-fi beats live",
            thumbnailUrl:
                "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=640&q=80",
        },
        isFollowing: true,
    },
    {
        id: "user-005",
        username: "travel_mia",
        email: "mia@letslive.dev",
        status: UserStatus.NORMAL,
        authProvider: AuthProvider.LOCAL,
        createdAt: daysAgo(30),
        displayName: "Mia Wanderlust",
        bio: "Travel vlogger streaming from around the world",
        profilePicture: "https://api.dicebear.com/9.x/avataaars/svg?seed=mia",
        followerCount: 7890,
        livestreamInformation: {
            title: "Live from Somewhere New",
            description: "Real-time travel content",
            thumbnailUrl:
                "https://images.unsplash.com/photo-1503220317375-aaad61436b1b?w=640&q=80",
        },
        isFollowing: false,
    },
];

// ---------------------------------------------------------------------------
// Seed: Livestreams
// ---------------------------------------------------------------------------

export const livestreams: Livestream[] = [
    {
        id: "ls-001",
        userId: "user-002",
        title: "Jane's Gaming Zone",
        description: "Playing the latest RPG — join the adventure!",
        thumbnailUrl:
            "https://images.unsplash.com/photo-1542751371-adc38448a05e?w=640&q=80",
        viewCount: 234,
        visibility: "public",
        startedAt: daysAgo(0),
        endedAt: null,
        createdAt: daysAgo(0),
        updatedAt: daysAgo(0),
        vodId: null,
    },
    {
        id: "ls-002",
        userId: "user-003",
        title: "Coding with Alex — building a CLI tool",
        description: "Live coding session — Rust CLI from scratch",
        thumbnailUrl:
            "https://images.unsplash.com/photo-1461749280684-dccba630e2f6?w=640&q=80",
        viewCount: 89,
        visibility: "public",
        startedAt: daysAgo(0),
        endedAt: null,
        createdAt: daysAgo(0),
        updatedAt: daysAgo(0),
        vodId: null,
    },
    {
        id: "ls-003",
        userId: "user-004",
        title: "Sam Beats — Late Night Vibes",
        description: "Chilling with some lo-fi production",
        thumbnailUrl:
            "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=640&q=80",
        viewCount: 512,
        visibility: "public",
        startedAt: daysAgo(1),
        endedAt: now(),
        createdAt: daysAgo(1),
        updatedAt: now(),
        vodId: "vod-003",
    },
];

// ---------------------------------------------------------------------------
// Seed: VODs
// ---------------------------------------------------------------------------

export const vods: VOD[] = [
    {
        id: "vod-001",
        livestreamId: null,
        userId: ME_USER_ID,
        title: "My First VOD",
        description: "Testing the upload feature",
        thumbnailUrl:
            "https://images.unsplash.com/photo-1511512578047-dfb367046420?w=640&q=80",
        visibility: "public",
        viewCount: 17,
        duration: 1800,
        playbackUrl: "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8",
        status: "ready",
        originalFileUrl: null,
        createdAt: daysAgo(10),
        updatedAt: daysAgo(10),
    },
    {
        id: "vod-002",
        livestreamId: null,
        userId: ME_USER_ID,
        title: "Private VOD — Do Not Share",
        description: "Just for me",
        thumbnailUrl: null,
        visibility: "private",
        viewCount: 2,
        duration: 600,
        playbackUrl: null,
        status: "ready",
        originalFileUrl: null,
        createdAt: daysAgo(5),
        updatedAt: daysAgo(5),
    },
    {
        id: "vod-003",
        livestreamId: "ls-003",
        userId: "user-004",
        title: "Sam Beats — Late Night Vibes VOD",
        description: "Recorded session from yesterday",
        thumbnailUrl:
            "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=640&q=80",
        visibility: "public",
        viewCount: 441,
        duration: 7200,
        playbackUrl: "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8",
        status: "ready",
        originalFileUrl: null,
        createdAt: daysAgo(1),
        updatedAt: daysAgo(1),
    },
    {
        id: "vod-004",
        livestreamId: null,
        userId: "user-002",
        title: "Jane's Best Moments — Highlight Reel",
        description: "Compilation of the best gaming moments",
        thumbnailUrl:
            "https://images.unsplash.com/photo-1542751371-adc38448a05e?w=640&q=80",
        visibility: "public",
        viewCount: 1102,
        duration: 3600,
        playbackUrl: "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8",
        status: "ready",
        originalFileUrl: null,
        createdAt: daysAgo(15),
        updatedAt: daysAgo(15),
    },
];

// ---------------------------------------------------------------------------
// Seed: VOD Comments
// ---------------------------------------------------------------------------

export const vodComments: VODComment[] = [
    {
        id: "comment-001",
        vodId: "vod-001",
        userId: "user-002",
        parentId: null,
        content: "Great stream! Keep it up 🔥",
        isDeleted: false,
        likeCount: 5,
        replyCount: 1,
        createdAt: daysAgo(9),
        updatedAt: daysAgo(9),
        user: {
            id: "user-002",
            username: "streamer_jane",
            displayName: "Jane Streams",
            profilePicture:
                "https://api.dicebear.com/9.x/avataaars/svg?seed=jane",
        },
    },
    {
        id: "comment-002",
        vodId: "vod-001",
        userId: "user-003",
        parentId: "comment-001",
        content: "Agreed! @streamer_jane always brings energy",
        isDeleted: false,
        likeCount: 2,
        replyCount: 0,
        createdAt: daysAgo(8),
        updatedAt: daysAgo(8),
        user: {
            id: "user-003",
            username: "coder_alex",
            displayName: "Alex Codes",
            profilePicture:
                "https://api.dicebear.com/9.x/avataaars/svg?seed=alex",
        },
    },
    {
        id: "comment-003",
        vodId: "vod-004",
        userId: ME_USER_ID,
        parentId: null,
        content: "This highlight reel is insane",
        isDeleted: false,
        likeCount: 3,
        replyCount: 0,
        createdAt: daysAgo(7),
        updatedAt: daysAgo(7),
        user: {
            id: ME_USER_ID,
            username: "mockuser",
            displayName: "Mock User",
            profilePicture:
                "https://api.dicebear.com/9.x/avataaars/svg?seed=mockuser",
        },
    },
];

// Track which comment IDs the current user has liked
export const likedCommentIds: Set<string> = new Set(["comment-001"]);

// ---------------------------------------------------------------------------
// Seed: Chat Commands
// ---------------------------------------------------------------------------

export const chatCommands: ChatCommand[] = [
    {
        id: "cmd-001",
        scope: "user",
        ownerId: ME_USER_ID,
        name: "!hello",
        response: "Hello there! Welcome to the stream! 👋",
        description: "Greeting command",
        createdAt: daysAgo(30),
    },
    {
        id: "cmd-002",
        scope: "user",
        ownerId: ME_USER_ID,
        name: "!socials",
        response:
            "Follow me on Twitter: @mockuser | GitHub: github.com/mockuser",
        description: "Social links",
        createdAt: daysAgo(20),
    },
    {
        id: "cmd-003",
        scope: "channel",
        ownerId: "user-002",
        name: "!discord",
        response: "Join Jane's Discord: discord.gg/janestreams",
        description: "Jane's Discord invite",
        createdAt: daysAgo(15),
    },
];

// ---------------------------------------------------------------------------
// Seed: Notifications
// ---------------------------------------------------------------------------

export const notifications: Notification[] = [
    {
        id: "notif-001",
        userId: ME_USER_ID,
        type: "follow",
        title: "New Follower",
        message: "streamer_jane started following you",
        actionUrl: "/en/user/user-002",
        actionLabel: "View Profile",
        referenceId: "user-002",
        isRead: false,
        createdAt: daysAgo(1),
    },
    {
        id: "notif-002",
        userId: ME_USER_ID,
        type: "vod_comment",
        title: "New Comment",
        message: 'coder_alex commented on your VOD: "Great stream!"',
        actionUrl: "/en/vod/vod-001",
        actionLabel: "View VOD",
        referenceId: "vod-001",
        isRead: false,
        createdAt: daysAgo(2),
    },
    {
        id: "notif-003",
        userId: ME_USER_ID,
        type: "system",
        title: "Welcome to LetsLive!",
        message: "Your account has been set up. Start streaming today!",
        isRead: true,
        createdAt: daysAgo(120),
    },
];

// ---------------------------------------------------------------------------
// Seed: DM Conversations
// ---------------------------------------------------------------------------

export const conversations: Conversation[] = [
    {
        _id: "conv-001",
        type: ConversationType.DM,
        name: null,
        avatarUrl: null,
        createdBy: ME_USER_ID,
        participants: [
            {
                userId: ME_USER_ID,
                username: "mockuser",
                displayName: "Mock User",
                profilePicture:
                    "https://api.dicebear.com/9.x/avataaars/svg?seed=mockuser",
                role: ParticipantRole.OWNER,
                joinedAt: daysAgo(10),
                lastReadMessageId: "msg-002",
                isMuted: false,
            },
            {
                userId: "user-002",
                username: "streamer_jane",
                displayName: "Jane Streams",
                profilePicture:
                    "https://api.dicebear.com/9.x/avataaars/svg?seed=jane",
                role: ParticipantRole.MEMBER,
                joinedAt: daysAgo(10),
                lastReadMessageId: "msg-002",
                isMuted: false,
            },
        ],
        lastMessage: {
            _id: "msg-002",
            senderId: "user-002",
            senderUsername: "streamer_jane",
            text: "Hey! Thanks for following 😊",
            createdAt: daysAgo(1),
        },
        createdAt: daysAgo(10),
        updatedAt: daysAgo(1),
    },
    {
        _id: "conv-002",
        type: ConversationType.GROUP,
        name: "Collab Squad",
        avatarUrl: null,
        createdBy: "user-003",
        participants: [
            {
                userId: ME_USER_ID,
                username: "mockuser",
                displayName: "Mock User",
                profilePicture:
                    "https://api.dicebear.com/9.x/avataaars/svg?seed=mockuser",
                role: ParticipantRole.MEMBER,
                joinedAt: daysAgo(5),
                lastReadMessageId: null,
                isMuted: false,
            },
            {
                userId: "user-003",
                username: "coder_alex",
                displayName: "Alex Codes",
                profilePicture:
                    "https://api.dicebear.com/9.x/avataaars/svg?seed=alex",
                role: ParticipantRole.OWNER,
                joinedAt: daysAgo(5),
                lastReadMessageId: "msg-004",
                isMuted: false,
            },
            {
                userId: "user-004",
                username: "music_sam",
                displayName: "Sam Beats",
                profilePicture:
                    "https://api.dicebear.com/9.x/avataaars/svg?seed=sam",
                role: ParticipantRole.MEMBER,
                joinedAt: daysAgo(5),
                lastReadMessageId: "msg-004",
                isMuted: false,
            },
        ],
        lastMessage: {
            _id: "msg-004",
            senderId: "user-003",
            senderUsername: "coder_alex",
            text: "Who's free for a collab stream this weekend?",
            createdAt: daysAgo(0),
        },
        createdAt: daysAgo(5),
        updatedAt: daysAgo(0),
    },
];

// ---------------------------------------------------------------------------
// Seed: DM Messages
// ---------------------------------------------------------------------------

export const dmMessages: Record<string, DmMessage[]> = {
    "conv-001": [
        {
            _id: "msg-001",
            conversationId: "conv-001",
            senderId: ME_USER_ID,
            senderUsername: "mockuser",
            type: DmMessageType.TEXT,
            text: "Hey Jane! Love your streams 🎮",
            isDeleted: false,
            readBy: [],
            createdAt: daysAgo(2),
            updatedAt: daysAgo(2),
        },
        {
            _id: "msg-002",
            conversationId: "conv-001",
            senderId: "user-002",
            senderUsername: "streamer_jane",
            type: DmMessageType.TEXT,
            text: "Hey! Thanks for following 😊",
            isDeleted: false,
            readBy: [],
            createdAt: daysAgo(1),
            updatedAt: daysAgo(1),
        },
    ],
    "conv-002": [
        {
            _id: "msg-003",
            conversationId: "conv-002",
            senderId: "user-004",
            senderUsername: "music_sam",
            type: DmMessageType.TEXT,
            text: "Let's set up a collab stream! I'm thinking music + coding?",
            isDeleted: false,
            readBy: [],
            createdAt: daysAgo(1),
            updatedAt: daysAgo(1),
        },
        {
            _id: "msg-004",
            conversationId: "conv-002",
            senderId: "user-003",
            senderUsername: "coder_alex",
            type: DmMessageType.TEXT,
            text: "Who's free for a collab stream this weekend?",
            isDeleted: false,
            readBy: [],
            createdAt: daysAgo(0),
            updatedAt: daysAgo(0),
        },
    ],
};

// ---------------------------------------------------------------------------
// Seed: Finance
// ---------------------------------------------------------------------------

const walletAccountId = "acct-me-001";

export const walletAccount: Account = {
    id: walletAccountId,
    ownerId: ME_USER_ID,
    type: AccountType.USER_WALLET,
    status: AccountStatus.ACTIVE,
    createdAt: daysAgo(120),
    updatedAt: now(),
};

export const walletBalances: AccountBalance[] = [
    {
        accountId: walletAccountId,
        currencyCode: CurrencyCode.SPARK,
        balance: "1250.00",
        lastEntryId: "entry-003",
    },
    {
        accountId: walletAccountId,
        currencyCode: CurrencyCode.FLARE,
        balance: "500.00",
        lastEntryId: "entry-006",
    },
];

export const currencies: Currency[] = [
    { code: CurrencyCode.SPARK, name: "Spark", precision: 2 },
    { code: CurrencyCode.FLARE, name: "Flare", precision: 2 },
];

const ledgerEntries: LedgerEntry[] = [
    {
        id: "entry-001",
        transactionId: "tx-001",
        accountId: walletAccountId,
        currencyCode: CurrencyCode.SPARK,
        amount: "1000.00",
        createdAt: daysAgo(30),
    },
    {
        id: "entry-002",
        transactionId: "tx-002",
        accountId: walletAccountId,
        currencyCode: CurrencyCode.SPARK,
        amount: "250.00",
        createdAt: daysAgo(10),
    },
    {
        id: "entry-003",
        transactionId: "tx-003",
        accountId: walletAccountId,
        currencyCode: CurrencyCode.SPARK,
        amount: "-100.00",
        createdAt: daysAgo(5),
    },
];

export const transactions: Transaction[] = [
    {
        id: "tx-001",
        type: TransactionType.REWARD,
        status: TransactionStatus.COMPLETED,
        reference: null,
        description: "Welcome bonus",
        actorId: ME_USER_ID,
        metadata: null,
        createdAt: daysAgo(30),
        updatedAt: daysAgo(30),
        entries: [ledgerEntries[0]],
    },
    {
        id: "tx-002",
        type: TransactionType.PURCHASE,
        status: TransactionStatus.COMPLETED,
        reference: "pay-001",
        description: "Deposit via Stripe",
        actorId: ME_USER_ID,
        metadata: { provider: "stripe" },
        createdAt: daysAgo(10),
        updatedAt: daysAgo(10),
        entries: [ledgerEntries[1]],
    },
    {
        id: "tx-003",
        type: TransactionType.DONATE,
        status: TransactionStatus.COMPLETED,
        reference: null,
        description: "Donation to streamer_jane",
        actorId: ME_USER_ID,
        metadata: { recipientId: "user-002" },
        createdAt: daysAgo(5),
        updatedAt: daysAgo(5),
        entries: [ledgerEntries[2]],
    },
];

export const payments: Payment[] = [
    {
        id: "pay-001",
        transactionId: "tx-002",
        provider: PaymentProvider.STRIPE,
        providerReference: "pi_mock_abc123",
        currencyCode: CurrencyCode.SPARK,
        amount: "250.00",
        status: PaymentStatus.COMPLETED,
        createdAt: daysAgo(10),
        updatedAt: daysAgo(10),
    },
];

// ---------------------------------------------------------------------------
// Chat messages (in-room, not DM) — keyed by roomId (= userId of streamer)
// ---------------------------------------------------------------------------

export type ChatMessage = {
    id: string;
    roomId: string;
    userId: string;
    username: string;
    content: string;
    createdAt: string;
};

export const chatMessages: Record<string, ChatMessage[]> = {
    "user-002": [
        {
            id: "chat-001",
            roomId: "user-002",
            userId: "user-003",
            username: "coder_alex",
            content: "Let's goooo! 🎮",
            createdAt: daysAgo(0),
        },
        {
            id: "chat-002",
            roomId: "user-002",
            userId: "user-004",
            username: "music_sam",
            content: "first!!",
            createdAt: daysAgo(0),
        },
    ],
};
