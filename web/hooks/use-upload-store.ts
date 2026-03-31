import { create } from "zustand";
import { ApiResponse } from "@/types/fetch-response";
import { VOD } from "@/types/vod";
import { uploadWithProgress, UploadClientError } from "@/utils/uploadClient";

const MAX_CONCURRENT = 3;

export type UploadStatus =
    | "queued"
    | "uploading"
    | "processing"
    | "completed"
    | "failed"
    | "cancelled";

export const UPLOAD_STATUS = {
    QUEUED: "queued",
    UPLOADING: "uploading",
    PROCESSING: "processing",
    COMPLETED: "completed",
    FAILED: "failed",
    CANCELLED: "cancelled",
} as const satisfies Record<string, UploadStatus>;

export type UploadItem = {
    id: string;
    file: File;
    title: string;
    description: string;
    visibility: string;
    status: UploadStatus;
    progress: number; // 0-100
    loaded: number;
    total: number;
    error?: string;
    /** Client-side error code for i18n (settings:upload.<errorCode>) */
    errorCode?: string;
    vod?: VOD;
    abort?: () => void;
};

export type UploadState = {
    items: UploadItem[];
    isCollapsed: boolean;

    enqueue: (
        file: File,
        title: string,
        description: string,
        visibility: string,
    ) => void;
    cancel: (id: string) => void;
    retry: (id: string) => void;
    dismiss: (id: string) => void;
    dismissCompleted: () => void;
    toggleCollapsed: () => void;
    setCollapsed: (collapsed: boolean) => void;
};

const generateId = () => crypto.randomUUID();

const useUploadStore = create<UploadState>((set, get) => ({
    items: [],
    isCollapsed: false,

    enqueue: (file, title, description, visibility) => {
        const item: UploadItem = {
            id: generateId(),
            file,
            title,
            description,
            visibility,
            status: UPLOAD_STATUS.QUEUED,
            progress: 0,
            loaded: 0,
            total: file.size,
        };

        set((state) => ({ items: [...state.items, item] }));
        processQueue(get, set);
    },

    cancel: (id) => {
        const item = get().items.find((i) => i.id === id);
        if (item?.abort) item.abort();

        set((state) => ({
            items: state.items.map((i) =>
                i.id === id ? { ...i, status: UPLOAD_STATUS.CANCELLED } : i,
            ),
        }));
        processQueue(get, set);
    },

    retry: (id) => {
        set((state) => ({
            items: state.items.map((i) =>
                i.id === id
                    ? {
                          ...i,
                          status: UPLOAD_STATUS.QUEUED,
                          progress: 0,
                          loaded: 0,
                          error: undefined,
                      }
                    : i,
            ),
        }));
        processQueue(get, set);
    },

    dismiss: (id) => {
        set((state) => ({
            items: state.items.filter((i) => i.id !== id),
        }));
    },

    dismissCompleted: () => {
        set((state) => ({
            items: state.items.filter(
                (i) =>
                    i.status !== UPLOAD_STATUS.COMPLETED &&
                    i.status !== UPLOAD_STATUS.FAILED &&
                    i.status !== UPLOAD_STATUS.CANCELLED,
            ),
        }));
    },

    toggleCollapsed: () => {
        set((state) => ({ isCollapsed: !state.isCollapsed }));
    },

    setCollapsed: (collapsed) => {
        set({ isCollapsed: collapsed });
    },
}));

function processQueue(
    get: () => UploadState,
    set: (updater: (state: UploadState) => Partial<UploadState>) => void,
) {
    const state = get();
    const activeCount = state.items.filter(
        (i) => i.status === UPLOAD_STATUS.UPLOADING,
    ).length;
    const availableSlots = MAX_CONCURRENT - activeCount;

    if (availableSlots <= 0) return;

    const queued = state.items.filter((i) => i.status === UPLOAD_STATUS.QUEUED);
    const toStart = queued.slice(0, availableSlots);

    for (const item of toStart) {
        startUpload(item, get, set);
    }
}

function startUpload(
    item: UploadItem,
    get: () => UploadState,
    set: (updater: (state: UploadState) => Partial<UploadState>) => void,
) {
    const formData = new FormData();
    formData.append("file", item.file);
    formData.append("title", item.title);
    formData.append("description", item.description);
    formData.append("visibility", item.visibility);

    const { promise, abort } = uploadWithProgress<ApiResponse<VOD>>(
        "/vods/upload",
        formData,
        (loaded, total) => {
            set((state) => ({
                items: state.items.map((i) =>
                    i.id === item.id
                        ? {
                              ...i,
                              loaded,
                              total,
                              progress:
                                  total > 0
                                      ? Math.round((loaded / total) * 100)
                                      : 0,
                          }
                        : i,
                ),
            }));
        },
    );

    // Mark as uploading and store abort handle
    set((state) => ({
        items: state.items.map((i) =>
            i.id === item.id
                ? { ...i, status: UPLOAD_STATUS.UPLOADING, abort }
                : i,
        ),
    }));

    promise
        .then((res) => {
            const currentItem = get().items.find((i) => i.id === item.id);
            if (currentItem?.status === UPLOAD_STATUS.CANCELLED) return;

            if (res.success) {
                set((state) => ({
                    items: state.items.map((i) =>
                        i.id === item.id
                            ? {
                                  ...i,
                                  status: UPLOAD_STATUS.COMPLETED,
                                  progress: 100,
                                  vod: res.data,
                              }
                            : i,
                    ),
                }));
            } else {
                set((state) => ({
                    items: state.items.map((i) =>
                        i.id === item.id
                            ? {
                                  ...i,
                                  status: UPLOAD_STATUS.FAILED,
                                  error: res.message,
                                  errorCode: undefined,
                              }
                            : i,
                    ),
                }));
            }
        })
        .catch((err) => {
            const currentItem = get().items.find((i) => i.id === item.id);
            if (currentItem?.status === UPLOAD_STATUS.CANCELLED) return;

            const isClientError = err instanceof UploadClientError;
            set((state) => ({
                items: state.items.map((i) =>
                    i.id === item.id
                        ? {
                              ...i,
                              status: UPLOAD_STATUS.FAILED,
                              error: isClientError ? err.message : err.message,
                              errorCode: isClientError ? err.code : undefined,
                          }
                        : i,
                ),
            }));
        })
        .finally(() => {
            processQueue(get, set);
        });
}

export default useUploadStore;
