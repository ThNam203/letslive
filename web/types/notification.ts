export type Notification = {
    id: string;
    userId: string;
    type: string;
    title: string;
    message: string;
    actionUrl: string | null;
    actionLabel: string | null;
    referenceId: string | null;
    isRead: boolean;
    createdAt: string;
};

export type UnreadCountResponse = {
    count: number;
};
