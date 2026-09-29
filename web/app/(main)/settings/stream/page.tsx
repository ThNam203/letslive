"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "@/components/utils/toast";
import { Button } from "@/components/ui/button";
import useUser from "@/hooks/user";
import { UpdateLivestreamInformation } from "@/lib/api/user";
import { unwrapResponse } from "@/lib/api/api-error";
import ImageField from "../_components/image-field";
import Section from "../_components/section";
import TextField from "../_components/text-field";
import TextAreaField from "../_components/textarea-field";
import IconLoader from "@/components/icons/loader";
import useT from "@/hooks/use-translation";
import {
    STREAM_TITLE_MAX_LENGTH,
    STREAM_DESCRIPTION_MAX_LENGTH,
} from "@/constant/field-limits";

export default function StreamEdit() {
    const { t } = useT(["settings", "api-response", "fetch-error"]);
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);

    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");

    // use undefined to indicate no image initially
    // use null to indicate that user has reset the image
    const [image, setImage] = useState<File | null | undefined>(undefined);
    const [imageUrl, setImageUrl] = useState<string | null>(null);
    const blobUrlRef = useRef<string | null>(null);

    useEffect(() => {
        return () => {
            if (blobUrlRef.current) URL.revokeObjectURL(blobUrlRef.current);
        };
    }, []);

    const handleImageChange = (file: File | null) => {
        if (file) {
            if (blobUrlRef.current) URL.revokeObjectURL(blobUrlRef.current);
            const url = URL.createObjectURL(file);
            blobUrlRef.current = url;
            setImage(file);
            setImageUrl(url);
        }
    };

    const handleResetImage = () => {
        if (blobUrlRef.current) {
            URL.revokeObjectURL(blobUrlRef.current);
            blobUrlRef.current = null;
        }
        setImage(null);
        setImageUrl(null);
    };

    useEffect(() => {
        if (!user) return;
        queueMicrotask(() => {
            setTitle(user.livestreamInformation.title || "");
            setDescription(user.livestreamInformation.description || "");
            setImageUrl(user.livestreamInformation.thumbnailUrl || null);
            setImage(null);
        });
    }, [user]);

    const updateStreamInfoMutation = useMutation({
        mutationFn: async () =>
            unwrapResponse(
                await UpdateLivestreamInformation(
                    image === undefined ? null : image,
                    image === null
                        ? null
                        : user!.livestreamInformation.thumbnailUrl,
                    title,
                    description,
                ),
            ),
        onSuccess: (data) => {
            if (!user || !data) return;
            if (blobUrlRef.current) {
                URL.revokeObjectURL(blobUrlRef.current);
                blobUrlRef.current = null;
            }
            updateUser({
                ...user,
                livestreamInformation: {
                    ...user.livestreamInformation,
                    ...data,
                },
            });
            toast.success(t("settings:stream.updated_success"));
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;
        updateStreamInfoMutation.mutate();
    };

    const isFormChange = useMemo(() => {
        return (
            title !== user?.livestreamInformation.title ||
            description !== user?.livestreamInformation.description ||
            imageUrl !== user?.livestreamInformation.thumbnailUrl
        );
    }, [title, description, imageUrl, user]);

    return (
        <Section
            title={t("settings:stream.title")}
            description={t("settings:stream.description")}
            contentClassName="p-4"
        >
            <form onSubmit={handleSubmit} className="space-y-4">
                <TextField
                    label={t("settings:stream.title_label")}
                    description={t("settings:stream.title_description")}
                    maxLength={STREAM_TITLE_MAX_LENGTH}
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                />
                <TextAreaField
                    label={t("settings:stream.description_label")}
                    maxLength={STREAM_DESCRIPTION_MAX_LENGTH}
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    rows={4}
                />
                <ImageField
                    label={t("settings:stream.thumbnail_label")}
                    description={t("settings:stream.thumbnail_description")}
                    imageUrl={imageUrl}
                    hoverText={t("settings:stream.thumbnail_hover")}
                    onImageChange={handleImageChange}
                    onResetImage={handleResetImage}
                    showCloseIcon={imageUrl !== null}
                />
                <div className="flex items-center justify-end">
                    <Button
                        disabled={
                            updateStreamInfoMutation.isPending || !isFormChange
                        }
                        type="submit"
                    >
                        {updateStreamInfoMutation.isPending && <IconLoader />}
                        {t("settings:stream.confirm_edit_button")}
                    </Button>
                </div>
            </form>
        </Section>
    );
}
