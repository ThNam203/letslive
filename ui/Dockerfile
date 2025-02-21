FROM node:22-alpine

RUN npm i -g npm

EXPOSE 5000

ENV PORT 5000
ENV NODE_ENV production

WORKDIR /home/nextjs/app

COPY package.json .
COPY package-lock.json .

RUN npm install --omit=optional
RUN npx browserslist@latest --update-db
RUN npx next telemetry disable

COPY . .

RUN npm run build

RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001

USER nextjs

CMD [ "npm", "start" ]
