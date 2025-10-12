"use client";

import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { Button } from "@/components/ui/button";
import useUser from "@/hooks/user";
import {
    UpdateBackgroundPicture,
    UpdateProfile,
    UpdateProfilePicture,
} from "@/lib/api/user";
import TextField from "../_components/text-field";
import ProfileBanner from "./_components/profile-banner";
import Section from "../_components/section";
import TextAreaField from "../_components/textarea-field";
import ThemeList from "@/app/[lng]/(main)/settings/profile/_components/theme-list";
import IconLoader from "@/components/icons/loader";
import DisableAccountDialog from "./_components/disable-account-dialog";
import useT from "@/hooks/use-translation";
import LanguageList from "@/app/[lng]/(main)/settings/profile/_components/language-list";
import { SocialMediaEdit } from "@/app/[lng]/(main)/settings/profile/_components/socials-media-link";

export default function ProfileSettings() {
    const { t } = useT(["settings", "common"]);
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);

    const [displayName, setDisplayName] = useState("");
    const [isDisplayNameChanged, setIsDisplayNameChanged] = useState(false);
    const [bio, setBio] = useState("");
    const [isBioChanged, setIsBioChanged] = useState(false);
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
        let hasError = false;

        if (backgroundImageFile) {
            await UpdateBackgroundPicture(backgroundImageFile)
            .then(res => {
                if (res.success) {
                    setBackgroundImageFile(null);
                    updateUser({
                        ...user!,
                        backgroundPicture: res.data,
                    });
                } else {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                    hasError = true;
                }
            })
            .catch((_) => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
                hasError = true;
            })
            .finally(() =>
                setIsUpdatingProfile(false),
            );
        }

        if (profileImageFile) {
                await UpdateProfilePicture(profileImageFile)
                .then(res => {
                    if (res.success) {
                        setProfileImageFile(null);
                        updateUser({
                            ...user!,
                            profilePicture: res.data,
                        });
                    } else {
                        toast.error(t(`api-response:${res.key}`), {
                            toastId: res.requestId,
                            type: "error",
                        });
                        hasError = true;
                    }
                })
                .catch((_) => {
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                    hasError = true;
                })
                .finally(() =>
                    setIsUpdatingProfile(false),
                );
        }

        if (bio || displayName) {
            await UpdateProfile({
                displayName: isDisplayNameChanged ? displayName : undefined,
                bio: isBioChanged ? bio : undefined,
            })
            .then(res => {
                if (res.success) {
                    updateUser({
                        ...user!,
                        ...res.data,
                    });
                } else {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                    hasError = true;
                }
            })
            .catch((_) => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
                hasError = true;
            })
            .finally(() => setIsUpdatingProfile(false));
        }

        if (!hasError) toast.success(t("settings:profile.update_success"));
    };

    useEffect(() => {
        if (!user) return;

        const isDisplayNameChange = (user.displayName ?? "") !== displayName;
        const isBioChange = (user.bio ?? "") !== bio;
        const isProfileImageChange = profileImageFile !== null;
        const isBackgroundImageChange = backgroundImageFile !== null;

        setIsDisplayNameChanged(isDisplayNameChange);
        setIsBioChanged(isBioChange);

        const isUserDataChange =
            isDisplayNameChange ||
            isBioChange ||
            isProfileImageChange ||
            isBackgroundImageChange;

        if (isUserDataChange) setIsButtonDisabled(false);
        else setIsButtonDisabled(true);
    }, [displayName, bio, user, profileImageFile, backgroundImageFile]);

    useEffect(() => {
        return () => {
            try {
                if (profileImageFile) URL.revokeObjectURL(profileImageFile?.name);
                if (backgroundImageFile)
                    URL.revokeObjectURL(backgroundImageFile?.name);
            } catch (e) {}
        }
    }, []);

    useEffect(() => {
        setDisplayName(user?.displayName ?? "");
        setBio(user?.bio ?? "");
    }, [user]);

    return (
        <>
            {/* Profile Settings Section */}
            <Section
                title={t("settings:profile.title")}
                description={t("settings:profile.description")}
                contentClassName="p-4"
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
                        label={t("settings:profile.username")}
                        disabled
                        value={user?.username}
                    />
                    <TextField
                        label={t("settings:profile.display_name")}
                        value={displayName}
                        onChange={(e) => setDisplayName(e.target.value)}
                        disabled={isUpdatingProfile}
                    />
                    <TextAreaField
                        label={t("settings:profile.bio")}
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
                            {isUpdatingProfile && <IconLoader />} {t("common:save_changes")}
                        </Button>
                    </div>
                </form>
            </Section>

            <Section
                title={"Edit Social Media Links"}
                description={"Click on a platform to add or update your profile link"}
                className="border-t border-border pt-8"
                contentClassName="p-4"
            >
                <SocialMediaEdit />
            </Section>

            <Section
                title={t("settings:themes.title")}
                description={t("settings:themes.description")}
                className="border-t border-border pt-8"
                contentClassName="p-4"
            >
                <ThemeList />
            </Section>

            <Section
                title={t("settings:language.title")}
                description={t("settings:language.description")}
                className="border-t border-border pt-8"
                contentClassName="p-4"
            >
                <LanguageList />
            </Section>

            <Section
                title={t("settings:disable.title")}
                description={t("settings:disable.description")}
                className="border-t border-border pt-8"
                contentClassName="p-4"
            >
                <div className="flex w-full items-center justify-between">
                    <p className="w-2/3 text-sm text-destructive">
                        {t("settings:disable.note")}
                    </p>
                    <DisableAccountDialog isUpdatingProfile={isUpdatingProfile} />
                </div>
            </Section>
        </>
    );
}
