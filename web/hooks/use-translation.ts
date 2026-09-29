"use client";

import { useTranslation } from "react-i18next";
import { UseTranslationOptions } from "react-i18next";

function useT(
    ns: string | string[] = "common",
    options?: UseTranslationOptions<undefined>,
) {
    return useTranslation(ns, options);
}

export default useT;
