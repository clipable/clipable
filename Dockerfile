FROM golang:alpine as backend-build

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app
COPY backend/ .
RUN go build -o clipable

FROM gplane/pnpm AS frontend-builder
WORKDIR /home/node/app
COPY frontend/pnpm-lock.yaml frontend/.npmr[c] ./
RUN pnpm fetch

COPY frontend/ .
RUN pnpm install -r --offline
RUN pnpm build


FROM node:alpine

RUN apk add --update --no-cache nginx

ENV NODE_ENV production

COPY --from=frontend-builder /home/node/app/next.config.js ./
COPY --from=frontend-builder /home/node/app/public ./public
COPY --from=frontend-builder /home/node/app/package.json ./package.json

# Automatically leverage output traces to reduce image size 
# https://nextjs.org/docs/advanced-features/output-file-tracing
# Some things are not allowed (see https://github.com/vercel/next.js/issues/38119#issuecomment-1172099259)
COPY --from=frontend-builder /home/node/app/.next/standalone ./
COPY --from=frontend-builder /home/node/app/.next/static ./.next/static

COPY --from=mwader/static-ffmpeg:6.0 /ffmpeg /usr/local/bin/
COPY --from=mwader/static-ffmpeg:6.0 /ffprobe /usr/local/bin/

COPY backend/migrations /migrations
COPY ./nginx.conf /etc/nginx/nginx.conf
COPY --from=backend-build /app/clipable /

ENTRYPOINT /clipable & node server.js & nginx -g "daemon off;"
