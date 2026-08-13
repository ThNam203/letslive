"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { CreateVODComment } from "@/lib/api/vod-comment";
import { unwrapResponse } from "@/lib/api/api-error";
import { VODComment } from "@/types/vod-comment";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { VOD_COMMENT_MAX_LENGTH } from "@/constant/field-limits";

interface CommentFormProps {
    vodId: string;
    parentId?: string;
    onCommentCreated: (comment: VODComment) => void;
    onCancel?: () => void;
    placeholder?: string;
    autoFocus?: boolean;
}

export default function CommentForm({
    vodId,
    parentId,
    onCommentCreated,
    onCancel,
    placeholder,
    autoFocus = false,
}: CommentFormProps) {
    const { t } = useT(["comments", "common"]);
    const user = useUser((state) => state.user);
    const [content, setContent] = useState("");

    const createComment = useMutation({
        mutationFn: async () =>
            unwrapResponse(
                await CreateVODComment(vodId, {
                    content: content.trim(),
                    parentId,
                }),
            ),
        onSuccess: (comment) => {
            setContent("");
            onCommentCreated(comment);
        },
    });

    const handleSubmit = () => {
        if (!content.trim() || createComment.isPending) return;
        createComment.mutate();
    };

    return (
        <div className="flex items-start gap-3">
            <Avatar className="h-8 w-8 flex-shrink-0">
                <AvatarImage
                    src={user?.profilePicture}
                    alt={user?.username ?? undefined}
                />
                <AvatarFallback>
                    {user?.username?.charAt(0).toUpperCase()}
                </AvatarFallback>
            </Avatar>
            <div className="flex-1 space-y-2">
                <Textarea
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    placeholder={placeholder ?? t("comments:write_comment")}
                    className="min-h-[60px] resize-none"
                    autoFocus={autoFocus}
                    maxLength={VOD_COMMENT_MAX_LENGTH}
                    aria-label={
                        parentId
                            ? t("comments:write_reply")
                            : t("comments:write_comment")
                    }
                />
                <div className="flex justify-between gap-2">
                    {content.length > 0 ? (
                        <span className="text-muted-foreground self-center text-xs">
                            {t("comments:char_remaining", {
                                count: VOD_COMMENT_MAX_LENGTH - content.length,
                            })}
                        </span>
                    ) : (
                        <span />
                    )}
                    <div className="flex gap-2">
                        {onCancel && (
                            <Button
                                variant="destructive"
                                size="sm"
                                onClick={onCancel}
                            >
                                {t("common:cancel")}
                            </Button>
                        )}
                        <Button
                            size="sm"
                            onClick={handleSubmit}
                            disabled={!content.trim() || createComment.isPending}
                        >
                            {createComment.isPending
                                ? t("common:loading")
                                : t("comments:post")}
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    );
}
