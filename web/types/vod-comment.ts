export type CommentUser = {
    id: string;
    username: string;
    profilePicture?: string;
};

export type VODComment = {
    id: string;
    vodId: string;
    userId: string;
    parentId: string | null;
    content: string;
    isDeleted: boolean;
    isEdited: boolean;
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

export type UpdateVODCommentRequest = {
    content: string;
};
