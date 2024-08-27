"use client"; // Error boundaries must be Client Components
import GlobalErrorComponent from "@/components/errors/GlobalError";

export default function GlobalError({
    error,
    reset,
}: {
    error: Error & { digest?: string };
    reset: () => void;
}) {
    return <GlobalErrorComponent error={error} reset={reset} type="500"/>;
}
