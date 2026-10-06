import { Header } from "@/app/(main)/_components/header/header";
import { MainBodyLayout } from "@/app/(main)/_components/main-body-layout";
import { RealtimeProvider } from "@/contexts/realtime-context";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <RealtimeProvider>
            <div className="flex h-screen w-screen flex-col overflow-hidden">
                <Header />
                <MainBodyLayout>{children}</MainBodyLayout>
            </div>
        </RealtimeProvider>
    );
}
