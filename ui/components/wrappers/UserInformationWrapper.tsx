"use client";

import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { GetMeProfile } from "@/lib/api/user";
import { useEffect } from "react";
import { toast } from "react-toastify";

export default function UserInformationWrapper({
    children,
}: {
    children: React.ReactNode;
}) {
    const { setUser, setIsLoading } = useUser();
    const { t } = useT(["fetch-error", "api-response"]);

    useEffect(() => {
        const fetchUser = async () => {
            setIsLoading(true);
            GetMeProfile()
                .then((userRes) => {
                    if (userRes.success && userRes.data) setUser(userRes.data)
                    else toast.error(t(`api-response:${userRes.key}`), { toastId: userRes.requestId })
                })
                .catch((e) => {
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
    }, []);

    return <>{children}</>;
}
