import Link from "next/link";
import SearchBar from "./search";
import UserInfo from "./userinfo";
import StreamGuiding from "./stream-guiding";
import ThemeSwitch from "../utils/theme-switch";

export async function Header() {
    return (
        <nav className="sticky flex h-12 w-full flex-row items-center justify-between bg-background px-4 py-2 text-xl font-semibold text-foreground border-b border-border">
            <div className="flex flex-row items-center max-md:gap-4 md:gap-10">
                <Link
                    href="/"
                    className="transition-all duration-200 hover:opacity-75"
                >
                    <span className="max-md:hidden">Home</span>
                </Link>
            </div>

            <div className="max-lg:w mx-2 lg:w-[400px]">
                <SearchBar />
            </div>
            <div className="flex flex-row items-center gap-4">
                <ThemeSwitch />
                <StreamGuiding />
                <UserInfo />
            </div>
        </nav>
    );
}
