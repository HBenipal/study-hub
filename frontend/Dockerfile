# CITATION
# THIS FILE IS AI ASSISTED
# PROMPT: GENERATE A DOCKERFILE FOR THE FRONTEND/ FOLDER IT SHOULD SUPPORT NEXTJS DYNAMIC ROUTES
# ENV VARIABLES AND NGINX

# Build stage
FROM node:20-alpine AS builder

# Accept build arguments
ARG NEXT_PUBLIC_ENV
ARG NEXT_PUBLIC_GITHUB_CLIENT_ID

# Set environment variables from build args
ENV NEXT_PUBLIC_ENV=$NEXT_PUBLIC_ENV
ENV NEXT_PUBLIC_GITHUB_CLIENT_ID=$NEXT_PUBLIC_GITHUB_CLIENT_ID

RUN mkdir -p /home/app
WORKDIR /home/app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source code
COPY . .

# Build Next.js
RUN npm run build

# Production stage with nginx and node
FROM node:20-alpine

# Install nginx
RUN apk add --no-cache nginx

RUN mkdir -p /home/app
WORKDIR /home/app

# Copy package files and install production dependencies only
COPY package*.json ./
RUN npm ci --omit=dev

# Copy built Next.js files from builder
COPY --from=builder /home/app/.next ./.next
COPY next.config.js ./

# Copy public directory if it exists (optional)
RUN mkdir -p public

# Copy nginx configuration
COPY nginx.conf /etc/nginx/http.d/default.conf

# Create startup script
RUN echo '#!/bin/sh' > /start.sh && \
    echo 'npm start &' >> /start.sh && \
    echo 'nginx -g "daemon off;"' >> /start.sh && \
    chmod +x /start.sh

ENV NODE_ENV=production
ENV PORT=3001

EXPOSE 80

CMD ["/start.sh"]