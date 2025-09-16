import "@/app/globals.css";
import LanguageSwitch from "@/components/utils/language-switch";
import ThemeSwitch from "@/components/utils/theme-switch";
import Link from "next/link";

export default async function RootLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <section className="flex h-screen w-screen items-center justify-center bg-background">
            <div className="absolute right-8 top-4 flex gap-4">
                <LanguageSwitch />
                <ThemeSwitch />
            </div>

            <div className="flex w-full max-w-[600px] flex-col justify-center rounded-xl p-12">
                <Link href={"/"}>
                    <h1 className="text-lg font-bold hover:underline">LET&apos;S LIVE</h1>
                </Link>
                {children}
            </div>
        </section>
    );
}
