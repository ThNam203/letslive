"use client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useState } from "react";

export default function StreamEdit() {
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const [image, setImage] = useState<File | null>(null);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
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
                        <Label htmlFor="title">
                            Title
                        </Label>
                        <Input
                            id="title"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            className="col-span-3"
                            required
                        />
                    </div>
                    <div className="grid grid-cols-4 gap-4">
                        <Label htmlFor="description">
                            Description
                        </Label>
                        <Textarea
                            id="description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            className="col-span-3 resize-none"
                            required
                        />
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="image">
                            Thumbnail
                        </Label>
                        <Input
                            id="image"
                            type="file"
                            accept="image/*"
                            onChange={(e) =>
                                setImage(e.target.files?.[0] || null)
                            }
                            className="col-span-3 pt-[6px] text-xs hover:cursor-pointer "
                            required
                        />
                    </div>
                    <div className="flex justify-end">
                        <Button type="submit">Confirm edit</Button>
                    </div>
                </form>
            </div>
        </div>
    );
}
