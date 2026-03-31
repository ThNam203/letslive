"use client";

import useT from "@/hooks/use-translation";
import type React from "react";
import { ChangeEvent, Dispatch, SetStateAction, useState } from "react";
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
import { DeleteVOD, UpdateVOD } from "@/lib/api/vod";
import { toast } from "@/components/utils/toast";
import { UploadFile } from "@/lib/api/utils";
import GLOBAL from "@/global";
import IconSave from "@/components/icons/save";
import { VOD } from "@/types/vod";
import {
    VOD_TITLE_MAX_LENGTH,
    VOD_DESCRIPTION_MAX_LENGTH,
} from "@/constant/field-limits";
import IconLoader from "@/components/icons/loader";
import VODCard from "@/components/livestream/vod-card";

export default function VODEditCard({
    vod,
    setVODS,
}: {
    vod: VOD;
    setVODS: Dispatch<SetStateAction<VOD[]>>;
}) {
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

    const [isSubmitting, setIsSubmitting] = useState(false);
    const handleImageChange = (event: ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (file) {
            const imageUrl = URL.createObjectURL(file);
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

    const handleConfirmDelete = async () => {
        setIsSubmitting(true);
        await DeleteVOD(vod.id).then((res) => {
            if (!res.success) {
                toast(t(`api-response:${res.key}`), { type: "error" });
            }
        });

        setIsDeleteDialogOpen(false);
        setIsSubmitting(false);
        setVODS((prev) => prev.filter((v) => v.id !== vod.id));
    };

    const handleSave = async () => {
        setIsSubmitting(true);
        var newThumbnailPath = "";

        if (formData.image) {
            await UploadFile(formData.image).then((res) => {
                if (!res.success) {
                    toast(t(`api-response:${res.key}`), { type: "error" });
                    setIsSubmitting(false);
                    setIsDialogOpen(false);
                } else {
                    newThumbnailPath = res.data?.newPath!;
                }
            });
        }

        await UpdateVOD(
            vod.id,
            formData.title,
            formData.description,
            formData.isPublic ? "public" : "private",
            newThumbnailPath.length > 0 ? newThumbnailPath : undefined,
        )
            .then((res) => {
                if (!res.success) {
                    toast(t(`api-response:${res.key}`), { type: "error" });
                    setIsSubmitting(false);
                    setIsDialogOpen(false);
                } else {
                    toast(t("settings:vods.edit_dialog.update_success"), {
                        type: "success",
                    });
                    setVODS((prev) =>
                        prev.map((v) =>
                            v.id === vod.id
                                ? {
                                      ...v,
                                      title: formData.title,
                                      description: formData.description,
                                      visibility: formData.isPublic
                                          ? "public"
                                          : "private",
                                      thumbnailUrl:
                                          newThumbnailPath.length > 0
                                              ? newThumbnailPath
                                              : v.thumbnailUrl
                                                ? v.thumbnailUrl
                                                : `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`,
                                  }
                                : v,
                        ),
                    );
                }
            })
            .finally(() => {
                setIsSubmitting(false);
                setIsDialogOpen(false);
            });
    };

    const handleCancel = () => {
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
            <VODCard
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
                            onClick={handleConfirmDelete}
                        >
                            {t("settings:vods.delete_dialog.delete")}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}
