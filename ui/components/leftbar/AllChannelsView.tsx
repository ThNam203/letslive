"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import useUser from "../../hooks/user";
import { User } from "../../types/user";
import { GetAllUsers } from "../../lib/api/user";

export default function AllChannelsView() {
    const curUser = useUser(state => state.user);
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
            <h2 className="font-semibold text-xl max-xl:hidden mb-2">Channels</h2>
            {users.map((user, idx) => {
                if (curUser && curUser.id === user?.id) return null;

                return <Link
                    key={user.id}
                    href={`/users/${user.id}`}
                    className="flex flex-row items-center gap-2 hover:bg-gray-300 rounded-full w-full"
                >
                    <Image
                        alt="channel avatar"
                        src={user.profilePicture ?? "https://github.com/shadcn.png"}
                        width={40}
                        height={40}
                        className="bg-black rounded-full max-h-[40px] max-w-[40px]"
                    />
                    <span className="text-sm font-semibold">{user.displayName ?? user.username}</span>
                </Link>
})}
        </div>
    );
}
