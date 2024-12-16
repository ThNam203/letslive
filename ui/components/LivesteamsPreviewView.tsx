"use client";

import { useState } from "react";
import { LuChevronDown } from "react-icons/lu";
import { User } from "@/types/user";
import Separator from "@/components/Separator";
import LivestreamPreviewView from "@/components/LivestreamPreviewView";

const LivestreamsPreviewView = ({ users }: { users: User[] }) => {
    const [limitView, setLimitView] = useState<number>(4);
    const visibileUsers = users.slice(0, limitView);

    return (
        <div className="flex flex-col gap-2 mt-8 pr-2">
            {visibileUsers.map((user, idx) => (
                <LivestreamPreviewView
                    key={idx}
                    viewers={123}
                    title={"A livestream ???? what else"}
                    tags={["boobies", "titties"]}
                    category={"porn"}
                    stream={user}
                />
            ))}
            {users.length > limitView && (
                <StreamsSeparator
                    onClick={() => setLimitView((prev) => prev + 8)}
                />
            )}
            {users.length == 0 && (
                <p className="text-lg text-center">
                    There is currently no one streaming
                </p>
            )}
        </div>
    );
};

const StreamsSeparator = ({ onClick }: { onClick: () => void }) => {
    return (
        <div className="w-full flex flex-row items-center justify-between gap-4">
            <Separator />
            <button
                className="px-2 py-1 hover:bg-hoverColor hover:text-primaryWord rounded-md text-xs font-semibold text-primary flex flex-row items-center justify-center text-nowrap ease-linear duration-100"
                onClick={onClick}
            >
                <span className="">Show more</span>
                <LuChevronDown />
            </button>
            <Separator />
        </div>
    );
};

export default LivestreamsPreviewView;