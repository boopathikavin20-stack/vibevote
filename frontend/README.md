# VibeVote frontend

## Deployment

When the frontend and backend use different public URLs, set `VITE_API_URL` to
the backend URL before running `npm run build`, for example:

```bash
VITE_API_URL=https://api.example.com npm run build
```

The backend must allow the deployed frontend origin through `FRONTEND_URL`.
When the Go server serves `frontend/dist`, leave `VITE_API_URL` empty so the
browser uses the same-origin `/api` routes.

This template provides a minimal setup to get React working in Vite with HMR and some Oxlint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Oxc](https://oxc.rs)
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/)

## React Compiler

The React Compiler is not enabled on this template because of its impact on dev & build performances. To add it, see [this documentation](https://react.dev/learn/react-compiler/installation).

## Expanding the Oxlint configuration

If you are developing a production application, we recommend using TypeScript with type-aware lint rules enabled. Check out the [TS template](https://github.com/vitejs/vite/tree/main/packages/create-vite/template-react-ts) for information on how to integrate TypeScript and Oxlint's TypeScript related rules in your project.
