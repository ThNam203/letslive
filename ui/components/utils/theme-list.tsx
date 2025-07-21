"use client";

import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { cn } from "@/utils/cn";
import { THEME_COLORS } from "@/constant/theme";

const ThemeList = ({ className }: { className?: string }) => {
  const [mounted, setMounted] = useState(false);
  const { theme, setTheme } = useTheme();

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
            "px-4 py-2 rounded-md border text-sm capitalize transition-all",
            "border-border",
            theme === color ? "ring-2 ring-border font-bold" : "",
            color === "light" ? "bg-white text-[#0f1729]" : "",
            color === "dark" ? "bg-[#0f1729] text-white" : "",
          )}
        >
          {color}
        </button>
      ))}
    </div>
  );
};

export default ThemeList;
