import {
    LuCopy,
    LuHeart,
    LuHome
} from "react-icons/lu";
import Link from "next/link";
import { SearchInput } from "@/components/Input";
import UserInfo from "@/components/header/userinfo";

export async function Header() {
    return (
        <nav className="w-full h-12 flex flex-row items-center justify-between text-xl font-semibold text-primaryWord bg-white px-4 py-2 shadow z-[49]">
            <div className="flex flex-row md:gap-10 max-md:gap-4 items-center">
                <Link href="/" className="hover:text-primary">
                    <span className="max-md:hidden">Home</span>
                    <LuHome size={20} className="md:hidden" />
                </Link>
                <Link href="/following" className="hover:text-primary">
                    <span className="max-md:hidden">Following</span>
                    <LuHeart size={20} className="md:hidden" />
                </Link>
                <Link href="/browse" className="hover:text-primary">
                    <span className="max-md:hidden">Browse</span>
                    <LuCopy size={20} className="md:hidden" />
                </Link>
            </div>

            <div className="lg:w-[400px] max-lg:w mx-2">
                <SearchInput
                    id="search-input"
                    placeholder="Search (Not implemented)"
                    className="text-base w-full"
                />
            </div>

            <UserInfo/>
        </nav>
    );
}
