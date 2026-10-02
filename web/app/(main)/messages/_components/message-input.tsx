"use client";

import type React from "react";
import { useState, useRef, useCallback } from "react";
import Image from "next/image";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import IconSend from "@/components/icons/send";
import IconPaperclip from "@/components/icons/paperclip";
import IconClose from "@/components/icons/close";
import EmotePicker from "@/components/emote-picker";
import { DM_MESSAGE_MAX_LENGTH } from "@/constant/field-limits";
import {
    CHAT_ATTACHMENT_IMAGE,
    GENERAL_UPLOAD_MAX_FILE_MB,
    IMAGE_INPUT_ACCEPT,
} from "@/constant/image";
import { IsValidFileSizeInMB } from "@/utils/file";
import {
    exportWholeImage,
    fitsSourceLimit,
    loadImageFile,
    naturalSize,
} from "@/utils/image-crop";
import { useUploadFiles } from "@/hooks/queries/use-file-upload";
import useT from "@/hooks/use-translation";

const MAX_FILES = 10;

type SelectedFile = {
    file: File;
    previewUrl: string;
};

type PreparedAttachment = { file: File } | { error: string };

export default function MessageInput({
    onSend,
    onTypingStart,
    onTypingStop,
}: {
    onSend: (text: string, imageUrls?: string[]) => void;
    onTypingStart: () => void;
    onTypingStop: () => void;
}) {
    const [text, setText] = useState("");
    const [selectedFiles, setSelectedFiles] = useState<SelectedFile[]>([]);
    const [uploadError, setUploadError] = useState<string | null>(null);
    const [isPreparing, setIsPreparing] = useState(false);
    const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const isTypingRef = useRef(false);
    const fileInputRef = useRef<HTMLInputElement>(null);
    const { t } = useT("messages");
    const uploadFiles = useUploadFiles();
    const isUploading = uploadFiles.isPending;
    const isBusy = isUploading || isPreparing;

    const handleTyping = useCallback(() => {
        if (!isTypingRef.current) {
            isTypingRef.current = true;
            onTypingStart();
        }

        if (typingTimeoutRef.current) {
            clearTimeout(typingTimeoutRef.current);
        }

        typingTimeoutRef.current = setTimeout(() => {
            isTypingRef.current = false;
            onTypingStop();
        }, 2000);
    }, [onTypingStart, onTypingStop]);

    const prepareAttachment = async (
        file: File,
    ): Promise<PreparedAttachment> => {
        const img = await loadImageFile(file);
        if (!fitsSourceLimit(naturalSize(img), CHAT_ATTACHMENT_IMAGE)) {
            return {
                error: t("image_exceeds_dimensions", {
                    name: file.name,
                    max: CHAT_ATTACHMENT_IMAGE.sourceMaxDimension,
                }),
            };
        }
        // a canvas export keeps only the first frame of a GIF
        if (file.type === "image/gif") return { file };

        return {
            file: await exportWholeImage(
                img,
                CHAT_ATTACHMENT_IMAGE,
                "attachment",
            ),
        };
    };

    const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = Array.from(e.target.files ?? []);
        // Reset input so the same files can be re-selected
        e.target.value = "";
        if (files.length === 0) return;

        setUploadError(null);

        const remaining = MAX_FILES - selectedFiles.length;
        if (files.length > remaining) {
            setUploadError(t("attach_images_limit", { max: MAX_FILES }));
        }

        setIsPreparing(true);
        const newFiles: SelectedFile[] = [];
        // one at a time: each decoded photo can take hundreds of MB
        for (const file of files.slice(0, remaining)) {
            if (!IsValidFileSizeInMB(file, GENERAL_UPLOAD_MAX_FILE_MB)) {
                setUploadError(
                    t("file_exceeds_limit", {
                        name: file.name,
                        size: GENERAL_UPLOAD_MAX_FILE_MB,
                    }),
                );
                continue;
            }

            try {
                const prepared = await prepareAttachment(file);
                if ("error" in prepared) {
                    setUploadError(prepared.error);
                    continue;
                }
                newFiles.push({
                    file: prepared.file,
                    previewUrl: URL.createObjectURL(prepared.file),
                });
            } catch {
                setUploadError(t("image_unreadable", { name: file.name }));
            }
        }
        setIsPreparing(false);

        if (newFiles.length > 0) {
            setSelectedFiles((prev) => [...prev, ...newFiles]);
        }
    };

    const removeFile = useCallback((index: number) => {
        setSelectedFiles((prev) => {
            const removed = prev[index];
            if (removed) {
                URL.revokeObjectURL(removed.previewUrl);
            }
            return prev.filter((_, i) => i !== index);
        });
        setUploadError(null);
    }, []);

    const clearAllFiles = useCallback(() => {
        setSelectedFiles((prev) => {
            for (const f of prev) {
                URL.revokeObjectURL(f.previewUrl);
            }
            return [];
        });
        setUploadError(null);
    }, []);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const trimmed = text.trim();
        if (!trimmed && selectedFiles.length === 0) return;

        // Clear typing state
        if (typingTimeoutRef.current) {
            clearTimeout(typingTimeoutRef.current);
        }
        if (isTypingRef.current) {
            isTypingRef.current = false;
            onTypingStop();
        }

        if (selectedFiles.length > 0) {
            setUploadError(null);
            uploadFiles.mutate(
                selectedFiles.map((sf) => sf.file),
                {
                    onSuccess: ({ uploadedUrls, hasError }) => {
                        if (uploadedUrls.length > 0) {
                            const msgText =
                                trimmed ||
                                (uploadedUrls.length === 1
                                    ? "Sent an image"
                                    : `Sent ${uploadedUrls.length} images`);
                            onSend(msgText, uploadedUrls);
                            setText("");
                            clearAllFiles();
                        }

                        if (hasError && uploadedUrls.length === 0) {
                            setUploadError(t("upload_failed"));
                        } else if (hasError) {
                            setUploadError(t("upload_some_failed"));
                        }
                    },
                    onError: () => setUploadError(t("upload_failed")),
                },
            );
        } else {
            onSend(trimmed);
            setText("");
        }
    };

    return (
        <div className="border-t px-4 py-3">
            {/* File previews */}
            {selectedFiles.length > 0 && (
                <div className="mb-2 flex flex-wrap gap-2">
                    {selectedFiles.map((sf, index) => (
                        <div key={sf.previewUrl} className="relative">
                            <Image
                                src={sf.previewUrl}
                                alt={t("preview_alt", { number: index + 1 })}
                                unoptimized
                                layout="fill"
                                className="h-20 w-20 rounded-lg border object-cover"
                            />
                            <button
                                type="button"
                                onClick={() => removeFile(index)}
                                className="bg-background/80 hover:bg-background absolute -top-1.5 -right-1.5 rounded-full border p-0.5"
                            >
                                <IconClose className="size-3" />
                            </button>
                        </div>
                    ))}
                </div>
            )}

            {/* Upload error */}
            {uploadError && (
                <p className="text-destructive mb-1 text-xs">{uploadError}</p>
            )}

            <form onSubmit={handleSubmit} className="flex items-start gap-2">
                {/* File upload button */}
                <input
                    ref={fileInputRef}
                    type="file"
                    accept={IMAGE_INPUT_ACCEPT}
                    multiple
                    className="hidden"
                    onChange={handleFileSelect}
                />
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9 shrink-0"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={isBusy || selectedFiles.length >= MAX_FILES}
                >
                    <IconPaperclip className="!h-5 !w-5" />
                </Button>
                <div className="relative flex-1">
                    <Input
                        type="text"
                        placeholder={t("placeholder_type_message")}
                        maxLength={DM_MESSAGE_MAX_LENGTH}
                        showCount
                        value={text}
                        onChange={(e) => {
                            setText(e.target.value);
                            handleTyping();
                        }}
                        disabled={isUploading}
                    />
                </div>
                <EmotePicker
                    disabled={isUploading}
                    searchPlaceholder={t("emote_search_placeholder")}
                    emptyStateText={t("emote_empty_state")}
                    getCategoryLabel={(category) =>
                        t(`emote_category_${category}`)
                    }
                    onSelect={(code) => setText((prev) => prev + code)}
                />
                <Button
                    type="submit"
                    disabled={
                        (!text.trim() && selectedFiles.length === 0) || isBusy
                    }
                    className="h-9 w-12 shrink-0 p-0"
                >
                    {isUploading ? (
                        <span className="size-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                    ) : (
                        <IconSend className="!h-5 !w-5" />
                    )}
                </Button>
            </form>
        </div>
    );
}
