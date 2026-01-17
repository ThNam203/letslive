import React, { Suspense } from "react";
import { Outlet, useParams } from "@tanstack/react-router";
import Loading from "@/routes/[lng]/loading";
import Toast from "@/components/utils/toast";
import { dir } from "i18next";
import UserInformationWrapper from "@/components/wrappers/UserInformationWrapper";

export default function LayoutComponent() {
    const { lng } = useParams({ from: "/$lng" });

    return (
        <div lang={lng} dir={dir(lng)}>
            <Suspense fallback={<Loading />}>
                <UserInformationWrapper>
                    <Outlet />
                </UserInformationWrapper>
                <Toast />
            </Suspense>
        </div>
    );
}
