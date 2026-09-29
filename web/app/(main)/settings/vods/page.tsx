"use client";

import useUser from "@/hooks/user";
import { MediaCardSkeleton } from "@/components/skeletons/media-card-skeleton";
import VODEditCard from "./vod";
import useT from "@/hooks/use-translation";
import { useAuthorVods } from "@/hooks/queries/use-vods";

const VOD_SKELETON_COUNT = 4;

export default function VODsEdit() {
    const { t } = useT(["settings", "api-response", "fetch-error"]);
    const user = useUser((state) => state.user);
    const { data: vods, isLoading } = useAuthorVods(Boolean(user));

    return (
        <>
            <div className="mb-4">
                <div className="space-y-1">
                    <h1 className="text-xl font-semibold">
                        {t("settings:vods.title")}
                    </h1>
                    <p className="text-foreground-muted text-sm">
                        {t("settings:vods.description")}
                    </p>
                </div>
            </div>

            <div className="flex flex-row flex-wrap gap-4">
                {isLoading
                    ? Array.from({ length: VOD_SKELETON_COUNT }, (_, i) => (
                          <div key={i} className="w-[350px]">
                              <MediaCardSkeleton withUser={false} />
                          </div>
                      ))
                    : (vods ?? []).map((vod) => (
                          <VODEditCard key={vod.id} vod={vod} />
                      ))}
            </div>
        </>
    );
}
