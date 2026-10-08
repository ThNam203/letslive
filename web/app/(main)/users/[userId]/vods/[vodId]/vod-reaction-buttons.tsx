"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { VOD, VODReaction, VODReactionState } from "@/types/vod";
import { RemoveVODReaction, SetVODReaction } from "@/lib/api/vod";
import { unwrapResponse } from "@/lib/api/api-error";
import {
    myVodReactionQueryKey,
    useMyVodReaction,
    vodQueryKey,
} from "@/hooks/queries/use-vods";
import useT from "@/hooks/use-translation";
import useCompactNumber from "@/hooks/use-compact-number";
import useUser from "@/hooks/user";
import { toast } from "@/components/utils/toast";
import IconThumbsUp from "@/components/icons/thumbs-up";
import IconThumbsDown from "@/components/icons/thumbs-down";

function likeDelta(from: VODReaction | null, to: VODReaction | null) {
    return (to === "like" ? 1 : 0) - (from === "like" ? 1 : 0);
}

export default function VODReactionButtons({ vod }: { vod: VOD }) {
    const { t } = useT("common");
    const formatCompact = useCompactNumber();
    const queryClient = useQueryClient();
    const user = useUser((state) => state.user);
    const reactionKey = myVodReactionQueryKey(vod.id, user?.id ?? "");
    const { data: myReaction } = useMyVodReaction(
        vod.id,
        user?.id,
        Boolean(user),
    );

    const reaction = user ? (myReaction?.reaction ?? null) : null;
    const likeCount = myReaction?.likeCount ?? vod.likeCount;

    const reactionMutation = useMutation({
        mutationFn: async (next: VODReaction | null) =>
            unwrapResponse(
                next
                    ? await SetVODReaction(vod.id, next)
                    : await RemoveVODReaction(vod.id),
            ),
        // flip the buttons right away; the server answer replaces this
        onMutate: async (next) => {
            await queryClient.cancelQueries({ queryKey: reactionKey });
            const previous =
                queryClient.getQueryData<VODReactionState>(reactionKey);
            queryClient.setQueryData<VODReactionState>(reactionKey, {
                likeCount: Math.max(likeCount + likeDelta(reaction, next), 0),
                reaction: next,
            });
            return { previous };
        },
        onError: (_err, _next, context) => {
            queryClient.setQueryData(reactionKey, context?.previous);
        },
        onSuccess: (data) => {
            queryClient.setQueryData(reactionKey, data);
            queryClient.setQueryData<VOD>(vodQueryKey(vod.id), (prev) =>
                prev ? { ...prev, likeCount: data.likeCount } : prev,
            );
        },
    });

    const handleReact = (clicked: VODReaction) => {
        if (!user) {
            toast(t("common:vod.login_to_react"), {
                toastId: "vod-reaction-login",
                type: "info",
            });
            return;
        }
        if (reactionMutation.isPending) return;
        reactionMutation.mutate(reaction === clicked ? null : clicked);
    };

    return (
        <div className="bg-muted text-foreground flex items-center overflow-hidden rounded-full text-sm font-medium">
            <button
                type="button"
                onClick={() => handleReact("like")}
                aria-pressed={reaction === "like"}
                aria-label={t(
                    reaction === "like"
                        ? "common:vod.remove_like"
                        : "common:vod.like",
                )}
                className="hover:bg-foreground/10 flex items-center gap-2 py-2 pr-4 pl-4 transition-colors"
            >
                <IconThumbsUp
                    filled={reaction === "like"}
                    width="1.1rem"
                    height="1.1rem"
                />
                <span>{formatCompact(likeCount)}</span>
            </button>
            <span className="bg-border h-6 w-px" aria-hidden />
            <button
                type="button"
                onClick={() => handleReact("dislike")}
                aria-pressed={reaction === "dislike"}
                aria-label={t(
                    reaction === "dislike"
                        ? "common:vod.remove_dislike"
                        : "common:vod.dislike",
                )}
                className="hover:bg-foreground/10 flex items-center py-2 pr-4 pl-3 transition-colors"
            >
                <IconThumbsDown
                    filled={reaction === "dislike"}
                    width="1.1rem"
                    height="1.1rem"
                />
            </button>
        </div>
    );
}
