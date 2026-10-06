import "@/app/globals.css";
import QueryProvider from "@/components/query-provider";

export const metadata = {
    title: "letslive admin",
};

export default function RootLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <html lang="en">
            <body>
                <QueryProvider>{children}</QueryProvider>
            </body>
        </html>
    );
}
