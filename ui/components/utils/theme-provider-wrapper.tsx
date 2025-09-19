"use client";

import { ThemeProvider } from "next-themes";
import { useEffect, useState } from "react";

export function ThemeProviderWrapper({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  if (!mounted) {
    // Render children without theming until hydrated
    return <>{children}</>;
  }

  return (
    <ThemeProvider attribute="data-theme" defaultTheme="system" enableSystem>
      {children}
    </ThemeProvider>
  );
}
