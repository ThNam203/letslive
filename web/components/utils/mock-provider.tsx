"use client";

import GLOBAL from "@/global";
import { useEffect, useState } from "react";

/**
 * Initializes MSW (Mock Service Worker) in the browser when
 * NEXT_PUBLIC_USE_MOCK_API=true is set.
 *
 * Renders children only after the worker is ready so the first
 * fetch calls are already intercepted.
 *
 * NOTE: The middleware must exclude /mockServiceWorker.js from
 * the locale-redirect logic (see middleware.ts matcher).
 */
export default function MockProvider({
    children,
}: {
    children: React.ReactNode;
}) {
    const [ready, setReady] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    useEffect(() => {
        async function startWorker() {
            try {
                const { worker } = await import("@/mocks/browser");
                await worker.start({
                    onUnhandledRequest(request, print) {
                        if (request.url.startsWith(GLOBAL.API_URL)) {
                            print.warning();
                        }
                    },
                    serviceWorker: {
                        url: "/mockServiceWorker.js",
                    },
                });
                setReady(true);
            } catch (err) {
                console.error(
                    "[MockProvider] Failed to start MSW worker:",
                    err,
                );
                // Fail loud: without the worker, requests would silently
                // hit the real backend while the dev believes mocks are on
                setError(
                    err instanceof Error ? err : new Error(String(err)),
                );
            }
        }
        startWorker();
    }, []);

    if (error) {
        return (
            <div className="p-6 font-mono">
                <h2 className="text-red-600 font-bold">
                    MSW worker failed to start — refusing to render with
                    mocks disabled
                </h2>
                <p>
                    NEXT_PUBLIC_USE_MOCK_API=true but the mock service
                    worker could not be installed, so requests would hit
                    the real backend. Fix the worker or unset the flag.
                </p>
                <pre className="whitespace-pre-wrap">{error.message}</pre>
            </div>
        );
    }

    if (!ready) return null;

    return <>{children}</>;
}
