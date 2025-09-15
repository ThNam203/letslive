import { headers } from "next/headers"
import { headerName } from "./settings"
import i18next from "@/lib/i18n/i18next"

export async function myGetT(ns: string | string[] = "translation") {
  const headerList = await headers()
  const lng = headerList.get(headerName)

  if (lng && i18next.resolvedLanguage !== lng) {
    await i18next.changeLanguage(lng)
  }
  if (ns && !i18next.hasLoadedNamespace(ns)) {
    await i18next.loadNamespaces(ns)
  }

  const language = lng ?? i18next.resolvedLanguage ?? null

  return {
    t: i18next.getFixedT(
      language,
      Array.isArray(ns) ? ns[0] : ns
    ),
    i18n: i18next,
  }
}