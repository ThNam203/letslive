"use client";

import { useState, useCallback, useEffect } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { CommentUser, VODComment } from "@/types/vod-comment";
import {
    GetCommentReplies,
    GetUserLikedCommentIds,
    LikeVODComment,
    UnlikeVODComment,
    DeleteVODComment,
} from "@/lib/api/vod-comment";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { dateDiffFromNow } from "@/utils/timeFormats";
import IconHeart from "@/components/icons/heart";
import IconHeartFilled from "@/components/icons/heart-filled";
import IconReply from "@/components/icons/reply";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogClose,
} from "@/components/ui/dialog";
import CommentForm from "./comment-form";
import CommentList from "./comment-list";

interface CommentItemProps {
    comment: VODComment;
    vodId: string;
    vodOwnerId?: string;
    likedIds?: Set<string>;
    onCommentDeleted: (commentId: string) => void;
    onLikedChanged?: (commentId: string, liked: boolean) => void;
    depth?: number;
}

export default function CommentItem({
    comment,
    vodId,
    vodOwnerId,
    likedIds,
    onCommentDeleted,
    onLikedChanged,
    depth = 0,
}: CommentItemProps) {
    const params = useParams();
    const lng = params?.lng as string | undefined;
    const { t } = useT(["comments", "common", "fetch-error", "api-response"]);
    const currentUser = useUser((state) => state.user);
    const commentUser = comment.user ?? null;
    const userProfileHref =
        lng && commentUser?.id ? `/${lng}/users/${commentUser.id}` : "#";
    const [isLiked, setIsLiked] = useState(likedIds?.has(comment.id) ?? false);
    const [likeCount, setLikeCount] = useState(comment.likeCount);

    useEffect(() => {
        setIsLiked(likedIds?.has(comment.id) ?? false);
    }, [likedIds, comment.id]);
    const [showReplyForm, setShowReplyForm] = useState(false);
    const [replies, setReplies] = useState<VODComment[]>([]);
    const [showReplies, setShowReplies] = useState(false);
    const [replyCount, setReplyCount] = useState(comment.replyCount);
    const [replyPage, setReplyPage] = useState(0);
    const [hasMoreReplies, setHasMoreReplies] = useState(
        comment.replyCount > 0,
    );
    const [replyLikedIds, setReplyLikedIds] = useState<Set<string>>(new Set());
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [isLiking, setIsLiking] = useState(false);
    const [isLoadingReplies, setIsLoadingReplies] = useState(false);

    const handleLike = async () => {
        if (!currentUser || isLiking) return;
        setIsLiking(true);
        try {
            if (isLiked) {
                const res = await UnlikeVODComment(comment.id);
                if (res.success) {
                    setIsLiked(false);
                    setLikeCount((prev) => Math.max(prev - 1, 0));
                    onLikedChanged?.(comment.id, false);
                } else {
                    toast(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                }
            } else {
                const res = await LikeVODComment(comment.id);
                if (res.success) {
                    setIsLiked(true);
                    setLikeCount((prev) => prev + 1);
                    onLikedChanged?.(comment.id, true);
                } else {
                    toast(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                }
            }
        } catch (_) {
            toast(t("fetch-error:client_fetch_error"), {
                toastId: "client-fetch-error-id",
                type: "error",
            });
        } finally {
            setIsLiking(false);
        }
    };

    const handleDelete = async () => {
        setDeleteDialogOpen(false);
        try {
            const res = await DeleteVODComment(comment.id);
            if (res.success) {
                onCommentDeleted(comment.id);
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
        }
    };

    const fetchReplyLikedIds = useCallback(
        async (replyList: VODComment[]) => {
            if (!currentUser) return;
            const ids = replyList.filter((c) => !c.isDeleted).map((c) => c.id);
            if (ids.length === 0) return;
            try {
                const res = await GetUserLikedCommentIds(ids);
                if (res.success && res.data) {
                    setReplyLikedIds((prev) => {
                        const next = new Set(prev);
                        for (const id of res.data!) next.add(id);
                        return next;
                    });
                }
            } catch (_) {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "reply-liked-ids-fetch-error",
                    type: "error",
                });
            }
        },
        [currentUser, t],
    );

    const handleLoadReplies = async () => {
        if (isLoadingReplies) return;
        setIsLoadingReplies(true);
        try {
            const res = await GetCommentReplies(comment.id, replyPage, 20);
            if (res.success) {
                const newReplies = res.data ?? [];
                setReplies((prev) => [...prev, ...newReplies]);
                setShowReplies(true);
                setHasMoreReplies(newReplies.length === 20);
                setReplyPage((prev) => prev + 1);
                fetchReplyLikedIds(newReplies);
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
            setIsLoadingReplies(false);
        }
    };

    const handleReplyCreated = (newReply: VODComment) => {
        const replyWithUser: VODComment =
            !newReply.user && currentUser && currentUser.id === newReply.userId
                ? {
                      ...newReply,
                      user: {
                          id: currentUser.id,
                          username: currentUser.username,
                          displayName: currentUser.displayName,
                          profilePicture: currentUser.profilePicture,
                      } satisfies CommentUser,
                  }
                : newReply;
        setReplies((prev) => [...prev, replyWithUser]);
        setReplyCount((prev) => prev + 1);
        setShowReplies(true);
        setShowReplyForm(false);
    };

    const handleReplyDeleted = (replyId: string) => {
        setReplies((prev) =>
            prev.map((r) =>
                r.id === replyId ? { ...r, content: "", isDeleted: true } : r,
            ),
        );
        setReplyCount((prev) => Math.max(prev - 1, 0));
    };

    const handleReplyLikedChanged = (commentId: string, liked: boolean) => {
        setReplyLikedIds((prev) => {
            const next = new Set(prev);
            if (liked) {
                next.add(commentId);
            } else {
                next.delete(commentId);
            }
            return next;
        });
    };

    const isAuthor = currentUser?.id === comment.userId;
    const isOwner = Boolean(vodOwnerId && comment.userId === vodOwnerId);
    const isYou = isAuthor;

    if (comment.isDeleted) {
        return (
            <div className="flex items-start gap-3 opacity-60">
                <Avatar className="h-8 w-8 shrink-0">
                    <AvatarFallback>?</AvatarFallback>
                </Avatar>
                <div className="flex-1">
                    <p className="text-muted-foreground text-sm italic">
                        {t("comments:deleted_comment")}
                    </p>
                    {replyCount > 0 && !showReplies && (
                        <Button
                            variant="link"
                            size="sm"
                            className="mt-1 h-auto cursor-pointer p-0 text-xs"
                            onClick={handleLoadReplies}
                            disabled={isLoadingReplies}
                        >
                            {isLoadingReplies
                                ? t("common:loading")
                                : t("comments:view_replies", {
                                      count: replyCount,
                                  })}
                        </Button>
                    )}
                    {showReplies && replies.length > 0 && (
                        <CommentList
                            comments={replies}
                            vodId={vodId}
                            vodOwnerId={vodOwnerId}
                            likedIds={replyLikedIds}
                            onCommentDeleted={handleReplyDeleted}
                            onLikedChanged={handleReplyLikedChanged}
                            isReplyList
                            depth={depth + 1}
                        />
                    )}
                    {showReplies && hasMoreReplies && (
                        <Button
                            variant="link"
                            size="sm"
                            className="mt-1 h-auto cursor-pointer p-0 text-xs"
                            onClick={handleLoadReplies}
                            disabled={isLoadingReplies}
                        >
                            {isLoadingReplies
                                ? t("common:loading")
                                : t("comments:load_more_replies")}
                        </Button>
                    )}
                </div>
            </div>
        );
    }

    return (
        <div className="flex items-start gap-3">
            {userProfileHref !== "#" ? (
                <Link
                    href={userProfileHref}
                    className="ring-offset-background focus-visible:ring-ring shrink-0 cursor-pointer rounded-full focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
                    aria-label={commentUser?.username ?? "View profile"}
                >
                    <Avatar className="h-8 w-8">
                        <AvatarImage
                            src={commentUser?.profilePicture}
                            alt={commentUser?.username}
                        />
                        <AvatarFallback>
                            {commentUser?.username?.charAt(0).toUpperCase() ??
                                "?"}
                        </AvatarFallback>
                    </Avatar>
                </Link>
            ) : (
                <Avatar className="h-8 w-8 shrink-0">
                    <AvatarImage
                        src={commentUser?.profilePicture}
                        alt={commentUser?.username}
                    />
                    <AvatarFallback>
                        {commentUser?.username?.charAt(0).toUpperCase() ?? "?"}
                    </AvatarFallback>
                </Avatar>
            )}
            <div className="flex-1">
                <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                    {userProfileHref !== "#" ? (
                        <Link
                            href={userProfileHref}
                            className="cursor-pointer text-sm font-semibold hover:underline"
                        >
                            {commentUser?.displayName ??
                                commentUser?.username ??
                                "..."}
                        </Link>
                    ) : (
                        <span className="text-sm font-semibold">
                            {commentUser?.displayName ??
                                commentUser?.username ??
                                "..."}
                        </span>
                    )}
                    {(isOwner || isYou) && (
                        <span className="flex items-center gap-1">
                            {isOwner && (
                                <span className="rounded bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium tracking-wide text-amber-600 dark:text-amber-400">
                                    {t("comments:owner")}
                                </span>
                            )}
                            {isYou && (
                                <span className="bg-primary/15 text-primary rounded px-1.5 py-0.5 text-[10px] font-medium">
                                    {t("comments:you")}
                                </span>
                            )}
                        </span>
                    )}
                    <span className="text-muted-foreground text-xs">
                        {dateDiffFromNow(comment.createdAt, t)}
                    </span>
                </div>
                <p className="mt-1 text-sm whitespace-pre-wrap">
                    {comment.content}
                </p>

                {/* Action buttons */}
                <div className="mt-1 flex items-center gap-3">
                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 cursor-pointer gap-1 px-2 text-xs"
                        onClick={handleLike}
                        disabled={!currentUser || isLiking}
                        aria-label={
                            isLiked ? t("comments:unlike") : t("comments:like")
                        }
                    >
                        {isLiking ? (
                            <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent" />
                        ) : isLiked ? (
                            <IconHeartFilled className="h-3.5 w-3.5 text-red-500" />
                        ) : (
                            <IconHeart className="h-3.5 w-3.5" />
                        )}
                        {likeCount > 0 && <span>{likeCount}</span>}
                    </Button>

                    {currentUser && (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 cursor-pointer gap-1 px-2 text-xs"
                            onClick={() => setShowReplyForm(!showReplyForm)}
                        >
                            <IconReply className="h-3.5 w-3.5" />
                            {t("comments:reply")}
                        </Button>
                    )}

                    {isAuthor && (
                        <Dialog
                            open={deleteDialogOpen}
                            onOpenChange={setDeleteDialogOpen}
                        >
                            <DialogTrigger asChild>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="text-destructive hover:text-destructive h-7 cursor-pointer px-2 text-xs"
                                >
                                    {t("comments:delete")}
                                </Button>
                            </DialogTrigger>
                            <DialogContent className="max-w-sm">
                                <DialogHeader>
                                    <DialogTitle>
                                        {t("comments:delete_confirm_title")}
                                    </DialogTitle>
                                    <DialogDescription>
                                        {t(
                                            "comments:delete_confirm_description",
                                        )}
                                    </DialogDescription>
                                </DialogHeader>
                                <DialogFooter>
                                    <DialogClose asChild>
                                        <Button variant="outline" size="sm">
                                            {t("common:cancel")}
                                        </Button>
                                    </DialogClose>
                                    <Button
                                        variant="destructive"
                                        size="sm"
                                        onClick={handleDelete}
                                    >
                                        {t("comments:delete")}
                                    </Button>
                                </DialogFooter>
                            </DialogContent>
                        </Dialog>
                    )}
                </div>

                {/* Reply form */}
                {showReplyForm && (
                    <div className="mt-2">
                        <CommentForm
                            vodId={vodId}
                            parentId={comment.id}
                            onCommentCreated={handleReplyCreated}
                            onCancel={() => setShowReplyForm(false)}
                            placeholder={t("comments:write_reply")}
                            autoFocus
                        />
                    </div>
                )}

                {/* Replies */}
                {replyCount > 0 && !showReplies && (
                    <Button
                        variant="link"
                        size="sm"
                        className="mt-1 h-auto cursor-pointer p-0 text-xs"
                        onClick={handleLoadReplies}
                        disabled={isLoadingReplies}
                    >
                        {isLoadingReplies
                            ? t("common:loading")
                            : t("comments:view_replies", { count: replyCount })}
                    </Button>
                )}

                {showReplies && replies.length > 0 && (
                    <CommentList
                        comments={replies}
                        vodId={vodId}
                        vodOwnerId={vodOwnerId}
                        likedIds={replyLikedIds}
                        onCommentDeleted={handleReplyDeleted}
                        onLikedChanged={handleReplyLikedChanged}
                        isReplyList
                        depth={depth + 1}
                    />
                )}

                {showReplies && hasMoreReplies && (
                    <Button
                        variant="link"
                        size="sm"
                        className="mt-1 h-auto cursor-pointer p-0 text-xs"
                        onClick={handleLoadReplies}
                        disabled={isLoadingReplies}
                    >
                        {isLoadingReplies
                            ? t("common:loading")
                            : t("comments:load_more_replies")}
                    </Button>
                )}
            </div>
        </div>
    );
}
