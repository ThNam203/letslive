"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Pencil } from "lucide-react";
import Link from "next/link";
import { User } from "@/types/user";
import { UpdateProfile } from "@/lib/api/user";
import { toast } from "react-toastify";

export default function ProfileSettings() {
    const [user, setUser] = useState<User | undefined>(undefined);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;

        const updatedUserResponse = await UpdateProfile({
            id: user.id,
            username: user.username,
            bio: user.bio,
        });

        if (updatedUserResponse.fetchError) {
            toast.error(updatedUserResponse.fetchError.message);
            return;
        }

        setUser({
            ...user,
            username: updatedUserResponse.user!.username,
            bio: updatedUserResponse.user!.bio,
        });

        toast.success("Profile updated successfully");
    };

    return (
        <div className="min-h-screen text-gray-900 p-6 overflow-y-auto">
            <div className="max-w-4xl space-y-8">
                {/* Profile Picture Section */}
                <section>
                    <h2 className="text-xl font-semibold mb-6">
                        Profile Picture
                    </h2>
                    <div className="flex gap-4 items-center">
                        <div className="relative w-20 h-20 rounded-full overflow-hidden bg-[#00ff00]">
                            <div className="absolute inset-0 flex items-center justify-center">
                                <svg
                                    className="w-12 h-12 text-black"
                                    viewBox="0 0 24 24"
                                    fill="currentColor"
                                >
                                    <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" />
                                </svg>
                            </div>
                        </div>
                        <div className="space-y-2">
                            <Button
                                variant="secondary"
                                className="bg-gray-300 hover:bg-gray-500 text-gray-900"
                            >
                                Add Profile Picture
                            </Button>
                            <p className="text-sm text-gray-400">
                                Must be JPEG, PNG, or GIF and cannot exceed
                                10MB.
                            </p>
                        </div>
                    </div>
                </section>

                {/* Profile Banner Section */}
                <section>
                    <h2 className="text-xl font-semibold mb-6">
                        Profile Banner
                    </h2>
                    <div className="space-y-4">
                        <div className="relative w-full h-40 rounded-lg overflow-hidden border-1 border-gray-900">
                            <div className="absolute inset-0 grid grid-cols-6 gap-2 p-2 w-1/2 m-2 ml-8 border-[1px] bg-gray-800 border-gray-600 rounded-lg">
                                {[...Array(18)].map((_, i) => (
                                    <svg
                                        key={i}
                                        className="w-8 h-8 text-white opacity-25"
                                        viewBox="0 0 24 24"
                                        fill="currentColor"
                                    >
                                        <path d="M21 3H3v18h18V3zm-9 14H7v-4h5v4zm0-6H7V7h5v4zm6 6h-4v-4h4v4zm0-6h-4V7h4v4z" />
                                    </svg>
                                ))}
                            </div>
                            <div className="absolute space-y-2 right-3 bottom-1/2 translate-y-1/2">
                                <Button
                                    variant="secondary"
                                    className="bg-gray-300 hover:bg-gray-500 text-gray-900"
                                >
                                    Update
                                </Button>
                                <p className="text-sm text-gray-400">
                                    File format: JPEG, PNG, GIF and cannot
                                    exceed 10MB
                                </p>
                            </div>
                        </div>
                    </div>
                </section>

                {/* Profile Settings Section */}
                <section className="p-8 border-gray-600 border-[1px] rounded-lg">
                    <h2 className="text-xl font-semibold mb-1">
                        Profile Settings
                    </h2>
                    <p className="text-sm text-gray-400 mb-6">
                        Change identifying details for your account
                    </p>

                    <div className="space-y-6">
                        <div>
                            <label className="block text-sm font-medium mb-2">
                                Username
                            </label>
                            <div className="relative">
                                <Input
                                    type="text"
                                    defaultValue="ssenlor203"
                                    className="text-gray-900 border-gray-700 pr-10"
                                />
                                <Button
                                    size="icon"
                                    variant="ghost"
                                    className="absolute right-2 top-1/2 -translate-y-1/2 hover:bg-gray-400 hover:bg-opacity-40 rounded-3xl"
                                >
                                    <Pencil className="h-2 w-2" />
                                </Button>
                            </div>
                            <p className="text-sm text-gray-400 mt-1">
                                You may update your username
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium mb-2">
                                Display Name
                            </label>
                            <Input
                                type="text"
                                defaultValue="ssenlor203"
                                className="text-gray-900 border-gray-700"
                            />
                            <p className="text-sm text-gray-400 mt-1">
                                Customize capitalization for your username
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium mb-2">
                                Bio
                            </label>
                            <Textarea className="text-gray-900 border-gray-700 min-h-[100px]" />
                            <p className="text-sm text-gray-400 mt-1">
                                Description for the About panel on your channel
                                page in under 300 characters
                            </p>
                        </div>
                    </div>

                    <div className="flex justify-end mt-6">
                        <Button className="bg-purple-600 hover:bg-purple-700">
                            Save Changes
                        </Button>
                    </div>
                </section>

                {/* Disable Account Section */}
                <section className="border-t border-gray-800 pt-8">
                    <h2 className="text-xl font-semibold mb-1">
                        Disabling your Let&apos;s Live account
                    </h2>
                    <p className="text-sm text-gray-400 mb-4">
                        Completely deactivate your account
                    </p>

                    <div className="border-1 rounded-md p-4">
                        <div className="flex items-center justify-between">
                            <p className="text-sm text-gray-800">
                                When you disable your account, your profile and
                                notifications will be hidden, and your account
                                will be deactivated. You can reactivate your
                                account at any time.
                            </p>
                            <button className="text-sm font-medium text-white bg-red-800 px-4 py-2 rounded-md hover:bg-red-700">
                                Disable Your Let&apos;s Live Account
                            </button>
                        </div>
                    </div>
                </section>
            </div>
        </div>
    );
}
