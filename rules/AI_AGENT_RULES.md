# AI Agent Rules: Sabuj Engineering Standards (Lightweight Architecture)

You are an AI assistant working on a Next.js 16 project that strictly follows the **Sabuj Engineering Standards — Lightweight Architecture**.
Your primary goal is to maintain the project architecture, coding patterns, and established best practices.

## 1. Technology Stack Overview

- **Framework:** Next.js 16 (App Router)
- **State Management:** Zustand (for global UI & Auth state only)
- **Server State / Data Fetching:** TanStack React Query v5
- **UI Components:** ShadCN UI + Tailwind CSS
- **Forms & Validation:** React Hook Form + Zod
- **Internationalization:** next-intl (Cookie-based, no Next.js routing middleware)
- **Tables:** TanStack Table v8
- **Page Loader:** nextjs-toploader (for perceived performance)

## 2. ⚠️ CRITICAL DIRECTIVES ⚠️

- **DO NOT modify** any generated Shadcn UI components in `src/components/ui/` unless explicitly requested by the user. Use them as-is.
- **NEVER use `useState` + `useEffect`** for data fetching. Always use React Query hooks in `src/hooks/api/`.
- **NEVER use `localStorage`** for authentication tokens. Use secure server-side, HttpOnly cookies (`src/lib/auth.ts`).
- **NEVER hardcode UI strings**. Always use `next-intl` messages and `useTranslations`.
- **NEVER store server data in Zustand**. Zustand is strictly for Global UI State (e.g., sidebar toggles, modals) and lightweight Auth State (with `persist` middleware).

## 3. Architecture & Data Flow

- **Data Flow Pattern:**
  `Page -> Server Component (requireAuth/requirePermission) -> Client component -> DynamicSearchControls + DataTable/Paginator -> React Query Hook -> Axios Client -> API`
- **Route Protection:** Handled via Next.js proxy at `src/proxy.ts`. Do not create separate middleware for route protection.
- **Token Strategy:** Server-first cookie approach. Use `src/lib/auth.ts` for server-side cookie helpers.

## 4. Coding Patterns & Best Practices

- **Form Handling:** Do not write manual `try/catch` blocks in every form. Use `createFormSubmitHandler()` and `handleFormResponse()` from `src/lib/form-helper.ts`.
- **Reusable Components:** Leverage custom components instead of building custom UI per page:
  - `DataTable`
  - `Paginator`
  - `ConfirmDialog`
  - `DynamicSearchControls`
  - `StatusBadge`
  - `PageHeader`
- **React Query Hooks:** Place all API hooks in `src/hooks/api/` (e.g., `useGetUsers`, `useCreateUser`). Use the pre-configured `apiClient` (`src/lib/axios`) for all HTTP requests.
- **RBAC / Permissions:** Check permissions using `<Can permission="x" />` (client-side) and `hasPermission` utility. Do not bypass the RBAC layer.

## 5. Workflow: Adding a New Feature

When asked to create a new feature, follow this exact sequence:

1. **Types:** Define TypeScript interfaces in `src/types/[feature].ts`.
2. **Validation:** Create Zod schemas in `src/lib/validations.ts`.
3. **API Hooks:** Write React Query hooks (`useQuery`, `useMutation`) in `src/hooks/api/use[Feature].ts`.
4. **Form:** Build the feature form in `src/components/forms/[Feature]Form.tsx`.
5. **Page:** Create the UI page in `src/app/(dashboard)/[feature]/page.tsx` (utilize `DataTable` and `DynamicSearchControls`).
6. **Permissions:** Add to `PERMISSIONS` list and protect UI elements with the `<Can />` component.
7. **Nav:** Update `NAV_ITEMS` in the Sidebar layout.
8. **Messages:** Add English translations to `src/messages/en.json`.
9. **Tests:** Write unit tests (Vitest) for validation logic and E2E tests (Playwright) for CRUD operations.

## 6. Anti-Patterns to Avoid

- ❌ `useEffect` + `fetch` for data -> **✅ use React Query hooks**
- ❌ `localStorage` for tokens -> **✅ use HttpOnly cookies**
- ❌ Manual try/catch in every form -> **✅ use `createFormSubmitHandler()`**
- ❌ Hardcoded UI strings -> **✅ use `next-intl` messages**
- ❌ Editing ShadCN-generated files -> **✅ use `shadcn CLI` or use as-is**
- ❌ Building custom filter UIs per page -> **✅ use `DynamicSearchControls`**
- ❌ Putting server data in Zustand -> **✅ let React Query cache it**
