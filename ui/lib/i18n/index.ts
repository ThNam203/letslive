import i18next from "./i18next"
import { headers } from "next/headers"
import { headerName } from "./settings"

export async function myGetT(ns: string | string[] = "translation") {
  const headerList = headers()
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