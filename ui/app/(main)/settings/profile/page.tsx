"use client";

import { Loader } from "lucide-react";
import Image from "next/image";
import { useEffect, useRef, useState } from "react";
import {
    UpdateBackgroundPicture,
    UpdateProfile,
    UpdateProfilePicture,
} from "../../../../lib/api/user";
import { toast } from "react-toastify";
import useUser from "../../../../hooks/user";
import { Button } from "../../../../components/ui/button";
import { Input } from "../../../../components/ui/input";
import { Textarea } from "../../../../components/ui/textarea";

export default function ProfileSettings() {
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);

    const [displayName, setDisplayName] = useState(user?.displayName || "");
    const [bio, setBio] = useState(user?.bio || "");
    const [isButtonDisabled, setIsButtonDisabled] = useState(true);

    const [isUpdatingProfile, setIsUpdatingProfile] = useState(false);
    const [isUpdatingProfilePicture, setIsUpdatingProfilePicture] =
        useState(false);
    const [isUpdatingBackgroundPicture, setIsUpdatingBackgroundPicture] =
        useState(false);
    const profileImageInputRef = useRef<HTMLInputElement>(null);
    const backgroundImageInputRef = useRef<HTMLInputElement>(null);

    const handleProfileUpdateButtonClick = () => {
        profileImageInputRef.current?.click(); // Trigger file input
    };

    const handleBackgroundUpdateButtonClick = () => {
        backgroundImageInputRef.current?.click(); // Trigger file input
    };

    const handleProfileImageChange = async (
        event: React.ChangeEvent<HTMLInputElement>
    ) => {
        if (event.target.files && event.target.files.length > 0) {
            const file = event.target.files[0];

            setIsUpdatingProfilePicture(true);
            const { newPath, fetchError } = await UpdateProfilePicture(file);
            setIsUpdatingProfilePicture(false);
            if (fetchError) {
                toast(fetchError.message, { type: "error" });
            } else {
                updateUser({
                    ...user!,
                    profilePicture: newPath!,
                });

                toast("Profile picture updated successfully!", {
                    type: "success",
                });
            }
        }
    };

    const handleBackgroundImageChange = async (
        event: React.ChangeEvent<HTMLInputElement>
    ) => {
        if (event.target.files && event.target.files.length > 0) {
            const file = event.target.files[0];

            setIsUpdatingBackgroundPicture(true);
            const { newPath, fetchError } = await UpdateBackgroundPicture(file);
            setIsUpdatingBackgroundPicture(false);
            if (fetchError) {
                toast(fetchError.message, { type: "error" });
            } else {
                updateUser({
                    ...user!,
                    backgroundPicture: newPath!,
                });

                toast("Background picture updated successfully!", {
                    type: "success",
                });
            }
        }
    };

    const handleUpdateProfileInformation = async (
        event: React.FormEvent<HTMLFormElement>
    ) => {
        event.preventDefault();
        setIsUpdatingProfile(true);

        const { updatedUser, fetchError } = await UpdateProfile({
            ...user!,
            displayName,
            bio,
        });

        if (fetchError) {
            toast(fetchError.message, { type: "error" });
        } else {
            toast("Profile information updated successfully!", {
                type: "success",
            });

            updateUser(updatedUser!);
        }

        setIsUpdatingProfile(false);
    };

    useEffect(() => {
        if (!user) return; 

        const normalizedDisplayName = user.displayName ?? "";
        const normalizedBio = user.bio ?? "";

        if (normalizedDisplayName !== displayName || normalizedBio !== bio) {
            setIsButtonDisabled(false);
        } else {
            setIsButtonDisabled(true);
        }
    }, [displayName, bio]);

    useEffect(() => {
        setDisplayName(user?.displayName ?? "");
        setBio(user?.bio ?? "");
    }, [user]);

    return (
        <div className="min-h-screen text-gray-900 p-6 overflow-y-auto">
            <div className="max-w-4xl space-y-8">
                {/* Profile Picture Section */}
                <section>
                    <h2 className="text-xl font-semibold mb-6">
                        Profile Picture
                    </h2>
                    <div className="flex gap-4 items-center">
                        <div className="relative w-20 h-20 rounded-full overflow-hidden">
                            <Image
                                src={
                                    user?.profilePicture ??
                                    "https://github.com/shadcn.png"
                                }
                                alt="Profile Picture"
                                layout="fill"
                                objectFit="cover"
                            />
                        </div>
                        <div className="space-y-2">
                            <div>
                                <input
                                    type="file"
                                    ref={profileImageInputRef}
                                    className="hidden"
                                    onChange={handleProfileImageChange}
                                />
                                <Button
                                    variant="secondary"
                                    disabled={isUpdatingProfilePicture}
                                    className="bg-gray-300 hover:bg-gray-500 text-gray-900"
                                    onClick={handleProfileUpdateButtonClick}
                                >
                                    {isUpdatingProfilePicture && (
                                        <Loader className="animate-spin" />
                                    )}{" "}
                                    Update picture
                                </Button>
                            </div>
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
                                {user && user.backgroundPicture ? (
                                    <Image
                                        src={user.backgroundPicture}
                                        alt="Profile Banner"
                                        layout="fill"
                                        objectFit="cover"
                                        className="rounded-lg"
                                    />
                                ) : (
                                    generateBackground()
                                )}
                            </div>
                            <div className="absolute space-y-2 right-3 bottom-1/2 translate-y-1/2">
                                <input
                                    type="file"
                                    ref={backgroundImageInputRef}
                                    className="hidden"
                                    onChange={handleBackgroundImageChange}
                                />
                                <Button
                                    variant="secondary"
                                    className="bg-gray-300 hover:bg-gray-500 text-gray-900"
                                    onClick={handleBackgroundUpdateButtonClick}
                                    disabled={isUpdatingBackgroundPicture}
                                >
                                    {isUpdatingBackgroundPicture && (
                                        <Loader className="animate-spin" />
                                    )}{" "}
                                    Update background
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

                    <form
                        className="space-y-6"
                        onSubmit={handleUpdateProfileInformation}
                    >
                        <div>
                            <label className="block text-sm font-medium mb-2">
                                Username
                            </label>
                            <div className="relative">
                                <Input
                                    type="text"
                                    readOnly
                                    defaultValue={user?.username}
                                    className="text-gray-900 border-gray-700 pr-10"
                                />
                                {/* <Button
                                    size="icon"
                                    variant="ghost"
                                    className="absolute right-2 top-1/2 -translate-y-1/2 hover:bg-gray-400 hover:bg-opacity-40 rounded-3xl"
                                >
                                    <Pencil className="h-2 w-2" />
                                </Button> */}
                            </div>
                            <p className="text-sm mt-1 text-red-500">
                                You can&apos;t update your username now.
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium mb-2">
                                Display Name
                            </label>
                            <Input
                                type="text"
                                defaultValue={user?.displayName}
                                value={displayName}
                                onChange={(e) => setDisplayName(e.target.value)}
                                disabled={isUpdatingProfile}
                                className="text-gray-900 border-gray-700"
                            />
                            <p className="text-sm text-gray-400 mt-1">
                                Create an alternative for your username
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium mb-2">
                                Bio
                            </label>
                            <Textarea
                                className="text-gray-900 border-gray-700 min-h-[100px]"
                                defaultValue={user?.bio}
                                value={bio}
                                onChange={(e) => setBio(e.target.value)}
                                disabled={isUpdatingProfile}
                            />
                            <p className="text-sm text-gray-400 mt-1">
                                Give everyone a description your channel in
                                under 300 characters
                            </p>
                        </div>

                        <div className="flex justify-end mt-6">
                            <Button
                                className="bg-purple-600 hover:bg-purple-700 text-white disabled:bg-gray-400"
                                disabled={isUpdatingProfile || isButtonDisabled}
                                type="submit"
                            >
                                {isUpdatingProfile && (
                                    <Loader className="animate-spin" />
                                )}{" "}
                                Save Changes
                            </Button>
                        </div>
                    </form>
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

const generateBackground = () => {
    return (
        <>
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
        </>
    );
};
