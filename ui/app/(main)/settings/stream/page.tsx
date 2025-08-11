"use client";

import { useEffect, useMemo, useState } from "react";
import { toast } from "react-toastify";
import { Button } from "../../../../components/ui/button";
import useUser from "../../../../hooks/user";
import { UpdateLivestreamInformation } from "../../../../lib/api/user";
import ImageField from "../_components/image-field";
import Section from "../_components/section";
import TextField from "../_components/text-field";
import TextAreaField from "../_components/textarea-field";
import IconLoader from "@/components/icons/loader";

export default function StreamEdit() {
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
        const { updatedInfo, fetchError } = await UpdateLivestreamInformation(
            image === undefined ? null : image,
            image === null ? null : user!.livestreamInformation.thumbnailUrl,
            title,
            description,
        );
        setIsSubmitting(false);
        if (fetchError) {
            toast(fetchError.message, { type: "error" });
            return;
        }

        if (updatedInfo) {
            toast.success("Livestream information updated successfully");
            updateUser({
                ...user,
                livestreamInformation: {
                    userId: updatedInfo.userId,
                    title: updatedInfo.title,
                    description: updatedInfo.description,
                    thumbnailUrl: updatedInfo.thumbnailUrl,
                },
            });
        }
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
            title="Livestream"
            description={`Your next livestream information will be based on the information.\nIt won't change even after livestream ends.`}
            hasBorder
        >
            <form onSubmit={handleSubmit} className="space-y-4">
                <TextField
                    label="Title"
                    description="If empty, the title will be generated automatically."
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                />
                <TextAreaField
                    label="Description"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    rows={4}
                />
                <ImageField
                    label="Thumbnail"
                    description="If empty, the thumbnail will be generated automatically."
                    imageUrl={imageUrl}
                    hoverText="Change thumbnail"
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
                        Confirm edit
                    </Button>
                </div>
            </form>
        </Section>
    );
}
