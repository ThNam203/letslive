"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { PublicUser } from "@/types/user";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import useT from "@/hooks/use-translation";
import ProfileHomeTab from "./profile-home-tab";
import ProfileVideosTab from "./profile-videos-tab";
import ProfileAboutTab from "./profile-about-tab";

const CHANNEL_TABS = ["home", "videos", "about"] as const;
type ChannelTab = (typeof CHANNEL_TABS)[number];
const DEFAULT_TAB: ChannelTab = "home";

function parseTab(value: string | null): ChannelTab {
    return CHANNEL_TABS.find((tab) => tab === value) ?? DEFAULT_TAB;
}

const TAB_LIST_CLASS =
    "border-border h-auto w-full justify-start gap-6 rounded-none border-b bg-transparent p-0";
const TAB_TRIGGER_CLASS =
    "text-muted-foreground hover:text-foreground data-[state=active]:text-foreground data-[state=active]:border-primary -mb-px rounded-none border-b-2 border-transparent bg-transparent px-1 pt-1 pb-3 text-sm font-semibold data-[state=active]:bg-transparent data-[state=active]:shadow-none";

export default function ProfileTabs({ user }: { user: PublicUser }) {
    const { t } = useT("users");
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();
    const tab = parseTab(searchParams.get("tab"));

    const selectTab = (next: ChannelTab) => {
        const params = new URLSearchParams(searchParams.toString());
        if (next === DEFAULT_TAB) {
            params.delete("tab");
        } else {
            params.set("tab", next);
        }
        const query = params.toString();
        router.replace(query ? `${pathname}?${query}` : pathname, {
            scroll: false,
        });
    };

    return (
        <Tabs value={tab} onValueChange={(value) => selectTab(parseTab(value))}>
            <TabsList className={TAB_LIST_CLASS}>
                {CHANNEL_TABS.map((key) => (
                    <TabsTrigger
                        key={key}
                        value={key}
                        className={TAB_TRIGGER_CLASS}
                    >
                        {t(`users:channel.tabs.${key}`)}
                    </TabsTrigger>
                ))}
            </TabsList>
            <TabsContent value="home" className="@container mt-6">
                <ProfileHomeTab
                    user={user}
                    onViewAll={() => selectTab("videos")}
                />
            </TabsContent>
            <TabsContent value="videos" className="@container mt-6">
                <ProfileVideosTab user={user} />
            </TabsContent>
            <TabsContent value="about" className="@container mt-6">
                <ProfileAboutTab user={user} />
            </TabsContent>
        </Tabs>
    );
}
