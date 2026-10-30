import { Suspense } from "react";
import Link from "next/link";
import SearchBar from "./search";
import UserInfo from "./userinfo";
import NotificationBell from "./notification-bell";
import MessagesIcon from "./messages-icon";
import WalletIcon from "./wallet-icon";
import StreamGuide from "./stream-guide";
import UploadIcon from "./upload-icon";
import { myGetT } from "@/lib/i18n";

export async function Header() {
    const { t } = await myGetT();

    return (
        <nav className="border-border bg-background text-foreground sticky z-40 flex h-14 w-full flex-row items-center border-b px-4 py-2 text-xl font-semibold">
            <div className="flex flex-row items-center max-md:gap-4 md:gap-10">
                <Link
                    href="/"
                    className="whitespace-nowrap transition-all duration-200 hover:scale-[1.02]"
                >
                    {t("app_title")}
                </Link>
            </div>

            <Suspense fallback={<div className="mx-2 flex-1" />}>
                <SearchBar className="mx-2 flex w-full flex-1" />
            </Suspense>

            <div className="flex flex-row items-center justify-end gap-2 md:gap-4">
                <StreamGuide />
                <MessagesIcon />
                <WalletIcon />
                <NotificationBell />
                <UploadIcon />
                <UserInfo />
            </div>
        </nav>
    );
}
