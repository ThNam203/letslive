"use client";

import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { GetMeProfile } from "@/lib/api/user";
import { useEffect, useRef } from "react";
import { usePathname, useRouter } from "next/navigation";
import { toast } from "@/components/utils/toast";
import { I18N_LANGUAGES } from "@/lib/i18n/settings";
import { switchLocale } from "@/lib/i18n/switch-locale";
import ReactivateAccountDialog from "@/components/wrappers/ReactivateAccountDialog";

export default function UserInformationWrapper({
    children,
}: {
    children: React.ReactNode;
}) {
    const { user, setUser, setIsLoading } = useUser();
    const { t } = useT(["fetch-error", "api-response"]);
    const router = useRouter();
    const pathname = usePathname();
    const hasFetchedRef = useRef(false);

    // hydrate FE locale from the user's saved preference (login or session-restore)
    useEffect(() => {
        if (!user?.locale || !I18N_LANGUAGES.includes(user.locale)) return;
        const currentLocale = pathname.split("/")[1];
        if (currentLocale === user.locale) return;
        switchLocale(router, pathname, user.locale, { syncToBackend: false });
    }, [user, pathname, router]);

    useEffect(() => {
        if (hasFetchedRef.current) return;
        hasFetchedRef.current = true;

        const fetchUser = async () => {
            setIsLoading(true);
            GetMeProfile()
                .then((userRes) => {
                    if (userRes.success && userRes.data) {
                        setUser(userRes.data);
                        if (
                            userRes.data.username === "" &&
                            !pathname.includes("account-setup")
                        ) {
                            router.push("/account-setup");
                        }
                    } else if (!userRes.success && userRes.statusCode != 401)
                        toast.error(t(`api-response:${userRes.key}`), {
                            toastId: userRes.requestId,
                        });
                })
                .catch((_) => {
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                })
                .finally(() => {
                    setIsLoading(false);
                });
        };

        fetchUser();
    }, [setUser, setIsLoading, router, pathname, t]);

    // Render children immediately - user fetch happens in background
    return (
        <>
            {children}
            <ReactivateAccountDialog />
        </>
    );
}
