FROM golang:alpine as backend_build

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app
COPY backend/ .

RUN go build -o clipable

FROM mitchpash/pnpm AS frontend_deps
RUN apk add --no-cache libc6-compat
WORKDIR /home/node/app
COPY frontend/pnpm-lock.yaml frontend/.npmr[c] ./

RUN ls -lash && exit 1

RUN pnpm fetch && exit 1

FROM mitchpash/pnpm AS frontend_builder
WORKDIR /home/node/app
COPY --from=frontend_deps /home/node/app/node_modules ./node_modules
COPY frontend/ .
RUN ls -lash ./node_modules
RUN pnpm install -r --offline

RUN pnpm build

FROM ghcr.io/jrottenberg/ffmpeg:5-alpine

RUN apk add --update nodejs npm
RUN npm --silent install --global --depth 0 pnpm

# Backend
COPY backend/migrations /migrations
COPY --from=backend_build /app/clipable /

# Frontend
ENV NODE_ENV production

COPY --from=frontend_builder /home/node/app/next.config.js ./
COPY --from=frontend_builder /home/node/app/public ./public
COPY --from=frontend_builder /home/node/app/package.json ./package.json

# Automatically leverage output traces to reduce image size 
# https://nextjs.org/docs/advanced-features/output-file-tracing
# Some things are not allowed (see https://github.com/vercel/next.js/issues/38119#issuecomment-1172099259)
COPY --from=frontend_builder --chown=node:node /home/node/app/.next/standalone ./
COPY --from=frontend_builder --chown=node:node /home/node/app/.next/static ./.next/static

EXPOSE 3000

ENV PORT 3000

# Start 
ENTRYPOINT /clipable & node server.js
