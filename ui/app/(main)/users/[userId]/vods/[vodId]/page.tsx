"use client";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { User } from "../../../../../../types/user";
import { VideoInfo } from "../../../../../../components/custom_react_player/streaming_frame";
import { GetAllLivestreamOfUser, GetVODInformation } from "../../../../../../lib/api/livestream";
import { GetUserById } from "../../../../../../lib/api/user";
import { VODFrame } from "../../../../../../components/custom_react_player/vod_frame";
import ProfileView from "../../profile";
import VODLink from "../../../../../../components/vodlink";
import GLOBAL from "../../../../../../global";
import { Livestream } from "../../../../../../types/livestream";

export default function VODPage() {
    const [user, setUser] = useState<User | null>(null);
    const [vods, setVods] = useState<Livestream[]>([]);
        
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
            const { vod, fetchError } = await GetVODInformation(
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
                videoTitle: vod!.title,
                videoUrl: newUrl,
                streamer: {
                    name: user?.displayName ?? user?.username ?? "Streamer",
                }
            }));
        };

        fetchVODInfo();
    }, [params.userId, params.vodId, user]);

    useEffect(() => {
        if (!user) {
            return;
        }

        const fetchVODs = async () => {
            const { livestreams, fetchError } = await GetAllLivestreamOfUser(user.id);

            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "vod-fetch-error",
                    type: "error",
                });
            } else {
                setVods(livestreams);
            }
        };

        fetchVODs();
    }, [user]);

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
        <div className="overflow-y-auto h-full mt-2">
            <div className="flex lg:flex-row max-lg:flex-col">
                <div className="w-[910px]">
                    <div className="w-full aspect-video bg-black mb-4">
                        <VODFrame videoInfo={playerInfo} />
                    </div>
                    {user && (
                        <ProfileView
                            user={user}
                            updateUser={updateUser}
                            vods={vods.filter((v) => v.id !== params.vodId)}
                            showRecentActivity={false}
                        />
                    )}
                </div>

                <div className="w-[300px] mx-4 fixed right-0 top-14 bottom-4 overflow-hidden">
                    <div className="w-full h-full font-sans border border-gray-200 rounded-md bg-gray-50 p-4">
                        <h2 className="text-xl mb-4 font-semibold">
                            Other streams
                        </h2>
                        <div className="overflow-y-auto h-full pr-1">
                            {vods
                                    ?.filter(
                                        (v) =>
                                            v.id !== params.vodId &&
                                            v.status !== "live"
                                    )
                                    .map((vod, idx) => (
                                        <VODLink key={idx} vod={vod} classname="mb-2" />
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
