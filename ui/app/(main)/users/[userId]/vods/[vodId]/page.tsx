"use client";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { User } from "../../../../../../types/user";
import { VideoInfo } from "../../../../../../components/custom_react_player/streaming_frame";
import { GetVODInformation } from "../../../../../../lib/api/vod";
import { GetUserById } from "../../../../../../lib/api/user";
import { VODFrame } from "../../../../../../components/custom_react_player/vod_frame";
import ProfileView from "../../profile";
import VODLink from "../../../../../../components/vodlink";
import { ScrollArea } from "../../../../../../components/ui/scroll-area";
import GLOBAL from "../../../../../../global";

export default function VODPage() {
    const [user, setUser] = useState<User | null>(null);
    const updateUser = (newUserInfo: User) => {
        setUser((prev) => {
            if (prev)
                return {
                    ...prev,
                    ...newUserInfo,
                };

            return newUserInfo;
        });
    };
    const params = useParams<{ userId: string; vodId: string }>();

    const [playerInfo, setPlayerInfo] = useState<VideoInfo>({
        videoTitle: "Live Streaming",
        streamer: {
            name: "Streamer",
        },
        videoUrl: null,
    });

    useEffect(() => {
        const fetchVODInfo = async () => {
            const { vodInfo, fetchError } = await GetVODInformation(
                params.vodId
            );

            if (fetchError) {
                toast.error(fetchError.message, {
                    toastId: "vod-fetch-error",
                });
                return;
            }

            const newUrl = `${GLOBAL.API_URL}/transcode/vods/${params.vodId}/index.m3u8`;
            setPlayerInfo((prev) => ({
                ...prev,
                videoTitle: vodInfo!.title,
                videoUrl: newUrl,
                streamer: {
                    name: user?.displayName ?? user?.username ?? "Streamer",
                }
            }));
        };

        fetchVODInfo();
    }, [params.userId, params.vodId, user]);

    useEffect(() => {
        const fetchUserInfo = async () => {
            const { user, fetchError } = await GetUserById(params.userId);
            if (fetchError != undefined) {
                toast.error(fetchError.message, {
                    toastId: "user-fetch-error",
                });
            } else {
                setUser(user!);
            }
        };

        fetchUserInfo();
    }, [params.userId]);

    return (
        <div className="overflow-y-auto h-full">
            <div className="flex lg:flex-row max-lg:flex-col">
                <div className="w-[1200px] min-w-[1200px]">
                    <div className="w-full h-[675px] bg-black mb-4">
                        <VODFrame videoInfo={playerInfo} />
                    </div>
                    {user && (
                        <ProfileView
                            user={user}
                            showSavedStream={false}
                            updateUser={updateUser}
                        />
                    )}
                </div>

                <div className="w-[400px] mx-4 fixed right-0 top-14 bottom-4 overflow-hidden">
                    <div className="w-full h-full font-sans border border-gray-200 rounded-md bg-gray-50 p-4">
                        <h2 className="text-xl mb-4 font-semibold">
                            Other streams
                        </h2>
                        <div className="overflow-y-auto h-full pr-1">
                            {user &&
                                user.vods
                                    ?.filter(
                                        (v) =>
                                            v.id !== params.vodId &&
                                            v.status !== "live"
                                    )
                                    .map((vod, idx) => (
                                        <VODLink key={idx} item={vod} classname="mb-2" />
                                    ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

// const ServersChooser = () => {
//     return (
//         <div className="w-full font-sans mt-4 overflow-x-auto whitespace-nowrap">
//             {['1','2','3'].map((_, idx) => (
//                 <Button
//                     key={idx}
//                     onClick={() => setServerIndex(idx)}
//                     className={cn(
//                         "mr-4",
//                         serverIndex == idx ? "bg-green-700" : ""
//                     )}
//                 >
//                     Server {idx + 1}
//                 </Button>
//             ))}
//         </div>
//     );
// };
