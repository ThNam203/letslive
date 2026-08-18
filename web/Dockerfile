FROM node:22-alpine AS deps
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

FROM node:22-alpine AS builder
WORKDIR /app

# Accept build arguments for Next.js public environment variables
ARG NEXT_PUBLIC_BACKEND_PROTOCOL
ARG NEXT_PUBLIC_BACKEND_IP_ADDRESS
ARG NEXT_PUBLIC_BACKEND_WS_PROTOCOL
ARG NEXT_PUBLIC_BACKEND_PORT
ARG NEXT_PUBLIC_ENVIRONMENT
ARG NEXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY
ARG CI

# Set as environment variables (needed at build time for Next.js)
ENV NEXT_PUBLIC_BACKEND_PROTOCOL=${NEXT_PUBLIC_BACKEND_PROTOCOL}
ENV NEXT_PUBLIC_BACKEND_IP_ADDRESS=${NEXT_PUBLIC_BACKEND_IP_ADDRESS}
ENV NEXT_PUBLIC_BACKEND_WS_PROTOCOL=${NEXT_PUBLIC_BACKEND_WS_PROTOCOL}
ENV NEXT_PUBLIC_BACKEND_PORT=${NEXT_PUBLIC_BACKEND_PORT}
ENV NEXT_PUBLIC_ENVIRONMENT=${NEXT_PUBLIC_ENVIRONMENT}
ENV NEXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY=${NEXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY}
ENV CI=${CI}

COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN npx browserslist@latest --update-db
RUN npx next telemetry disable

# Use BuildKit secret for SENTRY_AUTH_TOKEN to avoid exposing it in image layers
RUN --mount=type=secret,id=SENTRY_AUTH_TOKEN \
    export SENTRY_AUTH_TOKEN=$(cat /run/secrets/SENTRY_AUTH_TOKEN) && \
    npm run build

FROM node:22-alpine AS runner
WORKDIR /home/nextjs/app

ENV NODE_ENV=production
ENV PORT=5000
ENV HOSTNAME=0.0.0.0
ENV APP_ENV=production

EXPOSE 5000

RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001

# Standalone output already contains only the production deps actually used
COPY --from=builder --chown=nextjs:nodejs /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

CMD ["node", "server.js"]
