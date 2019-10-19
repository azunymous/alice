FROM node:12 as build-stage
COPY . /
RUN npm install && npm run build
COPY ./nginx/nginx.conf /nginx.conf


FROM nginx:1.17.4-alpine
COPY --from=build-stage /nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build-stage /build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]