"use client";

import useUploadStore, {
    UploadItem,
    UPLOAD_STATUS,
} from "@/hooks/use-upload-store";
import useT from "@/hooks/use-translation";
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
    const { t } = useT(["settings"]);
    const { cancel, retry, dismiss } = useUploadStore();

    return (
        <div className="border-border flex items-center gap-3 border-b px-4 py-3 last:border-b-0">
            <div className="min-w-0 flex-1">
                <p className="text-foreground truncate text-sm font-medium">
                    {item.title}
                </p>
                <div className="mt-1">
                    {item.status === UPLOAD_STATUS.QUEUED && (
                        <p className="text-foreground-muted text-xs">
                            {t("settings:upload.manager_waiting")}
                        </p>
                    )}
                    {item.status === UPLOAD_STATUS.UPLOADING && (
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
                    {item.status === UPLOAD_STATUS.PROCESSING && (
                        <div className="flex items-center gap-1.5">
                            <IconLoader className="h-3 w-3" />
                            <p className="text-foreground-muted text-xs">
                                {t("settings:upload.manager_processing")}
                            </p>
                        </div>
                    )}
                    {item.status === UPLOAD_STATUS.COMPLETED && (
                        <p className="text-xs text-green-500">
                            {t("settings:upload.manager_complete")}
                        </p>
                    )}
                    {item.status === UPLOAD_STATUS.FAILED && (
                        <p className="text-xs text-red-500">
                            {item.errorCode
                                ? t(`settings:upload.${item.errorCode}`)
                                : item.error ||
                                  t("settings:upload.manager_failed")}
                        </p>
                    )}
                    {item.status === UPLOAD_STATUS.CANCELLED && (
                        <p className="text-foreground-muted text-xs">
                            {t("settings:upload.manager_cancelled")}
                        </p>
                    )}
                </div>
            </div>
            <div className="flex shrink-0 items-center gap-1">
                {(item.status === UPLOAD_STATUS.UPLOADING ||
                    item.status === UPLOAD_STATUS.QUEUED) && (
                    <button
                        onClick={() => cancel(item.id)}
                        className="text-foreground-muted hover:text-foreground rounded p-1 transition-colors"
                        title={t("settings:upload.manager_cancel")}
                    >
                        <IconClose width="16" height="16" />
                    </button>
                )}
                {(item.status === UPLOAD_STATUS.FAILED ||
                    item.status === UPLOAD_STATUS.CANCELLED) && (
                    <>
                        <button
                            onClick={() => retry(item.id)}
                            className="text-primary hover:text-primary/80 rounded px-2 py-0.5 text-xs font-medium transition-colors"
                        >
                            {t("settings:upload.manager_retry")}
                        </button>
                        <button
                            onClick={() => dismiss(item.id)}
                            className="text-foreground-muted hover:text-foreground rounded p-1 transition-colors"
                            title={t("settings:upload.manager_dismiss")}
                        >
                            <IconClose width="16" height="16" />
                        </button>
                    </>
                )}
                {item.status === UPLOAD_STATUS.COMPLETED && (
                    <button
                        onClick={() => dismiss(item.id)}
                        className="text-foreground-muted hover:text-foreground rounded p-1 transition-colors"
                        title={t("settings:upload.manager_dismiss")}
                    >
                        <IconClose width="16" height="16" />
                    </button>
                )}
            </div>
        </div>
    );
}

export default function UploadManager() {
    const { t } = useT(["settings"]);
    const { items, isCollapsed, toggleCollapsed, dismissCompleted } =
        useUploadStore();

    if (items.length === 0) return null;

    const activeCount = items.filter(
        (i) =>
            i.status === UPLOAD_STATUS.UPLOADING ||
            i.status === UPLOAD_STATUS.QUEUED,
    ).length;
    const completedCount = items.filter(
        (i) => i.status === UPLOAD_STATUS.COMPLETED,
    ).length;
    const failedCount = items.filter(
        (i) => i.status === UPLOAD_STATUS.FAILED,
    ).length;

    let headerText: string;
    if (activeCount > 0) {
        headerText = t("settings:upload.manager_header_uploading", {
            count: activeCount,
        });
    } else if (failedCount > 0) {
        headerText = t("settings:upload.manager_header_failed", {
            count: failedCount,
        });
    } else {
        headerText = t("settings:upload.manager_header_complete", {
            count: completedCount,
        });
    }

    return (
        <div className="border-border bg-background fixed right-4 bottom-4 w-96 rounded-md border shadow-lg">
            {/* Header - always visible */}
            <button
                onClick={toggleCollapsed}
                className="border-border flex w-full items-center justify-between border-b px-4 py-3"
            >
                <div className="flex items-center gap-2">
                    {activeCount > 0 && <IconLoader className="h-4 w-4" />}
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
                            {t("settings:upload.manager_clear")}
                        </span>
                    )}
                    {isCollapsed ? (
                        <IconChevronUp width="18" height="18" />
                    ) : (
                        <IconChevronDown width="18" height="18" />
                    )}
                </div>
            </button>

            {/* Expandable body */}
            {!isCollapsed && (
                <div className="max-h-80 overflow-y-auto">
                    {items.map((item) => (
                        <UploadItemRow key={item.id} item={item} />
                    ))}
                </div>
            )}
        </div>
    );
}
