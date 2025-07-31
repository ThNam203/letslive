"use client";

import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { Button } from "../../../../components/ui/button";
import useUser from "../../../../hooks/user";
import {
    UpdateBackgroundPicture,
    UpdateProfile,
    UpdateProfilePicture,
} from "../../../../lib/api/user";
import TextField from "../_components/text-field";
import ProfileBanner from "./_components/profile-banner";
import Section from "../_components/section";
import TextAreaField from "../_components/textarea-field";
import ThemeSwitch from "@/components/utils/theme-switch";
import ThemeList from "@/components/utils/theme-list";
import IconLoader from "@/components/icons/loader";

export default function ProfileSettings() {
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);

    const [displayName, setDisplayName] = useState("");
    const [bio, setBio] = useState("");
    const [isButtonDisabled, setIsButtonDisabled] = useState(true);
    const [isUpdatingProfile, setIsUpdatingProfile] = useState(false);
    const [profileImageFile, setProfileImageFile] = useState<File | null>(null);
    const [backgroundImageFile, setBackgroundImageFile] = useState<File | null>(
        null,
    );

    const handleProfileImageChange = (file: File | null) => {
        setProfileImageFile(file);
    };

    const handleBackgroundImageChange = (file: File | null) => {
        setBackgroundImageFile(file);
    };

    const handleUpdateProfileInformation = async (
        event: React.FormEvent<HTMLFormElement>,
    ) => {
        event.preventDefault();
        setIsUpdatingProfile(true);

        if (backgroundImageFile) {
            const {
                fetchError: backgroundImageError,
                newPath: backgroundImagePath,
            } = await UpdateBackgroundPicture(backgroundImageFile).finally(() =>
                setIsUpdatingProfile(false),
            );
            if (backgroundImageError) {
                toast.error(backgroundImageError.message);
                return;
            }
            setBackgroundImageFile(null);
            updateUser({
                ...user!,
                backgroundPicture: backgroundImagePath,
            });
        }

        if (profileImageFile) {
            const { fetchError: profileImageError, newPath: profileImagePath } =
                await UpdateProfilePicture(profileImageFile).finally(() =>
                    setIsUpdatingProfile(false),
                );
            if (profileImageError) {
                toast.error(profileImageError.message);
                return;
            }
            setProfileImageFile(null);
            updateUser({
                ...user!,
                profilePicture: profileImagePath,
            });
        }

        if (bio || displayName) {
            const { updatedUser, fetchError } = await UpdateProfile({
                ...user!,
                displayName,
                bio,
            }).finally(() => setIsUpdatingProfile(false));
            if (fetchError) {
                toast.error(fetchError.message);
                return;
            }
            updateUser({
                ...user!,
                ...updatedUser!,
            });
        }

        toast.success("Profile information updated successfully!");
    };

    useEffect(() => {
        if (!user) return;

        const isDisplayNameChange = (user.displayName ?? "") !== displayName;
        const isBioChange = (user.bio ?? "") !== bio;
        const isProfileImageChange = profileImageFile !== null;
        const isBackgroundImageChange = backgroundImageFile !== null;

        console.log(
            isDisplayNameChange,
            isBioChange,
            isProfileImageChange,
            isBackgroundImageChange,
        );

        const isUserDataChange =
            isDisplayNameChange ||
            isBioChange ||
            isProfileImageChange ||
            isBackgroundImageChange;

        if (isUserDataChange) setIsButtonDisabled(false);
        else setIsButtonDisabled(true);
    }, [displayName, bio, user, profileImageFile, backgroundImageFile]);

    useEffect(() => {
        setDisplayName(user?.displayName ?? "");
        setBio(user?.bio ?? "");
    }, [user]);

    return (
        <>
            {/* Profile Settings Section */}
            <Section
                title="Profile Settings"
                description="Change identifying details for your account"
                hasBorder
            >
                <form
                    className="space-y-6"
                    onSubmit={handleUpdateProfileInformation}
                >
                    <ProfileBanner
                        className="mb-10"
                        onProfileImageChange={handleProfileImageChange}
                        onBackgroundImageChange={handleBackgroundImageChange}
                    />
                    <TextField
                        label="Username"
                        disabled
                        defaultValue={user?.username}
                    />
                    <TextField
                        label="Display Name"
                        defaultValue={user?.displayName}
                        value={displayName}
                        onChange={(e) => setDisplayName(e.target.value)}
                        disabled={isUpdatingProfile}
                    />
                    <TextAreaField
                        label="Bio"
                        defaultValue={user?.bio}
                        value={bio}
                        onChange={(e) => setBio(e.target.value)}
                        disabled={isUpdatingProfile}
                        className="min-h-[100px]"
                    />

                    <div className="mt-6 flex justify-end">
                        <Button
                            disabled={isUpdatingProfile || isButtonDisabled}
                            type="submit"
                        >
                            {isUpdatingProfile && <IconLoader />} Save Changes
                        </Button>
                    </div>
                </form>
            </Section>

            <Section
                title="Themes"
                description="Customize the look and feel of the website"
                className="border-t border-border pt-8"
                contentClassName="border border-border rounded-md p-4"
            >
                <ThemeList />
            </Section>

            <Section
                title="Disabling your Let's Live account"
                description="Completely deactivate your account"
                className="border-t border-border pt-8"
                contentClassName="border border-border rounded-md p-4"
            >
                <div className="flex w-full items-center justify-between">
                    <p className="w-2/3 text-sm text-destructive">
                        When you disable your account, your profile and
                        notifications will be hidden, and your account will be
                        deactivated. You can reactivate your account at any
                        time.
                    </p>
                    <button className="rounded-md bg-destructive px-4 py-2 text-sm font-medium text-destructive-foreground hover:bg-destructive-hover">
                        Disable Your Let&apos;s Live Account
                    </button>
                </div>
            </Section>
        </>
    );
}
