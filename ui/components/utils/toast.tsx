"use client";

import { Bounce, ToastContainer, toast as toastify } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import i18next from "@/lib/i18n/i18next";

const DEFAULT_ERROR_KEY = "api-response:default_error";

function resolveMessage(message: string | undefined): string {
    if (message != null && String(message).trim() !== "") {
        return message;
    }
    return i18next.t(DEFAULT_ERROR_KEY);
}

function toastWithDefault(
    message: string | undefined,
    options?: Parameters<typeof toastify>[1],
) {
    return toastify(resolveMessage(message), {
        type: "error",
        ...options,
    });
}

toastWithDefault.error = (message?: string, options?: Parameters<typeof toastify>[1]) =>
    toastify.error(resolveMessage(message), options);
toastWithDefault.success = (message?: string, options?: Parameters<typeof toastify>[1]) =>
    toastify.success(message ?? resolveMessage(message), options);
toastWithDefault.info = (message?: string, options?: Parameters<typeof toastify>[1]) =>
    toastify.info(message ?? resolveMessage(message), options);
toastWithDefault.warning = (message?: string, options?: Parameters<typeof toastify>[1]) =>
    toastify.warning(message ?? resolveMessage(message), options);
toastWithDefault.warn = toastWithDefault.warning;
toastWithDefault.dismiss = toastify.dismiss;
toastWithDefault.promise = toastify.promise;
toastWithDefault.isActive = toastify.isActive;
toastWithDefault.update = toastify.update;
toastWithDefault.onChange = toastify.onChange;

export const toast = toastWithDefault;

const Toast = () => {
    return (
        <ToastContainer
            position="bottom-right"
            autoClose={4000}
            hideProgressBar={false}
            newestOnTop={true}
            closeOnClick
            rtl={false}
            pauseOnFocusLoss
            draggable
            pauseOnHover
            theme="colored"
            transition={Bounce}
        />
    );
};

export default Toast;
