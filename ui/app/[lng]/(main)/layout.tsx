import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";
import { Header } from "@/components/header/header";
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
                    className="bg-background"
                >
                    {children}
                </ResizablePanel>
            </ResizablePanelGroup>
        </div>
    );
}