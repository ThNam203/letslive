"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import useUser from "../../hooks/user";
import { User } from "../../types/user";
import { GetAllUsers } from "../../lib/api/user";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";

export default function AllChannelsView() {
    const curUser = useUser(state => state.user);
    const [users, setUsers] = useState<User[]>([]);
    useEffect(() => {
        const fetchAllUsers = async () => {
            const { users, fetchError } = await GetAllUsers();

            if (fetchError != undefined) {
                // toast.error(fetchError.message, {
                //     toastId: "all-channels-fetch-error",
                // });
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
                    <Avatar>
                        <AvatarImage
                            src={user.profilePicture}
                            alt="avatar"
                            className="bg-black rounded-full max-h-[40px] max-w-[40px]"
                        />
                        <AvatarFallback>
                            {user.username.charAt(0).toUpperCase()}
                        </AvatarFallback>
                    </Avatar>
                    <span className="text-sm font-semibold">{user.displayName ?? user.username}</span>
                </Link>
            })}
        </div>
    );
}
