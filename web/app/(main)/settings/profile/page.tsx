"use client";

import { useEffect, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "@/components/utils/toast";
import { Button } from "@/components/ui/button";
import useUser from "@/hooks/user";
import {
    UpdateBackgroundPicture,
    UpdateProfile,
    UpdateProfilePicture,
} from "@/lib/api/user";
import TextField from "../_components/text-field";
import ProfileBanner, {
    type PendingAvatar,
    type PendingImage,
} from "./_components/profile-banner";
import Section from "../_components/section";
import TextAreaField from "../_components/textarea-field";
import ThemeList from "@/app/(main)/settings/profile/_components/theme-list";
import IconLoader from "@/components/icons/loader";
import DisableAccountDialog from "./_components/disable-account-dialog";
import useT from "@/hooks/use-translation";
import { USERNAME_MAX_LENGTH, BIO_MAX_LENGTH } from "@/constant/field-limits";
import LanguageList from "@/app/(main)/settings/profile/_components/language-list";
import { SocialMediaEdit } from "@/app/(main)/settings/profile/_components/socials-media-link";

export default function ProfileSettings() {
    const { t } = useT(["settings", "common"]);
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);

    const [username, setUsername] = useState("");
    const [bio, setBio] = useState("");
    const [pendingAvatar, setPendingAvatar] = useState<PendingAvatar | null>(
        null,
    );
    const [pendingBackground, setPendingBackground] =
        useState<PendingImage | null>(null);

    const isUsernameChanged = user != null && user.username !== username;
    const isBioChanged = user != null && (user.bio ?? "") !== bio;
    const isButtonDisabled =
        !isUsernameChanged &&
        !isBioChanged &&
        pendingAvatar === null &&
        pendingBackground === null;

    const updateProfileMutation = useMutation({
        mutationFn: async () => {
            let hasError = false;

            if (pendingBackground) {
                const res = await UpdateBackgroundPicture(
                    pendingBackground.file,
                );
                if (res.success) {
                    setPendingBackground(null);
                    updateUser({ ...user!, backgroundPicture: res.data });
                } else {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                    hasError = true;
                }
            }

            if (pendingAvatar) {
                const res = await UpdateProfilePicture(
                    pendingAvatar.file,
                    pendingAvatar.crop,
                );
                if (res.success) {
                    setPendingAvatar(null);
                    updateUser({ ...user!, profilePicture: res.data });
                } else {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                    hasError = true;
                }
            }

            if (isUsernameChanged || isBioChanged) {
                const res = await UpdateProfile({
                    username: isUsernameChanged ? username : undefined,
                    bio: isBioChanged ? bio : undefined,
                });
                if (res.success) {
                    updateUser({ ...user!, ...res.data });
                } else {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                    hasError = true;
                }
            }

            return hasError;
        },
        onSuccess: (hasError) => {
            if (!hasError) toast.success(t("settings:profile.update_success"));
        },
    });

    const handleUpdateProfileInformation = (
        event: React.FormEvent<HTMLFormElement>,
    ) => {
        event.preventDefault();
        updateProfileMutation.mutate();
    };

    useEffect(() => {
        if (!user) return;

        queueMicrotask(() => {
            setUsername(user.username);
            setBio(user.bio ?? "");
        });
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
                        pendingAvatar={pendingAvatar}
                        pendingBackground={pendingBackground}
                        onAvatarChange={setPendingAvatar}
                        onBackgroundChange={setPendingBackground}
                    />
                    <TextField
                        label={t("settings:profile.username")}
                        maxLength={USERNAME_MAX_LENGTH}
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        disabled={updateProfileMutation.isPending}
                    />
                    <TextAreaField
                        label={t("settings:profile.bio")}
                        maxLength={BIO_MAX_LENGTH}
                        value={bio}
                        onChange={(e) => setBio(e.target.value)}
                        disabled={updateProfileMutation.isPending}
                        className="min-h-[100px]"
                    />

                    <div className="mt-6 flex justify-end">
                        <Button
                            disabled={
                                updateProfileMutation.isPending ||
                                isButtonDisabled
                            }
                            type="submit"
                        >
                            {updateProfileMutation.isPending && <IconLoader />}{" "}
                            {t("common:save_changes")}
                        </Button>
                    </div>
                </form>
            </Section>

            <Section
                title={t("settings:social_media_links.title")}
                description={t("settings:social_media_links.description")}
                className="border-border border-t pt-8"
                contentClassName="p-4"
            >
                <SocialMediaEdit initialLinks={user?.socialMediaLinks} />
            </Section>

            <Section
                title={t("settings:themes.title")}
                description={t("settings:themes.description")}
                className="border-border border-t pt-8"
                contentClassName="p-4"
            >
                <ThemeList />
            </Section>

            <Section
                title={t("settings:language.title")}
                description={t("settings:language.description")}
                className="border-border border-t pt-8"
                contentClassName="p-4"
            >
                <LanguageList />
            </Section>

            <Section
                title={t("settings:disable.title")}
                description={t("settings:disable.description")}
                className="border-border border-t pt-8"
                contentClassName="p-4"
            >
                <div className="flex w-full items-center justify-between">
                    <p className="text-destructive w-2/3 text-sm">
                        {t("settings:disable.note")}
                    </p>
                    <DisableAccountDialog
                        isUpdatingProfile={updateProfileMutation.isPending}
                    />
                </div>
            </Section>
        </>
    );
}
