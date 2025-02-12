FROM node:18-alpine

WORKDIR /home/node/app

COPY package*.json ./

RUN npm i

COPY . .

RUN npm run build