FROM node:22-alpine

RUN npm i -g npm

EXPOSE 5000

ENV PORT=5000
ENV APP_ENV=production

WORKDIR /home/nextjs/app

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

COPY package.json .
COPY package-lock.json .

RUN npm install --omit=optional
RUN npx browserslist@latest --update-db
RUN npx next telemetry disable

COPY . .

# Use BuildKit secret for SENTRY_AUTH_TOKEN to avoid exposing it in image layers
RUN --mount=type=secret,id=SENTRY_AUTH_TOKEN \
    export SENTRY_AUTH_TOKEN=$(cat /run/secrets/SENTRY_AUTH_TOKEN) && \
    npm run build

RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001

USER nextjs

CMD [ "npm", "start" ]
