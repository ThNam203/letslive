"use client";

import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { GetMeProfile } from "@/lib/api/user";
import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { toast } from "@/components/utils/toast";

export default function UserInformationWrapper({
    children,
}: {
    children: React.ReactNode;
}) {
    const { setUser, setIsLoading } = useUser();
    const { t } = useT(["fetch-error", "api-response"]);
    const router = useRouter();
    const pathname = usePathname();

    useEffect(() => {
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
    }, []);

    // Render children immediately - user fetch happens in background
    return <>{children}</>;
}
