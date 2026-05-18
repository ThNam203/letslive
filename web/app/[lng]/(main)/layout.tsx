import { Header } from "@/app/[lng]/(main)/_components/header/header";
import { MainBodyLayout } from "@/app/[lng]/(main)/_components/main-body-layout";

type Params = Promise<{ lng: string }>;

export default async function RootLayout({
    children,
    params,
}: Readonly<{
    children: React.ReactNode;
    params: Params;
}>) {
    const { lng } = await params;
    return (
        <div className="flex h-screen w-screen flex-col overflow-hidden">
            <Header lng={lng} />
            <MainBodyLayout>{children}</MainBodyLayout>
        </div>
    );
}
