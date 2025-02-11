"use client";

import { GetAllUsers } from "@/lib/api/user";
import { User } from "@/types/user";
import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";

export default function AllChannelsView() {
    const [users, setUsers] = useState<User[]>([]);
    useEffect(() => {
        const fetchAllUsers = async () => {
            const { users, fetchError } = await GetAllUsers();

            if (fetchError != undefined) {
                toast.error(fetchError.message, {
                    toastId: "all-channels-fetch-error",
                });
            } else {
                setUsers(users ?? []);
            }
        };

        fetchAllUsers();
    }, []);

    return (
        <div className="flex flex-col gap-2 w-full">
            <h2 className="font-semibold text-xl max-xl:hidden">Channels</h2>
            {users.map((user, idx) => (
                <Link
                    key={user.id}
                    href={`/users/${user.id}`}
                    className="flex flex-row items-center gap-2 hover:bg-gray-300 rounded-full w-full"
                >
                    <Image
                        alt="channel avatar"
                        src={"https://github.com/shadcn.png"}
                        width={40}
                        height={40}
                        className="bg-black rounded-full max-h-[40px] max-w-[40px]"
                    />
                    <span className="text-sm font-semibold">{user.username}</span>
                </Link>
            ))}
        </div>
    );
}
