"use client";

import { useState } from "react";
import { CreateVODComment } from "@/lib/api/vod-comment";
import { VODComment } from "@/types/vod-comment";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

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
    const { t } = useT(["comments", "common", "fetch-error", "api-response"]);
    const user = useUser((state) => state.user);
    const [content, setContent] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleSubmit = async () => {
        if (!content.trim() || isSubmitting) return;

        setIsSubmitting(true);
        try {
            const res = await CreateVODComment(vodId, {
                content: content.trim(),
                parentId,
            });
            if (res.success && res.data) {
                setContent("");
                onCommentCreated(res.data);
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
            setIsSubmitting(false);
        }
    };

    return (
        <div className="flex items-start gap-3">
            <Avatar className="h-8 w-8 flex-shrink-0">
                <AvatarImage
                    src={user?.profilePicture}
                    alt={user?.username}
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
                    maxLength={2000}
                    aria-label={parentId ? t("comments:write_reply") : t("comments:write_comment")}
                />
                <div className="flex justify-between gap-2">
                    {content.length > 0 ? (
                        <span className="text-xs text-muted-foreground self-center">
                            {t("comments:char_remaining", { count: 2000 - content.length })}
                        </span>
                    ) : (
                        <span />
                    )}
                    {onCancel && (
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={onCancel}
                        >
                            {t("common:cancel")}
                        </Button>
                    )}
                    <Button
                        size="sm"
                        onClick={handleSubmit}
                        disabled={!content.trim() || isSubmitting}
                    >
                        {isSubmitting
                            ? t("common:loading")
                            : t("comments:post")}
                    </Button>
                </div>
            </div>
        </div>
    );
}
