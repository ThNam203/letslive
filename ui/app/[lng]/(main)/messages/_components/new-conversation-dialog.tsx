"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import useUser from "@/hooks/user";
import useDmStore from "@/hooks/use-dm-store";
import { CreateConversation } from "@/lib/api/dm";
import { SearchUsersByUsername } from "@/lib/api/user";
import { PublicUser } from "@/types/user";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export default function NewConversationDialog({
    onClose,
}: {
    onClose: () => void;
}) {
    const router = useRouter();
    const user = useUser((state) => state.user);
    const { addConversation } = useDmStore();

    const [searchQuery, setSearchQuery] = useState("");
    const [searchResults, setSearchResults] = useState<PublicUser[]>([]);
    const [selectedUsers, setSelectedUsers] = useState<PublicUser[]>([]);
    const [isGroup, setIsGroup] = useState(false);
    const [groupName, setGroupName] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [isSearching, setIsSearching] = useState(false);

    const handleSearch = async (query: string) => {
        setSearchQuery(query);
        if (query.trim().length < 2) {
            setSearchResults([]);
            return;
        }

        setIsSearching(true);
        try {
            const res = await SearchUsersByUsername(query.trim());
            if (res.data) {
                // Filter out current user and already selected users
                const filtered = res.data.filter(
                    (u) =>
                        u.id !== user?.id &&
                        !selectedUsers.some((s) => s.id === u.id),
                );
                setSearchResults(filtered);
            }
        } finally {
            setIsSearching(false);
        }
    };

    const handleSelectUser = (selectedUser: PublicUser) => {
        if (isGroup) {
            setSelectedUsers([...selectedUsers, selectedUser]);
        } else {
            setSelectedUsers([selectedUser]);
        }
        setSearchQuery("");
        setSearchResults([]);
    };

    const handleRemoveUser = (userId: string) => {
        setSelectedUsers(selectedUsers.filter((u) => u.id !== userId));
    };

    const handleCreate = async () => {
        if (selectedUsers.length === 0 || !user) return;

        setIsLoading(true);
        try {
            const participantUsernames: Record<string, string> = {};
            const participantDisplayNames: Record<string, string> = {};
            const participantProfilePictures: Record<string, string> = {};

            for (const u of selectedUsers) {
                participantUsernames[u.id] = u.username;
                if (u.displayName)
                    participantDisplayNames[u.id] = u.displayName;
                if (u.profilePicture)
                    participantProfilePictures[u.id] = u.profilePicture;
            }

            const res = await CreateConversation({
                type: isGroup ? "group" : "dm",
                participantIds: selectedUsers.map((u) => u.id),
                participantUsernames,
                participantDisplayNames,
                participantProfilePictures,
                creatorUsername: user.username,
                creatorDisplayName: user.displayName ?? undefined,
                creatorProfilePicture: user.profilePicture ?? undefined,
                name: isGroup ? groupName || undefined : undefined,
            });

            if (res.data) {
                addConversation(res.data);
                onClose();
                router.push(`./messages/${res.data._id}`);
            }
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="bg-background/80 fixed inset-0 z-50 flex items-center justify-center backdrop-blur-sm">
            <div className="bg-background w-full max-w-md rounded-lg border p-6 shadow-lg">
                <div className="mb-4 flex items-center justify-between">
                    <h2 className="text-lg font-semibold">New Conversation</h2>
                    <Button variant="ghost" size="sm" onClick={onClose}>
                        &times;
                    </Button>
                </div>

                <div className="mb-4 flex items-center gap-2">
                    <Button
                        variant={!isGroup ? "default" : "outline"}
                        size="sm"
                        onClick={() => {
                            setIsGroup(false);
                            setSelectedUsers(selectedUsers.slice(0, 1));
                        }}
                    >
                        Direct Message
                    </Button>
                    <Button
                        variant={isGroup ? "default" : "outline"}
                        size="sm"
                        onClick={() => setIsGroup(true)}
                    >
                        Group
                    </Button>
                </div>

                {isGroup && (
                    <Input
                        placeholder="Group name (optional)"
                        value={groupName}
                        onChange={(e) => setGroupName(e.target.value)}
                        className="mb-3"
                    />
                )}

                {/* Selected users */}
                {selectedUsers.length > 0 && (
                    <div className="mb-3 flex flex-wrap gap-2">
                        {selectedUsers.map((u) => (
                            <span
                                key={u.id}
                                className="bg-muted flex items-center gap-1 rounded-full px-3 py-1 text-sm"
                            >
                                {u.displayName || u.username}
                                <button
                                    onClick={() => handleRemoveUser(u.id)}
                                    className="text-muted-foreground ml-1 hover:text-red-500"
                                >
                                    &times;
                                </button>
                            </span>
                        ))}
                    </div>
                )}

                {/* Search */}
                <Input
                    placeholder="Search users..."
                    value={searchQuery}
                    onChange={(e) => handleSearch(e.target.value)}
                    className="mb-3"
                />

                {/* Search results */}
                <div className="max-h-48 overflow-y-auto">
                    {isSearching && (
                        <p className="text-muted-foreground py-2 text-center text-sm">
                            Searching...
                        </p>
                    )}
                    {searchResults.map((result) => (
                        <button
                            key={result.id}
                            onClick={() => handleSelectUser(result)}
                            className="hover:bg-accent flex w-full items-center gap-3 rounded px-3 py-2"
                        >
                            <Avatar className="h-8 w-8">
                                {result.profilePicture && (
                                    <AvatarImage src={result.profilePicture} />
                                )}
                                <AvatarFallback>
                                    {result.username.charAt(0).toUpperCase()}
                                </AvatarFallback>
                            </Avatar>
                            <div className="text-left">
                                <p className="text-sm font-medium">
                                    {result.displayName || result.username}
                                </p>
                                <p className="text-muted-foreground text-xs">
                                    @{result.username}
                                </p>
                            </div>
                        </button>
                    ))}
                </div>

                <div className="mt-4 flex justify-end gap-2">
                    <Button variant="outline" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button
                        disabled={selectedUsers.length === 0 || isLoading}
                        onClick={handleCreate}
                    >
                        {isLoading ? "Creating..." : "Create"}
                    </Button>
                </div>
            </div>
        </div>
    );
}
