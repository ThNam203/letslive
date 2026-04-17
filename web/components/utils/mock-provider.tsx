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
                console.error("[MockProvider] Failed to start MSW worker:", err);
                // Still render children so the app is usable (API calls will fail)
                setReady(true);
            }
        }
        startWorker();
    }, []);

    if (!ready) return null;

    return <>{children}</>;
}
