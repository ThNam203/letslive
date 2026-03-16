import { Header } from "@/app/[lng]/(main)/_components/header/header";
import { ResizableLayoutClient } from "@/app/[lng]/(main)/_components/resizable-layout-client";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <div className="flex h-screen w-screen flex-col overflow-hidden">
            <Header />
            <ResizableLayoutClient>{children}</ResizableLayoutClient>
        </div>
    );
}
