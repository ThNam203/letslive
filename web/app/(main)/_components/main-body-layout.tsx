import { cookies } from "next/headers";
import {
    LEFT_BAR_COOKIE_NAME,
    parseLeftBarLayout,
} from "@/components/leftbar/left-bar-layout";
import { MainBodyPanels } from "./main-body-panels";

export async function MainBodyLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const cookieStore = await cookies();
    const leftBarLayout = parseLeftBarLayout(
        cookieStore.get(LEFT_BAR_COOKIE_NAME)?.value,
    );

    return (
        <MainBodyPanels initialLeftBarLayout={leftBarLayout}>
            {children}
        </MainBodyPanels>
    );
}
