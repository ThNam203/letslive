"use client";

import { useEffect, useMemo, useState } from "react";
import { toast } from "react-toastify";
import { Button } from "@/components/ui/button";
import useUser from "@/hooks/user";
import { UpdateLivestreamInformation } from "@/lib/api/user";
import ImageField from "../_components/image-field";
import Section from "../_components/section";
import TextField from "../_components/text-field";
import TextAreaField from "../_components/textarea-field";
import IconLoader from "@/components/icons/loader";
import useT from "@/hooks/use-translation";

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
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleImageChange = (file: File | null) => {
        if (file) {
            const imageUrl = URL.createObjectURL(file);
            setImage(file);
            setImageUrl(imageUrl);
        }
    };

    const handleResetImage = () => {
        setImage(null);
        setImageUrl(null);
    };

    useEffect(() => {
        if (user) {
            setTitle(user.livestreamInformation.title || "");
            setDescription(user.livestreamInformation.description || "");
            setImageUrl(user.livestreamInformation.thumbnailUrl || null);
            setImage(null);
        }
    }, [user]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;

        setIsSubmitting(true);
        await UpdateLivestreamInformation(
            image === undefined ? null : image,
            image === null ? null : user!.livestreamInformation.thumbnailUrl,
            title,
            description,
        ).then((res) => {
            if (res.success) {
                if (res.data && res.data) {
                    updateUser({
                        ...user,
                        livestreamInformation: {
                            ...user.livestreamInformation,
                            ...res.data,
                        },
                    });
                    toast.success(t("settings:stream.updated_success"));
                }
            } else {
                toast(t(`api-response:${res.key}`), {
                    toastId: res.requestId,
                    type: "error",
                });
            }
        })
        .catch((_) => {
            toast(t("fetch-error:client_fetch_error"), {
                toastId: "client-fetch-error-id",
                type: "error",
            });
        })
        .finally(() => {
            setIsSubmitting(false);
        });
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
            hasBorder
        >
            <form onSubmit={handleSubmit} className="space-y-4">
                <TextField
                    label={t("settings:stream.title_label")}
                    description={t("settings:stream.title_description")}
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                />
                <TextAreaField
                    label={t("settings:stream.description_label")}
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
                        disabled={isSubmitting || !isFormChange}
                        type="submit"
                    >
                        {isSubmitting && <IconLoader />}
                        {t("settings:stream.confirm_edit_button")}
                    </Button>
                </div>
            </form>
        </Section>
    );
}
