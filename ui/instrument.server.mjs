import * as Sentry from "@sentry/tanstackstart-react";

Sentry.init({
  dsn: "https://990ba0f96a5ad34873933929847aaf41@o4510607010299904.ingest.de.sentry.io/4510607015084112",
  
  // Adds request headers and IP for users, for more info visit:
  // https://docs.sentry.io/platforms/javascript/guides/tanstackstart-react/configuration/options/#sendDefaultPii
  sendDefaultPii: true,
});