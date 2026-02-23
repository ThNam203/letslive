import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";
import { Header } from "@/app/[lng]/(main)/_components/header/header";
import LeftBar from "@/components/leftbar/left-bar";

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
                <ResizablePanel
                    id="2"
                    order={2}
                    className="min-h-0 overflow-hidden bg-background"
                >
                    {children}
                </ResizablePanel>
            </ResizablePanelGroup>
        </div>
    );
}
