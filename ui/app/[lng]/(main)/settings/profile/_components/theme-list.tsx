"use client";

import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { cn } from "@/utils/cn";
import { THEME_COLORS } from "@/constant/theme";
import useT from "@/hooks/use-translation";

const ThemeList = ({ className }: { className?: string }) => {
    const [mounted, setMounted] = useState(false);
    const { theme, setTheme } = useTheme();
    const { t } = useT(["theme"]);

    useEffect(() => {
        setMounted(true);
    }, []);

    if (!mounted) return null;

    return (
        <div className={cn("flex gap-2", className)}>
            {Object.values(THEME_COLORS).map((color) => (
                <button
                    key={color}
                    onClick={() => setTheme(color)}
                    className={cn(
                        "rounded-md border px-4 py-2 text-sm capitalize transition-all",
                        "border-border",
                        theme === color ? "font-bold ring-2 ring-border" : "",
                        color === "light" ? "bg-white text-[#0f1729]" : "",
                        color === "dark" ? "bg-[#0f1729] text-white" : "",
                    )}
                >
                    {t(`theme:${color}`)}
                </button>
            ))}
        </div>
    );
};

export default ThemeList;
