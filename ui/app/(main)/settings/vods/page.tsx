"use client";

import useUser from "../../../../hooks/user";
import VODLink from "../../../../components/vodlink";
import { useEffect, useState } from "react";
import { GetAllVODsAsAuthor } from "../../../../lib/api/livestream";
import { toast } from "react-toastify";
import { Livestream } from "../../../../types/livestream";
import VODEditCard from "./vod";

export default function VODsEdit() {
    const user = useUser((state) => state.user);
    const [vods, setVODS] = useState<Livestream[]>([]);
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
        <div className="min-h-screen text-gray-900 p-6">
            <div className="space-y-6 mb-4 max-w-4xl">
                <div className="space-y-1">
                    <h1 className="text-xl font-semibold">VODs Manager</h1>
                    <p className="text-sm text-gray-400">
                        Manage your VODs and edit their information here.
                    </p>
                </div>
            </div>

            <div className="flex flex-row flex-wrap gap-4">
                {vods.map((vod) => {
                    return (
                        <VODEditCard key={vod.id} vod={vod} setVODS={setVODS} />
                    );
                })}
            </div>
        </div>
    );
}
