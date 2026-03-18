"use client";

import { useRef, useState } from "react";
import { toast } from "@/components/utils/toast";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import useUser from "@/hooks/user";
import useUploadStore from "@/hooks/use-upload-store";
import Section from "../_components/section";
import TextField from "../_components/text-field";
import TextAreaField from "../_components/textarea-field";
import useT from "@/hooks/use-translation";
import {
    VOD_TITLE_MAX_LENGTH,
    VOD_DESCRIPTION_MAX_LENGTH,
} from "@/constant/field-limits";

const ALLOWED_EXTENSIONS = [".mp4", ".mov", ".avi", ".mkv", ".webm"];
const MAX_FILE_SIZE = 2 * 1024 * 1024 * 1024; // 2GB
const MAX_CONCURRENT = 3;

function formatFileSize(bytes: number): string {
    if (bytes < 1024 * 1024) {
        return `${(bytes / 1024).toFixed(1)} KB`;
    }
    if (bytes < 1024 * 1024 * 1024) {
        return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    }
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

export default function UploadVODPage() {
    const { t } = useT(["settings", "api-response", "fetch-error"]);
    const user = useUser((state) => state.user);
    const { enqueue, items } = useUploadStore();

    const fileInputRef = useRef<HTMLInputElement>(null);
    const [file, setFile] = useState<File | null>(null);
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const [isPublic, setIsPublic] = useState(true);

    const activeCount = items.filter(
        (i) => i.status === "uploading" || i.status === "queued",
    ).length;

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const selected = e.target.files?.[0];
        if (!selected) return;

        const ext = "." + selected.name.split(".").pop()?.toLowerCase();
        if (!ALLOWED_EXTENSIONS.includes(ext)) {
            toast(t("settings:upload.error_invalid_format"), { type: "error" });
            return;
        }

        if (selected.size > MAX_FILE_SIZE) {
            toast(t("settings:upload.error_file_too_large"), { type: "error" });
            return;
        }

        setFile(selected);
        if (!title) {
            setTitle(selected.name.replace(/\.[^/.]+$/, ""));
        }
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;

        if (!file) {
            toast(t("settings:upload.error_no_file"), { type: "error" });
            return;
        }

        if (!title.trim()) {
            toast(t("settings:upload.error_no_title"), { type: "error" });
            return;
        }

        if (activeCount >= MAX_CONCURRENT) {
            toast(
                t("settings:upload.error_max_concurrent", {
                    defaultValue: `Maximum ${MAX_CONCURRENT} concurrent uploads allowed. Your video will be queued.`,
                }),
                { type: "info" },
            );
        }

        enqueue(
            file,
            title.trim(),
            description.trim(),
            isPublic ? "public" : "private",
        );

        toast(t("settings:upload.upload_queued", {
            defaultValue: "Video added to upload queue",
        }), { type: "success" });

        // Reset form for next upload
        setFile(null);
        setTitle("");
        setDescription("");
        setIsPublic(true);
        if (fileInputRef.current) {
            fileInputRef.current.value = "";
        }
    };

    return (
        <Section
            title={t("settings:upload.title")}
            description={t("settings:upload.description")}
            contentClassName="p-4"
        >
            <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                    <label className="mb-2 block text-sm font-medium">
                        {t("settings:upload.select_video")}
                    </label>
                    <div
                        className="border-border hover:bg-muted/50 flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 transition-colors"
                        onClick={() => fileInputRef.current?.click()}
                    >
                        <input
                            ref={fileInputRef}
                            type="file"
                            accept="video/mp4,video/quicktime,video/x-msvideo,video/x-matroska,video/webm,.mp4,.mov,.avi,.mkv,.webm"
                            onChange={handleFileChange}
                            className="hidden"
                        />
                        {file ? (
                            <div className="text-center">
                                <p className="text-foreground text-sm font-medium">
                                    {t("settings:upload.selected_file", {
                                        filename: file.name,
                                        size: formatFileSize(file.size),
                                    })}
                                </p>
                                <p className="text-foreground-muted mt-1 text-xs">
                                    {t("settings:upload.change_video")}
                                </p>
                            </div>
                        ) : (
                            <div className="text-center">
                                <p className="text-foreground-muted text-sm">
                                    {t("settings:upload.select_video")}
                                </p>
                                <p className="text-foreground-muted mt-1 text-xs">
                                    {t(
                                        "settings:upload.select_video_description",
                                    )}
                                </p>
                            </div>
                        )}
                    </div>
                </div>

                <TextField
                    label={t("settings:upload.title_label")}
                    description={t("settings:upload.title_description")}
                    maxLength={VOD_TITLE_MAX_LENGTH}
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                />

                <TextAreaField
                    label={t("settings:upload.description_label")}
                    maxLength={VOD_DESCRIPTION_MAX_LENGTH}
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    rows={4}
                />

                <div className="flex items-center space-x-2">
                    <Switch
                        id="isPublic"
                        checked={isPublic}
                        onCheckedChange={setIsPublic}
                    />
                    <Label htmlFor="isPublic">
                        {t("settings:upload.visibility_label")}
                    </Label>
                </div>

                <div className="flex items-center justify-end">
                    <Button disabled={!file || !title.trim()} type="submit">
                        {t("settings:upload.upload_button")}
                    </Button>
                </div>
            </form>
        </Section>
    );
}
