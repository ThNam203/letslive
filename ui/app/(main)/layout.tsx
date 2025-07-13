import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { Header } from "../../components/header/header";
import LeftBar from "../../components/leftbar/left-bar";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="h-screen w-screen overflow-hidden">
      <Header />
      <ResizablePanelGroup
        autoSaveId={"sidebar-size"}
        direction="horizontal"
      >
        <LeftBar />
        <ResizableHandle />
        <ResizablePanel order={2} className="bg-background">
          {children}
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}