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
                foreground: "hsl(var(--foreground))",
				"background-hover": "hsl(var(--background-hover))",
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
            },
        },
    },
    darkMode: ["class", "class"],
    plugins: [require("tailwindcss-animate")],
};

export default config;
