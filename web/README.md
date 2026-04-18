This is a [Next.js](https://nextjs.org/) project bootstrapped with [`create-next-app`](https://github.com/vercel/next.js/tree/canary/packages/create-next-app).

## Getting started

Requires Node.js **20.9+** (see the Next.js CLI check).

### Run against a real backend

1. Copy `.env.example` to `.env` and fill in values for your environment.
2. Start the dev server:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) (or the host/port from your `.env` `PORT`).

### Mock API mode (easiest web-only testing)

For UI and frontend work you can run the app **without** the Go API or Docker stack:

```bash
npm run dev:mock
```

This loads [`.env.mock`](./.env.mock), which sets `NEXT_PUBLIC_USE_MOCK_API=true` so browser requests are handled by [MSW](https://mswjs.io/) instead of a live backend. Handlers live under [`mocks/handlers/`](./mocks/handlers/).


## Learn more

- [Next.js documentation](https://nextjs.org/docs)
- [Learn Next.js](https://nextjs.org/learn)
