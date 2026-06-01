# Vercel Setup — Merchant Portal & Admin Dashboard

Frontends deploy to **Vercel via native Git integration** — no GitHub Actions job,
no `VERCEL_TOKEN`. Vercel watches the repo and builds automatically:

- Push to **`main`** → Production deploy
- Any **pull request** → Preview deploy

You create **two separate Vercel projects** from the **same** GitHub repo
(`theetaz/open-pay`), each pointed at a different subdirectory.

---

## 1. Create the two projects

For **each** app, in the Vercel dashboard → **Add New… → Project** → import
`theetaz/open-pay`, then set:

| Setting | Merchant Portal | Admin Dashboard |
| --- | --- | --- |
| **Project Name** | `openpay-merchant` | `openpay-admin` |
| **Framework Preset** | Vite | Vite |
| **Root Directory** | `apps/merchant-portal` | `apps/admin-dashboard` |
| **Build Command** | `pnpm build` (default) | `pnpm build` (default) |
| **Output Directory** | `dist` (default) | `dist` (default) |
| **Install Command** | `pnpm install` (default) | `pnpm install` (default) |

> **Root Directory is the critical field.** Setting it makes Vercel treat that
> subfolder as the app root, so the existing `apps/*/vercel.json` (SPA rewrites)
> and `package.json` are picked up correctly. There is no pnpm workspace, so each
> app installs independently — no extra monorepo config needed.

## 2. Production branch

Project → **Settings → Git → Production Branch** → set to **`main`** for both.
(Pushes to `main` deploy production; PRs get preview URLs automatically.)

## 3. Environment variable

Both apps read the API base URL from `VITE_API_URL` at build time.
Project → **Settings → Environment Variables**, add for **Production** (and
Preview if you want previews to hit prod):

```
VITE_API_URL = https://olp-api.nipuntheekshana.com
```

Redeploy after adding it (env vars are baked in at build time for Vite).

## 4. Custom domains

Project → **Settings → Domains** → Add:

| Project | Domain |
| --- | --- |
| `openpay-merchant` | `olp-merchant.nipuntheekshana.com` |
| `openpay-admin` | `olp-admin.nipuntheekshana.com` |

Vercel will show a **CNAME** target (usually `cname.vercel-dns.com`). Add it in
**Cloudflare**:

| Type | Name | Target | Proxy |
| --- | --- | --- | --- |
| CNAME | `olp-merchant` | `cname.vercel-dns.com` (use the value Vercel shows) | **DNS only (grey cloud)** |
| CNAME | `olp-admin` | `cname.vercel-dns.com` (use the value Vercel shows) | **DNS only (grey cloud)** |

> For Vercel-hosted domains, set the Cloudflare record to **DNS only (grey
> cloud)** — Vercel manages its own TLS. (This differs from the API record, which
> is **Proxied/orange** because TLS for the VPS origin is terminated at
> Cloudflare.)

## 5. Verify

After the first deploy + DNS propagation:

```bash
curl -I https://olp-merchant.nipuntheekshana.com   # 200, served by Vercel
curl -I https://olp-admin.nipuntheekshana.com      # 200, served by Vercel
```

Open each in a browser, log in, and confirm API calls reach
`https://olp-api.nipuntheekshana.com` (check the Network tab).

---

## Summary of all DNS records (Cloudflare)

| Type | Name | Value | Proxy | TLS |
| --- | --- | --- | --- | --- |
| A | `olp-api` | `178.104.111.17` | **Proxied (orange)** | Cloudflare Full |
| CNAME | `olp-merchant` | `cname.vercel-dns.com` | **DNS only (grey)** | Vercel-managed |
| CNAME | `olp-admin` | `cname.vercel-dns.com` | **DNS only (grey)** | Vercel-managed |
