"use client";

import { useEffect, useState } from "react";
import useDmStore from "@/hooks/use-dm-store";
import useUser from "@/hooks/user";
import { GetConversations, GetUnreadCounts } from "@/lib/api/dm";
import ConversationList from "./_components/conversation-list";
import NewConversationDialog from "./_components/new-conversation-dialog";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";

export default function MessagesPage() {
    const user = useUser((state) => state.user);
    const {
        conversations,
        setConversations,
        setUnreadCounts,
        setIsLoading,
        isLoading,
    } = useDmStore();
    const [showNewConversation, setShowNewConversation] = useState(false);
    const { t } = useT("messages");

    useEffect(() => {
        if (!user) return;

        const fetchData = async () => {
            setIsLoading(true);
            try {
                const [convRes, unreadRes] = await Promise.all([
                    GetConversations(0),
                    GetUnreadCounts(),
                ]);

                if (convRes.data) {
                    setConversations(convRes.data);
                }
                if (unreadRes.data) {
                    setUnreadCounts(unreadRes.data);
                }
            } finally {
                setIsLoading(false);
            }
        };

        fetchData();
    }, [user, setConversations, setUnreadCounts, setIsLoading]);

    if (!user) {
        return (
            <div className="flex h-full w-full items-center justify-center">
                <p className="text-muted-foreground">
                    {t("messages:login_required")}
                </p>
            </div>
        );
    }

    return (
        <div className="flex h-full w-full">
            <div className="flex h-full w-full flex-col md:w-80 md:border-r">
                <div className="flex items-center justify-between border-b p-4">
                    <h1 className="text-lg font-semibold">
                        {t("messages:title")}
                    </h1>
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setShowNewConversation(true)}
                    >
                        +
                    </Button>
                </div>

                <ConversationList
                    conversations={conversations}
                    isLoading={isLoading}
                />

                {showNewConversation && (
                    <NewConversationDialog
                        onClose={() => setShowNewConversation(false)}
                    />
                )}
            </div>
            <div className="hidden flex-1 items-center justify-center md:flex">
                <p className="text-muted-foreground">
                    {t("messages:select_conversation")}
                </p>
            </div>
        </div>
    );
}
