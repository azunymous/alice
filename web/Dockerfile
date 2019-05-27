# DOCKERFILE expects to be built from the project root
FROM node:8

WORKDIR /usr/src/app

COPY ./web/package*.json ./

RUN npm ci --only=production
# Bundle app source
COPY ./web/server.js ./
COPY ./web/public/ ./public/

EXPOSE 3000
CMD [ "npm", "start" ]