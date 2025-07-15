import type { Config } from "tailwindcss";

const config: Config = {
    content: [
        "./pages/**/*.{js,ts,jsx,tsx,mdx}",
        "./components/**/*.{js,ts,jsx,tsx,mdx}",
        "./app/**/*.{js,ts,jsx,tsx,mdx}",
    ],
    theme: {
        extend: {
            colors: {
                background: "hsl(var(--background))",
				"background-hover": "hsl(var(--background-hover))",
                foreground: "hsl(var(--foreground))",
                "foreground-muted": "hsl(var(--foreground-muted))",
                primary: "hsl(var(--primary))",
                "primary-hover": "hsl(var(--primary-hover))",
                "primary-foreground": "hsl(var(--primary-foreground))",
                secondary: "hsl(var(--secondary))",
                "secondary-hover": "hsl(var(--secondary-hover))",
                "secondary-foreground": "hsl(var(--secondary-foreground))",
                muted: "hsl(var(--muted))",
                border: "hsl(var(--border))",
                accent: "hsl(var(--accent))",
                "accent-hover": "hsl(var(--accent-hover))",
                "accent-foreground": "hsl(var(--accent-foreground))",
                destructive: "hsl(var(--destructive))",
                "destructive-hover": "hsl(var(--destructive-hover))",
                "destructive-foreground": "hsl(var(--destructive-foreground))",
            },
        },
    },
    plugins: [require("tailwindcss-animate")],
};

export default config;
