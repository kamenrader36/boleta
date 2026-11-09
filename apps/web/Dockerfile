# Etapa 1: Construcción (Build)
FROM node:20-alpine AS build-stage
WORKDIR /app
COPY apps/web/package*.json ./
RUN npm install
COPY apps/web/ ./
RUN npm run build

# Etapa 2: Producción (Nginx)
FROM nginxinc/nginx-unprivileged:alpine AS production-stage
USER 0:0
RUN apk update && apk upgrade --no-cache
COPY --from=build-stage /app/dist /usr/share/nginx/html
COPY apps/web/nginx.conf /etc/nginx/conf.d/default.conf
USER 101:101
EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
