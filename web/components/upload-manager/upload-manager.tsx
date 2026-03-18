"use client";

import useUploadStore, { UploadItem } from "@/hooks/use-upload-store";
import IconChevronUp from "@/components/icons/chevron-up";
import IconChevronDown from "@/components/icons/chevron-down";
import IconClose from "@/components/icons/close";
import IconLoader from "@/components/icons/loader";

function formatBytes(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 * 1024 * 1024)
        return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

function UploadItemRow({ item }: { item: UploadItem }) {
    const { cancel, retry, dismiss } = useUploadStore();

    return (
        <div className="border-border flex items-center gap-3 border-b px-4 py-3 last:border-b-0">
            <div className="min-w-0 flex-1">
                <p className="text-foreground truncate text-sm font-medium">
                    {item.title}
                </p>
                <div className="mt-1">
                    {item.status === "queued" && (
                        <p className="text-foreground-muted text-xs">
                            Waiting...
                        </p>
                    )}
                    {item.status === "uploading" && (
                        <>
                            <div className="bg-muted h-1.5 w-full overflow-hidden rounded-full">
                                <div
                                    className="bg-primary h-full rounded-full transition-all duration-300"
                                    style={{ width: `${item.progress}%` }}
                                />
                            </div>
                            <p className="text-foreground-muted mt-1 text-xs">
                                {formatBytes(item.loaded)} /{" "}
                                {formatBytes(item.total)} ({item.progress}%)
                            </p>
                        </>
                    )}
                    {item.status === "processing" && (
                        <div className="flex items-center gap-1.5">
                            <IconLoader className="h-3 w-3" />
                            <p className="text-foreground-muted text-xs">
                                Processing on server...
                            </p>
                        </div>
                    )}
                    {item.status === "completed" && (
                        <p className="text-xs text-green-500">
                            Upload complete
                        </p>
                    )}
                    {item.status === "failed" && (
                        <p className="text-xs text-red-500">
                            {item.error || "Upload failed"}
                        </p>
                    )}
                    {item.status === "cancelled" && (
                        <p className="text-foreground-muted text-xs">
                            Cancelled
                        </p>
                    )}
                </div>
            </div>
            <div className="flex shrink-0 items-center gap-1">
                {(item.status === "uploading" || item.status === "queued") && (
                    <button
                        onClick={() => cancel(item.id)}
                        className="text-foreground-muted hover:text-foreground rounded p-1 transition-colors"
                        title="Cancel"
                    >
                        <IconClose width={16} height={16} />
                    </button>
                )}
                {(item.status === "failed" ||
                    item.status === "cancelled") && (
                    <>
                        <button
                            onClick={() => retry(item.id)}
                            className="text-primary hover:text-primary/80 rounded px-2 py-0.5 text-xs font-medium transition-colors"
                        >
                            Retry
                        </button>
                        <button
                            onClick={() => dismiss(item.id)}
                            className="text-foreground-muted hover:text-foreground rounded p-1 transition-colors"
                            title="Dismiss"
                        >
                            <IconClose width={16} height={16} />
                        </button>
                    </>
                )}
                {item.status === "completed" && (
                    <button
                        onClick={() => dismiss(item.id)}
                        className="text-foreground-muted hover:text-foreground rounded p-1 transition-colors"
                        title="Dismiss"
                    >
                        <IconClose width={16} height={16} />
                    </button>
                )}
            </div>
        </div>
    );
}

export default function UploadManager() {
    const { items, isCollapsed, toggleCollapsed, dismissCompleted } =
        useUploadStore();

    if (items.length === 0) return null;

    const activeCount = items.filter(
        (i) => i.status === "uploading" || i.status === "queued",
    ).length;
    const completedCount = items.filter(
        (i) => i.status === "completed",
    ).length;
    const failedCount = items.filter((i) => i.status === "failed").length;

    let headerText: string;
    if (activeCount > 0) {
        headerText = `Uploading ${activeCount} video${activeCount > 1 ? "s" : ""}`;
    } else if (failedCount > 0) {
        headerText = `${failedCount} upload${failedCount > 1 ? "s" : ""} failed`;
    } else {
        headerText = `${completedCount} upload${completedCount > 1 ? "s" : ""} complete`;
    }

    return (
        <div className="fixed right-4 bottom-4 z-50 w-96 overflow-hidden rounded-lg shadow-lg">
            {/* Header - always visible */}
            <button
                onClick={toggleCollapsed}
                className="bg-card border-border flex w-full items-center justify-between border px-4 py-3"
            >
                <div className="flex items-center gap-2">
                    {activeCount > 0 && (
                        <IconLoader className="h-4 w-4" />
                    )}
                    <span className="text-foreground text-sm font-medium">
                        {headerText}
                    </span>
                </div>
                <div className="flex items-center gap-1">
                    {completedCount > 0 && activeCount === 0 && (
                        <span
                            onClick={(e) => {
                                e.stopPropagation();
                                dismissCompleted();
                            }}
                            className="text-foreground-muted hover:text-foreground cursor-pointer px-1 text-xs"
                        >
                            Clear
                        </span>
                    )}
                    {isCollapsed ? (
                        <IconChevronUp width={18} height={18} />
                    ) : (
                        <IconChevronDown width={18} height={18} />
                    )}
                </div>
            </button>

            {/* Expandable body */}
            {!isCollapsed && (
                <div className="bg-card border-border max-h-80 overflow-y-auto border-x border-b">
                    {items.map((item) => (
                        <UploadItemRow key={item.id} item={item} />
                    ))}
                </div>
            )}
        </div>
    );
}
