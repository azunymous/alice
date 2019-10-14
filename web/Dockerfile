FROM node:12
COPY . /
RUN npm run build

FROM httpd:2.4
COPY --from=0 /build  /usr/local/apache2/htdocs/
