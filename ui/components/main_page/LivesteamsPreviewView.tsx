"use client";

import { useEffect, useState } from "react";
import { LuChevronDown } from "react-icons/lu";
import { Play } from "lucide-react";
import { toast } from "react-toastify";
import { GetOnlineUsers } from "../../lib/api/user";
import LivestreamPreviewView from "./LivestreamPreviewView";
import { User } from "../../types/user";
import Separator from "../Separator";

const LivestreamsPreviewView = () => {
    const [limitView, setLimitView] = useState<number>(4);

    useEffect(() => {
        const fetchOnlineUsers = async () => {
            const { users, fetchError } = await GetOnlineUsers();
            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "online-users-fetch-error",
                    type: "error",
                });
            }

            setUsers(users ?? []);
        };

        fetchOnlineUsers();
    }, []);

    const [users, setUsers] = useState<User[]>([]);

    return (
        <div className="flex flex-col gap-2 pr-2">
            {users.slice(0, 4).map((user, idx) => (
                <LivestreamPreviewView
                    key={idx}
                    viewers={123}
                    title={user.livestreamInformation.title!}
                    tags={[]}
                    category={""}
                    user={user}
                />
            ))}
            {users.length > limitView && (
                <StreamsSeparator
                    onClick={() => setLimitView((prev) => prev + 8)}
                />
            )}
            {users.length == 0 && (
                <div className="flex flex-col items-center justify-center py-16 px-4 text-center">
                <div className="bg-muted/30 p-6 rounded-full mb-6">
                  <Play className="h-12 w-12 text-muted-foreground" />
                </div>
                <h2 className="text-2xl font-semibold mb-2">No Live Streams</h2>
                <p className="text-muted-foreground max-w-md">
                  There is currently no one streaming. Check back later or explore our video on demand content.
                </p>
              </div>
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