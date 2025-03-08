"use client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import useUser from "@/hooks/user";
import { UpdateLivestreamInformation } from "@/lib/api/user";
import { Loader } from "lucide-react";
import { ChangeEvent, useEffect, useState } from "react";
import { toast } from "react-toastify";

export default function StreamEdit() {
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);

    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const [thumbnailUrl, setThumbnailUrl] = useState("");
    const [image, setImage] = useState<File | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const [selectedImage, setSelectedImage] = useState<string | null>(null);
    const handleImageChange = (event: ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (file) {
            const imageUrl = URL.createObjectURL(file);
            setImage(file);
            setSelectedImage(imageUrl);
        }
    };

    useEffect(() => {
        console.log("BUON CUOI NHO:", user)
        if (user) {
            setTitle(user.livestreamInformation.title || "");
            setDescription(user.livestreamInformation.description || "");
            setThumbnailUrl(user.livestreamInformation.thumbnailUrl || "");
        }
    }, [user]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;

        setIsSubmitting(true);
        const { updatedInfo, fetchError } = await UpdateLivestreamInformation(
            image,
            user!.livestreamInformation.thumbnailUrl,
            title,
            description
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

            setSelectedImage(null);
            setImage(null);
        }
    };

    return (
        <div className="min-h-screen max-w-4xl text-gray-900 p-6">
            <div className="space-y-6 mb-4">
                <div className="space-y-1">
                    <h1 className="text-xl font-semibold">Livestream</h1>
                    <p className="text-sm text-gray-400">
                        Your next livestream information will be based on the
                        information.
                    </p>
                </div>
            </div>

            <div className="rounded-lg border-1 border-gray-900 p-4">
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="title">Title</Label>
                        <Input
                            id="title"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            className="col-span-3"
                            required
                        />
                    </div>
                    <div className="grid grid-cols-4 gap-4">
                        <Label htmlFor="description">Description</Label>
                        <Textarea
                            id="description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            className="col-span-3 resize-none"
                            required
                        />
                    </div>

                    <div className="grid grid-cols-4 gap-4">
                        <Label htmlFor="image-upload">Thumbnail</Label>
                        <div className="col-span-3 w-full max-w-3xl">
                            <label
                                htmlFor="image-upload"
                                className={`
                                    group
                                    relative
                                    flex items-center justify-center
                                    w-full aspect-video 
                                    border-2 border-dashed border-gray-300 
                                    rounded-lg 
                                    cursor-pointer
                                    transition-all duration-300 ease-in-out
                                    overflow-hidden
                                    bg-cover bg-center bg-no-repeat
                                    group-hover:bg-opacity-50
                                `}
                                style={{
                                    backgroundImage: selectedImage
                                        ? `url(${selectedImage})`
                                        : `url("${thumbnailUrl}")`,
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
                                    className={`
                                        absolute inset-0
                                        flex items-center justify-center
                                        opacity-0 group-hover:opacity-100
                                        transition-opacity duration-200
                                        bg-black/40
                                    `}
                                >
                                    <span className="text-lg font-medium text-white">
                                        Change thumbnail
                                    </span>
                                </div>
                            </label>
                        </div>
                    </div>
                    <div className="flex justify-end items-center">
                        <Button
                            className="disabled:bg-gray-200 disabled:hover:cursor-not-allowed"
                            disabled={isSubmitting}
                            type="submit"
                        >
                            {isSubmitting && (
                                <Loader className="animate-spin" />
                            )}
                            Confirm edit
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    );
}
