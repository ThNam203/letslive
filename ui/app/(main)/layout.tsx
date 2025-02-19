"use client"
import { Header } from "@/components/header/header";
import { LeftBar } from "@/components/LeftBar";
import { usePathname } from "next/navigation";
import { useEffect } from "react";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
    const path = usePathname();
    useEffect(() => {
        
    }, [path])

  return (
    <div className="h-screen w-screen">
      <Header />
      <div className="w-full h-[calc(100%-48px)] flex flex-row">
        <LeftBar />
        <div className="h-full w-full xl:ml-72 max-xl:ml-12">{children}</div>
      </div>
    </div>
  );
}