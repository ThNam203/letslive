"use client";

import { CustomLink } from "@/components/Hover3DBox";
import LivestreamsPreviewView from "@/components/LivesteamsPreviewView";
import { GetOnlineUsers } from "@/lib/api/user";
import { User } from "@/types/user";
import { use, useEffect, useState } from "react";
import { toast } from "react-toastify";

export default function HomePage() {
    useEffect(() => {
        const fetchOnlineUsers = async () => {
            const { users, fetchError } = await GetOnlineUsers();
            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "online-users-fetch-error",
                });
            }

            setUsers(users ?? []);
        };

        fetchOnlineUsers();
    }, []);

    const [users, setUsers] = useState<User[]>([]);

    return (
        <>
            {/* {fetchError != undefined && (
                <ShowToast id={fetchError.id} err={fetchError.message} />
            )} */}
            <div className="flex flex-col w-full max-h-full p-8 overflow-y-scroll overflow-x-hidden">
                <h1 className="font-semibold text-lg">
                    <CustomLink content="Live channels" href="" /> we think
                    you&#39;ll like
                </h1>

                <div className="w-full flex flex-row items-center justify-between gap-4">
                    <LivestreamsPreviewView users={users} />
                </div>
            </div>
        </>
    );
}
