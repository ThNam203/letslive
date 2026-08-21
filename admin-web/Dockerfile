FROM node:22-alpine AS deps
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

FROM node:22-alpine AS builder
WORKDIR /app

ARG NEXT_PUBLIC_ADMIN_API_URL
ENV NEXT_PUBLIC_ADMIN_API_URL=${NEXT_PUBLIC_ADMIN_API_URL}

COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN npx next telemetry disable
RUN npm run build

FROM node:22-alpine AS runner
WORKDIR /home/nextjs/app

ENV NODE_ENV=production
ENV PORT=5001
ENV HOSTNAME=0.0.0.0

EXPOSE 5001

RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001

# Standalone output already contains only the production deps actually used
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

CMD ["node", "server.js"]
