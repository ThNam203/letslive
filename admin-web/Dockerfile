FROM node:22-alpine

EXPOSE 5001

ENV PORT=5001

WORKDIR /home/nextjs/app

ARG NEXT_PUBLIC_ADMIN_API_URL

ENV NEXT_PUBLIC_ADMIN_API_URL=${NEXT_PUBLIC_ADMIN_API_URL}

COPY package.json .
COPY package-lock.json .

RUN npm install
RUN npx next telemetry disable

COPY . .

RUN npm run build

RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001

USER nextjs

CMD [ "npm", "start" ]
