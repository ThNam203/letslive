"use client";

import { useState, useEffect, useCallback } from "react";
import { CommentUser, VODComment } from "@/types/vod-comment";
import { GetVODComments, GetUserLikedCommentIds } from "@/lib/api/vod-comment";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import CommentList from "./comment-list";
import CommentForm from "./comment-form";
import { CommentEmpty } from "./comment-empty";
import { Button } from "@/components/ui/button";
import { cn } from "@/utils/cn";

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
    const [comments, setComments] = useState<VODComment[]>([]);
    const [page, setPage] = useState(0);
    const [hasMore, setHasMore] = useState(true);
    const [isLoading, setIsLoading] = useState(false);
    const [likedIds, setLikedIds] = useState<Set<string>>(new Set());
    const [totalComments, setTotalComments] = useState(0);
    const LIMIT = 10;

    const fetchLikedIds = useCallback(
        async (commentList: VODComment[]) => {
            if (!user) return;
            const ids = commentList
                .filter((c) => !c.isDeleted)
                .map((c) => c.id);
            if (ids.length === 0) return;
            try {
                const res = await GetUserLikedCommentIds(ids);
                if (res.success && res.data) {
                    setLikedIds((prev) => {
                        const next = new Set(prev);
                        for (const id of res.data!) next.add(id);
                        return next;
                    });
                }
            } catch (_) {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "liked-ids-fetch-error",
                    type: "error",
                });
            }
        },
        [user, t],
    );

    const fetchComments = useCallback(
        async (pageNum: number) => {
            setIsLoading(true);
            try {
                const res = await GetVODComments(vodId, pageNum, LIMIT);
                if (res.success) {
                    const newComments = res.data ?? [];
                    if (pageNum === 0) {
                        setComments(newComments);
                    } else {
                        setComments((prev) => [...prev, ...newComments]);
                    }
                    setHasMore(newComments.length === LIMIT);
                    if (res.meta?.total !== undefined) {
                        setTotalComments(res.meta.total);
                    }
                    fetchLikedIds(newComments);
                } else {
                    toast(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                }
            } catch (_) {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
            } finally {
                setIsLoading(false);
            }
        },
        [vodId, t, fetchLikedIds],
    );

    useEffect(() => {
        setPage(0);
        setComments([]);
        setLikedIds(new Set());
        fetchComments(0);
    }, [fetchComments]);

    const handleCommentCreated = (newComment: VODComment) => {
        const commentWithUser: VODComment =
            !newComment.user && user && user.id === newComment.userId
                ? {
                      ...newComment,
                      user: {
                          id: user.id,
                          username: user.username,
                          displayName: user.displayName,
                          profilePicture: user.profilePicture,
                      } satisfies CommentUser,
                  }
                : newComment;
        setComments((prev) => [commentWithUser, ...prev]);
        setTotalComments((prev) => prev + 1);
    };

    const handleCommentDeleted = (commentId: string) => {
        setComments((prev) =>
            prev.map((c) =>
                c.id === commentId ? { ...c, content: "", isDeleted: true } : c,
            ),
        );
        setTotalComments((prev) => Math.max(prev - 1, 0));
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

    const handleLoadMore = () => {
        const nextPage = page + 1;
        setPage(nextPage);
        fetchComments(nextPage);
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

            {hasMore && comments.length > 0 && (
                <div className="flex justify-center">
                    <Button
                        variant="ghost"
                        onClick={handleLoadMore}
                        disabled={isLoading}
                    >
                        {isLoading
                            ? t("common:loading")
                            : t("common:show_more")}
                    </Button>
                </div>
            )}
        </div>
    );
}
