"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { UpdateVODComment } from "@/lib/api/vod-comment";
import { unwrapResponse } from "@/lib/api/api-error";
import { VODComment } from "@/types/vod-comment";
import useT from "@/hooks/use-translation";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { VOD_COMMENT_MAX_LENGTH } from "@/constant/field-limits";

interface CommentEditFormProps {
    comment: VODComment;
    onCommentUpdated: (updated: VODComment) => void;
    onCancel: () => void;
}

export default function CommentEditForm({
    comment,
    onCommentUpdated,
    onCancel,
}: CommentEditFormProps) {
    const { t } = useT(["comments", "common"]);
    const [content, setContent] = useState(comment.content);

    const updateComment = useMutation({
        mutationFn: async () =>
            unwrapResponse(
                await UpdateVODComment(comment.id, { content: content.trim() }),
            ),
        onSuccess: (updated) => onCommentUpdated(updated),
    });

    const trimmed = content.trim();
    const isUnchanged = trimmed === comment.content;

    const handleSubmit = () => {
        if (!trimmed || updateComment.isPending) return;
        // nothing to save: close the editor instead of writing the same text
        if (isUnchanged) {
            onCancel();
            return;
        }
        updateComment.mutate();
    };

    return (
        <div className="mt-1 space-y-2">
            <Textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                className="min-h-[60px] resize-none"
                autoFocus
                maxLength={VOD_COMMENT_MAX_LENGTH}
                aria-label={t("comments:edit_comment")}
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
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={onCancel}
                        disabled={updateComment.isPending}
                    >
                        {t("common:cancel")}
                    </Button>
                    <Button
                        size="sm"
                        onClick={handleSubmit}
                        disabled={!trimmed || updateComment.isPending}
                    >
                        {updateComment.isPending
                            ? t("common:loading")
                            : t("comments:save")}
                    </Button>
                </div>
            </div>
        </div>
    );
}
