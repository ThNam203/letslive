"use client";

import Link from "next/link";
import { cn } from "@/lib/utils";
import { usePathname } from "next/navigation";

//     const [user, setUser] = useState<User | undefined>(undefined);

//     useEffect(() => {
//         const fetchProfile = async () => {
//             const { user, fetchError } = await GetMeProfile();

//             if (fetchError != undefined) {
//                 toast.error(fetchError.message, {
//                     toastId: "profile-fetch-error",
//                 });
//             } else {
//                 setUser(user);
//             }
//         };

//         fetchProfile();
//     }, []);

const navItems = [
    { name: "Profile", href: "/settings/profile" },
    { name: "Security", href: "/settings/security" }
];

export default function SettingsNav({
    children,
}: Readonly<{ children: React.ReactNode }>) {
    const pathname = usePathname();

    return (
        <>
            <div className="bg-white text-gray-900">
                <div className="max-w-7xl px-6">
                    <h1 className="text-4xl font-bold py-6">Settings</h1>
                    <nav className="border-b border-gray-200">
                        <ul className="flex flex-wrap gap-8">
                            {navItems.map((item) => (
                                <li key={item.href}>
                                    <Link
                                        href={item.href}
                                        className={cn(
                                            "inline-block py-4 text-sm relative hover:text-purple-600 transition-colors",
                                            pathname === item.href
                                                ? "text-gray-900"
                                                : "text-gray-500"
                                        )}
                                    >
                                        {item.name}
                                        {pathname === item.href && (
                                            <span className="absolute bottom-0 left-0 w-full h-0.5 bg-purple-600" />
                                        )}
                                    </Link>
                                </li>
                            ))}
                        </ul>
                    </nav>
                </div>
            </div>
            <div>{children}</div>
        </>
    );
}
