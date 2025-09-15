"use client";

import useUser from "@/hooks/user";
import { useEffect, useState } from "react";
import { GetAllVODsAsAuthor } from "@/lib/api/vod";
import { toast } from "react-toastify";
import VODEditCard from "./vod";
import { VOD } from "@/types/vod";
import useT from "@/hooks/use-translation";

export default function VODsEdit() {
    const { t } = useT("translation");
    const user = useUser((state) => state.user);
    const [vods, setVODS] = useState<VOD[]>([]);
    useEffect(() => {
        if (!user) {
            return;
        }

        const fetchVODs = async () => {
            const { vods, fetchError } = await GetAllVODsAsAuthor();

            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "vod-fetch-error",
                    type: "error",
                });
            } else {
                setVODS(vods);
            }
        };

        fetchVODs();
    }, [user]);

    return (
        <>
            <div className="mb-4">
                <div className="space-y-1">
                    <h1 className="text-xl font-semibold">{t("settings.vods.title")}</h1>
                    <p className="text-sm text-foreground-muted">{t("settings.vods.description")}</p>
                </div>
            </div>

            <div className="flex flex-row flex-wrap gap-4">
                {vods.map((vod) => {
                    return (
                        <VODEditCard key={vod.id} vod={vod} setVODS={setVODS} />
                    );
                })}
            </div>
        </>
    );
}
