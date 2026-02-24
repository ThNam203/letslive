import React from "react";

export default function DefaultBackgound() {
    return (
        <div className="border-border bg-muted absolute inset-0 grid h-[300px] grid-cols-6 gap-2 rounded-lg border p-2">
            {[...Array(18)].map((_, i) => (
                <svg
                    key={i}
                    className="text-foreground h-8 w-8 opacity-25"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                >
                    <path d="M21 3H3v18h18V3zm-9 14H7v-4h5v4zm0-6H7V7h5v4zm6 6h-4v-4h4v4zm0-6h-4V7h4v4z" />
                </svg>
            ))}
        </div>
    );
}
