import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/src/components/ui/resizable";
import { Header } from "@/src/routes/[lng]/(main)/_components/header/header";
import LeftBar from "@/src/components/leftbar/left-bar";

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <div className="flex h-screen w-screen flex-col overflow-hidden">
            <Header />
            <ResizablePanelGroup
                autoSaveId={"sidebar-size"}
                direction="horizontal"
                className="flex-1"
            >
                <LeftBar />
                <ResizableHandle />
                <ResizablePanel id="2" order={2} className="bg-background">
                    {children}
                </ResizablePanel>
            </ResizablePanelGroup>
        </div>
    );
}
