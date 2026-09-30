# Stage 1: Build the React app
FROM node:20-alpine AS build
WORKDIR /app

COPY frontend/package.json ./
RUN npm install

COPY frontend/. ./

# Vite inlines VITE_* env vars into the bundle at build time — this must
# be an address the BROWSER can reach (docker-compose passes
# http://localhost:8080/graphql), not a compose-internal hostname.
ARG VITE_GRAPHQL_URL=http://localhost:8080/graphql
ENV VITE_GRAPHQL_URL=${VITE_GRAPHQL_URL}

RUN npm run build

# Stage 2: Serve with nginx
FROM nginx:stable-alpine AS production
RUN rm /etc/nginx/conf.d/default.conf
COPY frontend/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
HEALTHCHECK --interval=15s --timeout=5s --retries=5 --start-period=10s \
  CMD wget -qO- http://localhost/healthz || exit 1

CMD ["nginx", "-g", "daemon off;"]
