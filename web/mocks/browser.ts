import { setupWorker } from "msw/browser";
import { authHandlers } from "./handlers/auth";
import { userHandlers } from "./handlers/user";
import { livestreamHandlers } from "./handlers/livestream";
import { vodHandlers } from "./handlers/vod";
import { chatHandlers } from "./handlers/chat";
import { dmHandlers } from "./handlers/dm";
import { financeHandlers } from "./handlers/finance";
import { shopHandlers } from "./handlers/shop";

export const worker = setupWorker(
    ...authHandlers,
    ...userHandlers,
    ...livestreamHandlers,
    ...vodHandlers,
    ...chatHandlers,
    ...dmHandlers,
    ...financeHandlers,
    ...shopHandlers,
);
