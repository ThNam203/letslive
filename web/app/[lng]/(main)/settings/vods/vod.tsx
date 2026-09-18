"use client";

import useT from "@/hooks/use-translation";
import type React from "react";
import { ChangeEvent, useEffect, useRef, useState } from "react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/utils/toast";
import GLOBAL from "@/global";
import IconSave from "@/components/icons/save";
import { VOD } from "@/types/vod";
import {
    VOD_TITLE_MAX_LENGTH,
    VOD_DESCRIPTION_MAX_LENGTH,
} from "@/constant/field-limits";
import IconLoader from "@/components/icons/loader";
import MediaCard from "@/components/livestream/media-card";
import { useDeleteVod, useUpdateVod } from "@/hooks/queries/use-vod-mutations";

export default function VODEditCard({ vod }: { vod: VOD }) {
    const { t } = useT(["common", "settings", "api-response"]);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [formData, setFormData] = useState<{
        title: string;
        description: string;
        thumbnailURL: string | null;
        image: File | undefined;
        selectedImage: string | undefined;
        isPublic?: boolean;
    }>({
        title: vod.title,
        description: vod.description || "",
        thumbnailURL: vod.thumbnailUrl
            ? vod.thumbnailUrl
            : `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`,
        image: undefined,
        selectedImage: undefined,
        isPublic: vod.visibility === "public",
    });

    const selectedImageRef = useRef<string | null>(null);
    const updateVod = useUpdateVod();
    const deleteVod = useDeleteVod();
    const isSubmitting = updateVod.isPending || deleteVod.isPending;

    const releaseSelectedImage = () => {
        if (selectedImageRef.current) {
            URL.revokeObjectURL(selectedImageRef.current);
            selectedImageRef.current = null;
        }
    };

    useEffect(() => {
        return () => {
            if (selectedImageRef.current)
                URL.revokeObjectURL(selectedImageRef.current);
        };
    }, []);

    const handleImageChange = (event: ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (file) {
            if (selectedImageRef.current)
                URL.revokeObjectURL(selectedImageRef.current);
            const imageUrl = URL.createObjectURL(file);
            selectedImageRef.current = imageUrl;
            setFormData((prev) => ({
                ...prev,
                image: file,
                selectedImage: imageUrl,
            }));
        }
    };

    const handleEdit = () => {
        setFormData({
            title: vod.title,
            description: vod.description || "",
            thumbnailURL: vod.thumbnailUrl
                ? vod.thumbnailUrl
                : `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`,
            image: undefined,
            selectedImage: undefined,
            isPublic: vod.visibility === "public",
        });
        setIsDialogOpen(true);
    };

    const handleDelete = () => {
        setIsDeleteDialogOpen(true);
    };

    const handleConfirmDelete = () => {
        deleteVod.mutate(vod.id, {
            onSettled: () => setIsDeleteDialogOpen(false),
        });
    };

    const handleSave = () => {
        updateVod.mutate(
            {
                vodId: vod.id,
                title: formData.title,
                description: formData.description,
                isPublic: Boolean(formData.isPublic),
                image: formData.image,
            },
            {
                onSuccess: () => {
                    toast(t("settings:vods.edit_dialog.update_success"), {
                        type: "success",
                    });
                },
                onSettled: () => {
                    releaseSelectedImage();
                    setIsDialogOpen(false);
                },
            },
        );
    };

    const handleCancel = () => {
        releaseSelectedImage();
        setIsDialogOpen(false);
    };

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target;
        setFormData((prev) => ({
            ...prev,
            [name]: value,
        }));
    };

    const handleSwitchChange = (checked: boolean) => {
        setFormData((prev) => ({
            ...prev,
            isPublic: checked,
        }));
    };

    return (
        <>
            <MediaCard
                kind="vod"
                vod={vod}
                variant="editable"
                onEdit={handleEdit}
                onDelete={handleDelete}
                className="w-[350px]"
            />

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>
                            {t("settings:vods.edit_dialog.title")}
                        </DialogTitle>
                        <DialogDescription>
                            {t("settings:vods.edit_dialog.description")}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="image-upload">
                                {t("settings:vods.edit_dialog.thumbnail")}
                            </Label>
                            <div className="col-span-3 w-full max-w-3xl">
                                <label
                                    htmlFor="image-upload"
                                    className={`group border-border group-hover:bg-opacity-50 relative flex aspect-video w-full cursor-pointer items-center justify-center overflow-hidden rounded-lg border-2 border-dashed bg-cover bg-center bg-no-repeat transition-all duration-300 ease-in-out`}
                                    style={{
                                        backgroundImage: formData.selectedImage
                                            ? `url(${formData.selectedImage})`
                                            : `url("${vod.thumbnailUrl ? vod.thumbnailUrl : `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`}")`,
                                    }}
                                >
                                    <input
                                        id="image-upload"
                                        type="file"
                                        accept="image/*"
                                        onChange={handleImageChange}
                                        className="hidden"
                                    />
                                    <div
                                        className={`absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 transition-opacity duration-200 group-hover:opacity-100`}
                                    >
                                        <span className="text-lg font-medium text-white">
                                            {t(
                                                "settings:vods.edit_dialog.change_thumbnail",
                                            )}
                                        </span>
                                    </div>
                                </label>
                            </div>
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="title">
                                {t("settings:vods.edit_dialog.title_label")}
                            </Label>
                            <Input
                                id="title"
                                name="title"
                                maxLength={VOD_TITLE_MAX_LENGTH}
                                showCount
                                value={formData.title}
                                onChange={handleChange}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="description">
                                {t(
                                    "settings:vods.edit_dialog.description_label",
                                )}
                            </Label>
                            <Input
                                id="description"
                                name="description"
                                type="textarea"
                                maxLength={VOD_DESCRIPTION_MAX_LENGTH}
                                showCount
                                value={formData.description}
                                onChange={handleChange}
                            />
                        </div>
                        <div className="flex items-center space-x-2">
                            <Switch
                                id="isPublic"
                                name="isPublic"
                                checked={formData.isPublic}
                                onCheckedChange={handleSwitchChange}
                            />
                            <Label htmlFor="isPublic">
                                {t("settings:vods.edit_dialog.public")}
                            </Label>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={handleCancel}>
                            {t("settings:vods.edit_dialog.cancel")}
                        </Button>
                        <Button onClick={handleSave}>
                            {isSubmitting ? (
                                <IconLoader className="h-4 w-4" />
                            ) : (
                                <IconSave className="mr-2 h-4 w-4" />
                            )}
                            {t("settings:vods.edit_dialog.save_changes")}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
            <Dialog
                open={isDeleteDialogOpen}
                onOpenChange={setIsDeleteDialogOpen}
            >
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>
                            {t("settings:vods.delete_dialog.title")}
                        </DialogTitle>
                        <DialogDescription>
                            {t("settings:vods.delete_dialog.description")}
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter className="mt-4">
                        <Button
                            variant="outline"
                            onClick={() => setIsDeleteDialogOpen(false)}
                        >
                            {t("settings:vods.delete_dialog.cancel")}
                        </Button>
                        <Button
                            variant="destructive"
                            disabled={isSubmitting}
                            onClick={handleConfirmDelete}
                        >
                            {isSubmitting && (
                                <IconLoader className="mr-2 h-4 w-4" />
                            )}
                            {t("settings:vods.delete_dialog.delete")}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}
