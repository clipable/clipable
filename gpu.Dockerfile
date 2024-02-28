FROM golang as backend-build

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

FROM ghcr.io/aperim/nvidia-cuda-ffmpeg:12.2.2-6.1.1-ubuntu22.04-0.3.6
WORKDIR /clipable
RUN apt update && apt install -y --no-install-recommends curl nginx && \
        curl -fsSL https://deb.nodesource.com/setup_21.x | bash - && \
        apt install -y nodejs supervisor && \
        npm install pnpm

ENV NODE_ENV production

COPY --from=frontend-builder /home/node/app/next.config.js ./
COPY --from=frontend-builder /home/node/app/public ./public
COPY --from=frontend-builder /home/node/app/package.json ./package.json

# Automatically leverage output traces to reduce image size 
# https://nextjs.org/docs/advanced-features/output-file-tracing
# Some things are not allowed (see https://github.com/vercel/next.js/issues/38119#issuecomment-1172099259)
COPY --from=frontend-builder /home/node/app/.next/standalone ./
COPY --from=frontend-builder /home/node/app/.next/static ./.next/static

COPY backend/migrations ./migrations
COPY ./nginx.conf /etc/nginx/nginx.conf
COPY gpu.entrypoint.sh .
COPY --from=backend-build /app/clipable .
COPY ./supervisord.conf /supervisord.conf

ENTRYPOINT ./gpu.entrypoint.sh
