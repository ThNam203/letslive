import RequireAuth from "@/components/wrappers/RequireAuth";
import NavTabs, { NavTab } from "./nav-tabs";

/** Title + tab bar over a signed-in-only content area (settings, wallet). */
export default function TabbedPageLayout({
    title,
    tabs,
    tabClassName,
    fallback,
    children,
}: {
    title: string;
    tabs: NavTab[];
    tabClassName?: string;
    fallback: React.ReactNode;
    children: React.ReactNode;
}) {
    return (
        <div className="bg-background text-foreground flex h-full flex-col">
            <div className="max-w-7xl px-6">
                <div className="mt-6 flex items-center">
                    <h1 className="text-4xl font-bold">{title}</h1>
                </div>
                <NavTabs tabs={tabs} tabClassName={tabClassName} />
            </div>
            <div className="text-foreground flex-1 overflow-y-auto p-6">
                <div className="max-w-4xl space-y-8">
                    <RequireAuth fallback={fallback}>{children}</RequireAuth>
                </div>
            </div>
        </div>
    );
}
