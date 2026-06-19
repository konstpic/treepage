# Frontend Reference Analysis — bot-pay/frontend

This report documents the design system and architecture of `/Users/k.pichugin/Documents/bot-pay/frontend`, used as the visual and structural foundation for the TreePage documentation platform.

## Project Structure

```
frontend/
├── app/                    # Next.js App Router pages
│   ├── layout.tsx          # Root layout (nav, footer, providers)
│   ├── globals.css         # Tailwind + design tokens
│   ├── page.tsx            # Landing
│   ├── auth/               # Authentication flows
│   ├── dashboard/          # Authenticated user dashboard
│   ├── settings/           # User settings
│   └── ...                 # pricing, terms, privacy, etc.
├── components/             # Reusable UI components
│   ├── navigation.tsx
│   ├── footer.tsx
│   ├── providers.tsx
│   ├── motion-wrapper.tsx
│   └── ...
├── lib/                    # Utilities, API client, stores
│   ├── api.ts
│   ├── store.ts
│   └── utils.ts
├── middleware.ts           # Edge middleware (feature gates)
├── tailwind.config.ts
└── package.json
```

**Stack:** Next.js 15, React 19, TypeScript, Tailwind CSS 3, Framer Motion, TanStack Query, Zustand, Lucide icons.

**Note:** The spec mentions MUI, but the reference project uses **Tailwind CSS with custom component classes** — not Material UI. TreePage adopts the reference design (Tailwind + glass morphism) rather than introducing MUI.

---

## Routing Architecture

- **Framework:** Next.js App Router (`app/` directory).
- **File-based routing:** Each `page.tsx` maps to a URL segment.
- **Layout nesting:** Root layout wraps all pages with `Navigation`, `main`, `Footer`, and `Providers`.
- **Client vs server:** Pages with interactivity use `"use client"` directive.
- **Middleware:** `middleware.ts` handles optional Basic Auth site gate via env feature flags.
- **Auth-gated routes:** Client-side checks in pages (redirect to `/auth` if not authenticated) rather than middleware-based JWT validation.

**TreePage adaptation:** React Router v6 with Vite (as specified), preserving the same route semantics (`/`, `/auth`, `/dashboard`, `/spaces/:slug`, `/docs/:path`).

---

## State Management Approach

### Zustand stores (`lib/store.ts`)

| Store | Purpose |
|-------|---------|
| `useAuthStore` | JWT tokens in localStorage, user profile, hydrate on mount |
| `useBrandingStore` | Public branding config from API |

**Auth flow:**
1. `hydrate()` reads `access_token` from localStorage on mount.
2. If authenticated, `Providers` fetches `/api/auth/me`.
3. `setAuth(access, refresh)` persists tokens.
4. `logout()` revokes refresh token server-side, clears localStorage.

### TanStack Query

Used implicitly via direct `api()` calls in `useEffect`; not wrapped in query hooks everywhere. Dashboard pages fetch data on mount with manual loading/error state.

**TreePage adaptation:** TanStack Query for spaces, documents, search; Zustand for auth (same pattern as reference).

---

## Component Organization

| Layer | Location | Examples |
|-------|----------|----------|
| Layout | `components/` | `navigation.tsx`, `footer.tsx`, `providers.tsx` |
| Motion | `components/motion-wrapper.tsx` | `FadeIn`, `StaggerContainer`, `StaggerItem` |
| Forms | `components/` | `password-field.tsx` |
| Feedback | `components/` | `toast-notifications.tsx` |
| Pages | `app/**/page.tsx` | Self-contained with local state |

**Patterns:**
- Icons from `lucide-react`.
- Conditional CSS via `cn()` utility (clsx + tailwind-merge).
- Framer Motion for page transitions and mobile menu animations.
- Glass-card UI containers (`.glass`, `.glass-hover`).

---

## UI Framework Usage

**Not MUI.** The design system is custom Tailwind:

| Token | Value |
|-------|-------|
| Background | `#07070f` with radial purple/cyan gradients |
| Primary brand | Violet scale (`brand-400`–`brand-700`, `#8b5cf6`) |
| Text | `slate-100` body, `slate-400` muted |
| Accent gradients | `from-brand-400 via-purple-400 to-cyan-400` |
| Borders | `white/[0.06]`–`white/[0.10]` |
| Font | Inter (Google Fonts, latin + cyrillic) |

**Component classes** (defined in `globals.css` `@layer components`):
- `.glass`, `.glass-hover` — frosted glass cards
- `.btn-primary`, `.btn-secondary`, `.btn-ghost`
- `.input-field`
- `.gradient-text`
- `.badge-active`, `.badge-expired`, `.badge-pending`
- `.skeleton`

---

## Theme Implementation

1. **Tailwind config** (`tailwind.config.ts`): extends `colors.brand` (50–950), custom animations (`float`, `fade-in`, `slide-up`, `shimmer`, `pulse-glow`).
2. **Global CSS** (`globals.css`): body background, scrollbar styling, component tokens.
3. **No dark/light toggle** — single dark theme only.
4. **Dynamic branding:** project name from API, applied to nav logo and `document.title`.

---

## Styling Methodology

- **Utility-first Tailwind** with semantic component classes for repeated patterns.
- **`cn()` helper** for conditional class merging.
- **Responsive:** `sm:` breakpoint for desktop nav vs mobile hamburger menu.
- **Spacing:** `max-w-6xl mx-auto px-4 sm:px-6` content container.
- **Sticky header:** `sticky top-0 z-50 backdrop-blur-xl`.
- **Animations:** Framer Motion for enter/exit; CSS keyframes for ambient effects.

---

## API Integration Patterns

### Client (`lib/api.ts`)

- Base URL from `NEXT_PUBLIC_API_URL`.
- Bearer token from localStorage on every request.
- **401 handling:** mutex-protected refresh token flow (`refreshTokens()`).
- **Error class:** `ApiError` with status and `retryAfter` for rate limits.
- **Detail parsing:** `formatApiDetail()` handles FastAPI-style error bodies.
- **Helpers:** `fetchWithAuth`, `downloadAuthenticatedFile`, `fetchAuthenticatedBlob`.

### Provider bootstrap (`components/providers.tsx`)

1. Hydrate auth from localStorage.
2. Fetch public branding (no auth).
3. Fetch user profile when authenticated.
4. SSE notifications when logged in.

### Auth endpoints (reference)

| Endpoint | Purpose |
|----------|---------|
| `POST /api/auth/refresh` | Token refresh |
| `POST /api/auth/logout` | Revoke refresh token |
| `GET /api/auth/me` | Current user |
| `GET /api/public/branding` | Public config |

**TreePage adaptation:** Same JWT + refresh pattern; OIDC login redirects to `backend-auth`; API calls to `backend-server`.

---

## Design Decisions for TreePage

| Aspect | Reference | TreePage |
|--------|-----------|---------|
| Framework | Next.js | Vite + React Router |
| Styling | Tailwind + glass | Same tokens copied |
| State | Zustand + TanStack Query | Same |
| Icons | Lucide | Lucide |
| Motion | Framer Motion | Framer Motion |
| Auth | Email/password + OAuth | OIDC via backend-auth |
| Layout | Nav + main + footer | Same structure + doc sidebar |

The documentation platform must look like a natural extension of bot-pay: dark violet glass UI, gradient accents, same button/input styles, and navigation patterns.
