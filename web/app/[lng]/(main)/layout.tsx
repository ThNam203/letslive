import { Header } from "@/app/[lng]/(main)/_components/header/header";
import { MainBodyLayout } from "@/app/[lng]/(main)/_components/main-body-layout";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <div className="flex h-screen w-screen flex-col overflow-hidden">
            <Header />
            <MainBodyLayout>{children}</MainBodyLayout>
        </div>
    );
}
