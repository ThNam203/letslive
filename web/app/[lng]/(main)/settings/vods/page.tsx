"use client";

import useUser from "@/hooks/user";
import VODEditCard from "./vod";
import useT from "@/hooks/use-translation";
import { useAuthorVods } from "@/hooks/queries/use-vods";

export default function VODsEdit() {
    const { t } = useT(["settings", "api-response", "fetch-error"]);
    const user = useUser((state) => state.user);
    const { data: vods } = useAuthorVods(Boolean(user));

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
                {(vods ?? []).map((vod) => (
                    <VODEditCard key={vod.id} vod={vod} />
                ))}
            </div>
        </>
    );
}
