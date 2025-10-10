"use client";

import Image from "next/image";
import Link from "next/link";
import useT from "@/hooks/use-translation";
import { dateDiffFromNow, formatSeconds } from "@/utils/timeFormats";
import type React from "react";
import { ChangeEvent, Dispatch, SetStateAction, useState } from "react";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
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
import { DeleteVOD, UpdateVOD } from "@/lib/api/vod";
import { toast } from "react-toastify";
import { UploadFile } from "@/lib/api/utils";
import GLOBAL from "@/global";
import IconEye from "@/components/icons/eye";
import IconEyeOff from "@/components/icons/eye-off";
import IconDotsVertical from "@/components/icons/dots-vertical";
import IconSave from "@/components/icons/save";
import { VOD } from "@/types/vod";
import IconLoader from "@/components/icons/loader";

export default function VODEditCard({
    vod,
    setVODS,
}: {
    vod: VOD;
    setVODS: Dispatch<SetStateAction<VOD[]>>;
}) {
    const { t } = useT(["settings", "api-response"]);
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
            <div
                className={`w-[350px] overflow-hidden rounded-sm bg-gray-200 shadow-sm`}
            >
                <Link
                    className={`inline-block h-[180px] w-full hover:cursor-pointer`}
                    href={`/users/${vod.userId}/vods/${vod.id}`}
                >
                    <div className="flex h-full flex-col items-center justify-center bg-black bg-opacity-50">
                        <Image
                            alt="vod icon"
                            src={
                                vod.thumbnailUrl ??
                                `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`
                            }
                            width={350}
                            height={180}
                            className="h-full w-full"
                        />
                    </div>
                </Link>
                <div className="p-4">
                    <h3 className="font-semibold text-foreground">
                        {vod.title}
                    </h3>
                    <p className="mt-1 text-sm text-gray-500">
                        {formatSeconds(vod.duration)} -{" "}
                        {vod.visibility === "public" ? (
                            <IconEye className="mr-1 inline-block h-4 w-4" />
                        ) : (
                            <IconEyeOff className="mr-1 inline-block h-4 w-4" />
                        )}
                    </p>
                    <p className="mt-1 text-sm text-foreground-muted">
                        {vod.description && vod.description.length > 50
                            ? `${vod.description.substring(0, 47)}...`
                            : vod.description}{" "}
                        • {dateDiffFromNow(vod.createdAt, t)}
                    </p>
                    <div className="mt-2 flex items-center text-sm text-foreground-muted">
                        <span>
                            {vod.viewCount}{" "}
                            {t(
                                `settings:vods.metadata.${vod.viewCount === 1 ? "view" : "views"}`,
                            )}
                        </span>
                        <div className="flex-1" />
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon">
                                    <IconDotsVertical className="h-5 w-5" />
                                    <span className="sr-only">Open menu</span>
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={handleEdit}>
                                    {t("settings:vods.actions.edit")}
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={handleDelete}>
                                    {t("settings:vods.actions.delete")}
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>
                </div>
            </div>

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
                                    className={`group relative flex aspect-video w-full cursor-pointer items-center justify-center overflow-hidden rounded-lg border-2 border-dashed border-border bg-cover bg-center bg-no-repeat transition-all duration-300 ease-in-out group-hover:bg-opacity-50`}
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
