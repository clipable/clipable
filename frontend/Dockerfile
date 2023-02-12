FROM mitchpash/pnpm AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /home/node/app
COPY pnpm-lock.yaml .npmr[c] ./

RUN pnpm fetch

FROM mitchpash/pnpm AS builder
WORKDIR /home/node/app
COPY --from=deps /home/node/app/node_modules ./node_modules
COPY . .

RUN pnpm install -r --offline

RUN pnpm build

FROM mitchpash/pnpm AS runner
WORKDIR /home/node/app

ENV NODE_ENV production

COPY --from=builder /home/node/app/next.config.js ./
COPY --from=builder /home/node/app/public ./public
COPY --from=builder /home/node/app/package.json ./package.json

# Automatically leverage output traces to reduce image size 
# https://nextjs.org/docs/advanced-features/output-file-tracing
# Some things are not allowed (see https://github.com/vercel/next.js/issues/38119#issuecomment-1172099259)
COPY --from=builder --chown=node:node /home/node/app/.next/standalone ./
COPY --from=builder --chown=node:node /home/node/app/.next/static ./.next/static

EXPOSE 3000

ENV PORT 3000

CMD ["node", "server.js"]