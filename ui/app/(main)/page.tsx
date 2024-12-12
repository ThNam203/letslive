"use client";

import { CustomLink } from "@/components/Hover3DBox";
import { RecommendStreamView } from "@/components/LivesteamView";
import Separator from "@/components/Separator";
import { useEffect, useState } from "react";
import { LuChevronDown } from "react-icons/lu";

export interface OnlineStream {
    id: string //user id
    username: string
    email: string
    isVerified: boolean
    isOnline: boolean
    createdAt: string
    streamAPIKey: string
  }

export default function HomePage() {
    const [limitView, setLimitView] = useState<number[]>([4, 4, 4]);
    const [streams, setStreams] = useState<OnlineStream[]>([]);

    const handleShowMoreBtn = (index: number) => {
        const newLimitView = [...limitView];
        newLimitView[index] += 8;
        setLimitView(newLimitView);
    };

    useEffect(() => {
        const getOnlineStreams = async () => {
            fetch("http://localhost:8000/user?isOnline=true")
                .then((res) => res.json())
                .then((data) => {
                    setStreams(data);
                });
            }

            getOnlineStreams();
    }, []);

    return (
        <div className="flex flex-col w-full max-h-full p-8 overflow-y-scroll overflow-x-hidden">
            <RecommendStreamView
                title={
                    <span>
                        <CustomLink content="Live channels" href="" /> we think
                        you&#39;ll like
                    </span>
                }
                streams={streams}
                limitView={limitView[0]}
                separate={
                    <div className="w-full flex flex-row items-center justify-between gap-4">
                        <Separator />
                        <button
                            className="px-2 py-1 hover:bg-hoverColor hover:text-primaryWord rounded-md text-xs font-semibold text-primary flex flex-row items-center justify-center text-nowrap ease-linear duration-100"
                            onClick={() => handleShowMoreBtn(0)}
                        >
                            <span className="">Show more</span>
                            <LuChevronDown />
                        </button>
                        <Separator />
                    </div>
                }
            />
        </div>
    );
}
