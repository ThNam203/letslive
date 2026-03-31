export type CommentUser = {
    id: string;
    username: string;
    displayName?: string;
    profilePicture?: string;
};

export type VODComment = {
    id: string;
    vodId: string;
    userId: string;
    parentId: string | null;
    content: string;
    isDeleted: boolean;
    likeCount: number;
    replyCount: number;
    createdAt: string;
    updatedAt: string;
    user?: CommentUser;
};

export type CreateVODCommentRequest = {
    content: string;
    parentId?: string;
};
