FROM node:24-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

ARG VITE_API_URL
ARG VITE_VK_CLIENT_ID
ARG VITE_VK_REDIRECT_URI
ARG VITE_YM_ID
ARG VITE_GA_ID

ENV VITE_API_URL=$VITE_API_URL
ENV VITE_VK_CLIENT_ID=$VITE_VK_CLIENT_ID
ENV VITE_VK_REDIRECT_URI=$VITE_VK_REDIRECT_URI
ENV VITE_YM_ID=$VITE_YM_ID
ENV VITE_GA_ID=$VITE_GA_ID

RUN npm run build

FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
