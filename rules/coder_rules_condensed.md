# Coder Rules (Condensed)

Generation-time constraints for the AI Coder. Keep this file small — it is
injected into every coder call. The full human-facing guide lives in
`nextjs-project-setup-guide-light.md` and is intentionally NOT sent to the model.

## Stack
- Next.js 16 (App Router) + TypeScript + Tailwind CSS.
- UI: shadcn/ui (import, never reconstruct, never edit files in `src/components/ui/`).
- Server state: TanStack React Query v5 (hooks in `src/hooks/api/`). Never fetch with `useEffect` + `fetch`.
- Global UI/auth state: Zustand only. Never store server data in Zustand.
- Forms: React Hook Form + Zod.
- i18n: next-intl. Never hardcode user-visible strings.
- Icons: lucide-react.

## Folder Structure (output target)
- Pages: `src/app/<route>/page.tsx`
- Custom components: `src/components/common/<Name>.tsx`
- shadcn components: `src/components/ui/<Name>.tsx` (installed via CLI, not generated)
- Layout/forms: `src/components/layout/`, `src/components/forms/`
- Hooks: `src/hooks/api/`, types: `src/types/`, messages: `src/messages/en.json`

## Hard Rules
- No unused imports/variables/functions. No `any` types — define interfaces.
- Navigation: `import Link from 'next/link'` and `import { useRouter } from 'next/navigation'`. Do NOT use `@/i18n/navigation`.
- All visible text goes through `next-intl` (`useTranslations`) and into the translations output.
- Use real HTML + Tailwind for layout; never render a whole page/section as one `<img>`.
- Never emit raw inline `<svg>`. Use downloaded assets or a `https://placehold.co` placeholder.
- Map Figma layout/typography/color data to Tailwind utilities; colors are provided as hex.
