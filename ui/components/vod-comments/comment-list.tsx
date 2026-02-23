"use client";

import { VODComment } from "@/types/vod-comment";
import CommentItem from "./comment-item";

interface CommentListProps {
    comments: VODComment[];
    vodId: string;
    vodOwnerId?: string;
    likedIds?: Set<string>;
    onCommentDeleted: (commentId: string) => void;
    onLikedChanged?: (commentId: string, liked: boolean) => void;
    isReplyList?: boolean;
    depth?: number;
}

export default function CommentList({
    comments,
    vodId,
    vodOwnerId,
    likedIds,
    onCommentDeleted,
    onLikedChanged,
    isReplyList = false,
    depth = 0,
}: CommentListProps) {
    if (comments.length === 0) return null;

    // Reduce indentation at deeper levels to prevent content from getting too narrow
    const replyIndent =
        depth <= 1
            ? "ml-8 mt-2 space-y-3 border-l-2 border-border pl-4"
            : "ml-4 mt-2 space-y-3 border-l-2 border-border pl-3";

    return (
        <div className={isReplyList ? replyIndent : "space-y-4"}>
            {comments.map((comment) => (
                <CommentItem
                    key={comment.id}
                    comment={comment}
                    vodId={vodId}
                    vodOwnerId={vodOwnerId}
                    likedIds={likedIds}
                    onCommentDeleted={onCommentDeleted}
                    onLikedChanged={onLikedChanged}
                    depth={depth}
                />
            ))}
        </div>
    );
}
