"use client";

import { useState, useEffect, useMemo, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { CommentUser, VODComment } from "@/types/vod-comment";
import { GetUserLikedCommentIds } from "@/lib/api/vod-comment";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import CommentList from "./comment-list";
import CommentForm from "./comment-form";
import { CommentEmpty } from "./comment-empty";
import { Button } from "@/components/ui/button";
import { cn } from "@/utils/cn";
import { useVodCommentsInfinite } from "@/hooks/queries/use-vod-comments";
import { prependVodComment, markVodCommentDeleted } from "@/lib/query/vod-comments-cache";

interface CommentSectionProps {
    vodId: string;
    vodOwnerId?: string;
    className?: string;
}

export default function CommentSection({
    vodId,
    vodOwnerId,
    className,
}: CommentSectionProps) {
    const { t } = useT(["comments", "common", "fetch-error", "api-response"]);
    const user = useUser((state) => state.user);
    const queryClient = useQueryClient();
    const [likedIds, setLikedIds] = useState<Set<string>>(new Set());
    const fetchedLikeIdsRef = useRef<Set<string>>(new Set());

    const { data, isLoading, isFetchingNextPage, hasNextPage, fetchNextPage } =
        useVodCommentsInfinite(vodId);
    const comments = useMemo(
        () => data?.pages.flatMap((p) => p.items) ?? [],
        [data],
    );
    const totalComments = data?.pages[0]?.total ?? 0;

    useEffect(() => {
        if (!user) return;
        const newIds = comments
            .filter((c) => !c.isDeleted && !fetchedLikeIdsRef.current.has(c.id))
            .map((c) => c.id);
        if (newIds.length === 0) return;
        newIds.forEach((id) => fetchedLikeIdsRef.current.add(id));

        GetUserLikedCommentIds(newIds)
            .then((res) => {
                if (res.success && res.data) {
                    setLikedIds((prev) => {
                        const next = new Set(prev);
                        for (const id of res.data!) next.add(id);
                        return next;
                    });
                }
            })
            .catch(() => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "liked-ids-fetch-error",
                    type: "error",
                });
            });
    }, [comments, user, t]);

    const handleCommentCreated = (newComment: VODComment) => {
        const commentWithUser: VODComment =
            !newComment.user && user && user.id === newComment.userId
                ? {
                      ...newComment,
                      user: {
                          id: user.id,
                          username: user.username,
                          profilePicture: user.profilePicture,
                      } satisfies CommentUser,
                  }
                : newComment;
        prependVodComment(queryClient, vodId, commentWithUser);
    };

    const handleCommentDeleted = (commentId: string) => {
        markVodCommentDeleted(queryClient, vodId, commentId);
    };

    const handleLikedChanged = (commentId: string, liked: boolean) => {
        setLikedIds((prev) => {
            const next = new Set(prev);
            if (liked) {
                next.add(commentId);
            } else {
                next.delete(commentId);
            }
            return next;
        });
    };

    return (
        <div className={cn("space-y-4", className)}>
            <h3 className="text-lg font-semibold">
                {t("comments:title")}
                {totalComments > 0 && (
                    <span className="text-muted-foreground ml-2 text-sm font-normal">
                        ({totalComments})
                    </span>
                )}
            </h3>

            {user ? (
                <CommentForm
                    vodId={vodId}
                    onCommentCreated={handleCommentCreated}
                />
            ) : (
                <p className="text-muted-foreground text-sm">
                    {t("comments:login_to_comment")}
                </p>
            )}

            {comments.length === 0 && !isLoading && (
                <CommentEmpty message={t("comments:no_comments")} />
            )}

            <CommentList
                comments={comments}
                vodId={vodId}
                vodOwnerId={vodOwnerId}
                likedIds={likedIds}
                onCommentDeleted={handleCommentDeleted}
                onLikedChanged={handleLikedChanged}
            />

            {hasNextPage && comments.length > 0 && (
                <div className="flex justify-center">
                    <Button
                        variant="ghost"
                        onClick={() => fetchNextPage()}
                        disabled={isLoading || isFetchingNextPage}
                    >
                        {isLoading || isFetchingNextPage
                            ? t("common:loading")
                            : t("common:show_more")}
                    </Button>
                </div>
            )}
        </div>
    );
}
