"use client";

import type React from "react";
import { useState, useRef, useCallback } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import IconSend from "@/components/icons/send";
import IconPaperclip from "@/components/icons/paperclip";
import IconClose from "@/components/icons/close";
import { DM_MESSAGE_MAX_LENGTH } from "@/constant/field-limits";
import {
    FILE_SIZE_LIMIT_BYTES_UNIT,
    FILE_SIZE_LIMIT_MB_UNIT,
} from "@/constant/image";
import { UploadFile } from "@/lib/api/utils";

const ACCEPTED_FILE_TYPES = "image/png,image/jpeg,image/gif,image/webp";
const MAX_FILES = 10;

type SelectedFile = {
    file: File;
    previewUrl: string;
};

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
    const [isUploading, setIsUploading] = useState(false);
    const [uploadError, setUploadError] = useState<string | null>(null);
    const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const isTypingRef = useRef(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

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

    const handleFileSelect = useCallback(
        (e: React.ChangeEvent<HTMLInputElement>) => {
            const files = e.target.files;
            if (!files || files.length === 0) return;

            setUploadError(null);

            const newFiles: SelectedFile[] = [];
            const remaining = MAX_FILES - selectedFiles.length;

            if (files.length > remaining) {
                setUploadError(`You can attach up to ${MAX_FILES} images`);
            }

            const count = Math.min(files.length, remaining);
            for (let i = 0; i < count; i++) {
                const file = files[i];
                if (file.size > FILE_SIZE_LIMIT_BYTES_UNIT) {
                    setUploadError(
                        `"${file.name}" exceeds ${FILE_SIZE_LIMIT_MB_UNIT}MB limit`,
                    );
                    continue;
                }
                newFiles.push({
                    file,
                    previewUrl: URL.createObjectURL(file),
                });
            }

            if (newFiles.length > 0) {
                setSelectedFiles((prev) => [...prev, ...newFiles]);
            }

            // Reset input so the same files can be re-selected
            if (fileInputRef.current) {
                fileInputRef.current.value = "";
            }
        },
        [selectedFiles.length],
    );

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
            setIsUploading(true);
            setUploadError(null);
            try {
                const uploadResults = await Promise.all(
                    selectedFiles.map((sf) => UploadFile(sf.file)),
                );

                const uploadedUrls: string[] = [];
                let hasError = false;

                for (const res of uploadResults) {
                    if (res.success && res.data?.newPath) {
                        uploadedUrls.push(res.data.newPath);
                    } else {
                        hasError = true;
                    }
                }

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
                    setUploadError("Failed to upload files");
                } else if (hasError) {
                    setUploadError("Some files failed to upload");
                }
            } catch {
                setUploadError("Failed to upload files");
            } finally {
                setIsUploading(false);
            }
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
                            <img
                                src={sf.previewUrl}
                                alt={`Preview ${index + 1}`}
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

            <form onSubmit={handleSubmit} className="flex items-center gap-2">
                {/* File upload button */}
                <input
                    ref={fileInputRef}
                    type="file"
                    accept={ACCEPTED_FILE_TYPES}
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
                    disabled={isUploading || selectedFiles.length >= MAX_FILES}
                >
                    <IconPaperclip className="!h-5 !w-5" />
                </Button>

                <div className="relative flex-1">
                    <Input
                        type="text"
                        placeholder="Type a message..."
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
                <Button
                    type="submit"
                    disabled={
                        (!text.trim() && selectedFiles.length === 0) ||
                        isUploading
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
