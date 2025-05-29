FROM node:18-alpine AS build

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .

RUN npm run build

FROM nginx:alpine

RUN apk add --no-cache openssl

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

RUN mkdir -p /etc/nginx/ssl && \
    openssl genrsa -out /etc/nginx/ssl/localhost.key 2048 && \
    openssl req -new -x509 -key /etc/nginx/ssl/localhost.key -out /etc/nginx/ssl/localhost.crt -days 365 \
        -subj "/C=RU/ST=Moscow/L=Moscow/O=SleekChat/OU=Development/CN=localhost" && \
    chmod 600 /etc/nginx/ssl/localhost.key && \
    chmod 644 /etc/nginx/ssl/localhost.crt

EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]