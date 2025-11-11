import Link from "next/link";
import SearchBar from "./search";
import UserInfo from "./userinfo";
import StreamGuide from "./stream-guide";
import { myGetT } from "@/lib/i18n";
import HeaderUtilsForNonLogged from "@/app/[lng]/(main)/_components/header/header_utils_for_non_logged";

export async function Header() {
    const { t } = await myGetT();

    return (
        <nav className="sticky flex h-14 w-full flex-row items-center border-b border-border bg-background px-4 py-2 text-xl font-semibold text-foreground">
            <div className="flex flex-1 flex-row items-center max-md:gap-4 md:gap-10">
                <Link
                    href="/"
                    className="transition-all duration-200 hover:scale-[1.02] max-md:hidden"
                >
                    {t("app_title")}
                </Link>
            </div>

            <div className="flex flex-1 justify-center">
                <SearchBar className="max-lg:w mx-2 lg:w-[400px]" />
            </div>
            
            <div className="flex flex-1 flex-row items-center justify-end gap-4">
                <StreamGuide />
                <HeaderUtilsForNonLogged/>
                <UserInfo />
            </div>
        </nav>
    );
}
