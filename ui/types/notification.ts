export type Notification = {
    id: string;
    userId: string;
    type: string;
    title: string;
    message: string;
    actionUrl?: string;
    actionLabel?: string;
    referenceId?: string;
    isRead: boolean;
    createdAt: string;
};

export type UnreadCountResponse = {
    count: number;
};
