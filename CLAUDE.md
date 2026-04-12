# shushingface

Voice-to-text refinement desktop app. Go backend (Wails v2) + React 19 / TypeScript frontend.

## Stack

- **Backend**: Go 1.26, Wails v2, SQLite (modernc), malgo audio
- **Frontend**: React 19, TypeScript, Tailwind CSS v4, Radix UI (via `radix-ui` package), sonner toasts, lucide-react icons
- **Font**: JetBrains Mono globally — do NOT add `font-mono` classes, it's the default
- **Build**: `go build ./...` for backend, `npx vite build` for frontend, `wails build` for full binary

## Architecture

- `internal/ai/ai.go` — Provider/Processor interfaces, registry, shared types
- `internal/ai/groq/` — Groq implementation (provider.go + processor.go)
- `internal/config/config.go` — Settings struct, built-in rules, default profiles, load/save/migration
- `internal/ai/factory/` — Builds Processor from config via provider registry
- `internal/core/engine.go` — Recording + transcription + refinement pipeline
- `internal/ui/desktop/app.go` — Wails-bound App methods
- `internal/ui/desktop/refine.go` — Prompt assembly (profile + built-in rules + user rules + context)

Adding a new AI provider: implement `ai.Provider`, call `ai.RegisterProvider()` in `init()`, add blank import in `factory.go`.

## Frontend conventions

### Imports

```tsx
// Value import when using .createFrom() or accessing namespace as value
import { config } from "../../wailsjs/go/models";
// Type-only import otherwise
import type { ai, history } from "../../wailsjs/go/models";
```

### View wrapper (all settings pages)

```tsx
<div className="flex-1 overflow-y-auto">
  <div className="space-y-4 p-6 max-w-2xl">
```

Always `max-w-2xl`. Always `space-y-4`. Always `p-6`.

### Card anatomy

```tsx
<Card>
  <CardHeader className="pb-3">
    <CardTitle className="flex items-center gap-2 text-sm">
      <Icon className="size-4" /> Title <InfoTip text="..." />
    </CardTitle>
    <CardDescription>Subtitle.</CardDescription>
  </CardHeader>
  <CardContent className="space-y-4">
    {/* fields */}
  </CardContent>
</Card>
```

- CardHeader: always `pb-3`
- CardTitle: always `flex items-center gap-2 text-sm`
- CardTitle icon: always `size-4`
- CardContent: `space-y-4` standard, `space-y-3` for compact sections
- Danger cards: `<Card className="border-destructive/30">`, title gets `text-destructive`
- Active cards: `<Card className={isActive ? "border-primary" : ""}>`

### Buttons

| Use | Variant | Size |
|-----|---------|------|
| Save / primary action | `default` | `sm` |
| Full-width CTA | `default` | default + `className="w-full"` |
| Secondary (restore, test) | `outline` | `sm` |
| Destructive (delete/reset) | `destructive` | `sm` — always wrap in `<ConfirmDialog>` |
| Subtle (activate, nav) | `ghost` | `sm` |
| Expand/collapse chevron | `ghost` | `icon` + `className="size-7"` |
| Tiny icon-only (delete example) | `ghost` | `icon` + `className="size-5"` |

Icon before text: `<Trash2 className="size-3.5" /> Delete`. Loading: `<Loader2 className="size-4 animate-spin" />`.

### Icons

| Context | Size |
|---------|------|
| Card title, sidebar main | `size-4` |
| Sidebar sub-item, button inline | `size-3.5` |
| Advanced toggle, chevron | `size-3` |
| Warning inline | `size-3` |
| Warning banner | `size-4 shrink-0` |
| Empty state | `size-10 opacity-30` |

Always use `AlertTriangle` (not `TriangleAlert`) for warnings. Always `shrink-0` on icons in flex rows.

### Form fields

Label + input: `<div className="space-y-1"><Label>Name</Label><Input .../></div>`

Textarea (raw HTML, no component):
```tsx
<textarea
  rows={3}
  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
/>
```

Smaller textareas (inside examples): same but `px-2 py-1 text-xs rounded`.

Password + reveal toggle: Input with `rounded-r-none`, Button with `rounded-l-none border-l-0`.

### Switch rows (preferences)

```tsx
<div className="flex items-center justify-between">
  <div className="space-y-0.5">
    <Label htmlFor="id">Name</Label>
    <p className="text-xs text-muted-foreground">Description</p>
  </div>
  <Switch id="id" checked={v} onCheckedChange={...} />
</div>
```

Separate consecutive rows with `<Separator />`.

### Danger zone rows

Same layout as switch rows but with ConfirmDialog trigger button on the right:
```tsx
<div className="flex items-center justify-between">
  <div className="space-y-0.5">
    <p className="text-sm font-medium">Action name</p>
    <p className="text-xs text-muted-foreground">Description</p>
  </div>
  <ConfirmDialog trigger={<Button variant="destructive" size="sm">Action</Button>} ... />
</div>
```

### Advanced/collapsible toggle

```tsx
<button
  type="button"
  className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
  onClick={() => setOpen(!open)}
>
  <Settings2 className="size-3" />
  Advanced
  {open ? <ChevronUp className="size-3" /> : <ChevronDown className="size-3" />}
</button>
{open && (
  <div className="space-y-2 rounded-md border border-border bg-muted/30 p-3">
    {/* content */}
  </div>
)}
```

### Warning banners

```tsx
<div className="flex items-center gap-3 rounded-md border border-amber-600/30 bg-amber-600/10 p-3 text-sm text-amber-500">
  <AlertTriangle className="size-4 shrink-0" />
  Message text.
</div>
```

Always `rounded-md`. Always `p-3`. Always `text-sm`.

### Separator usage

- Between major card groups: standalone `<Separator />` in the `space-y-4` wrapper
- Within cards (preference lists): `<Separator />` inside CardContent between rows
- Never inside CardHeader

### Toasts

```tsx
toast.success("Short confirmation");
toast.error(`Failed: ${err}`);
```

### State patterns

- Boolean toggles (switches): save immediately, no Save button
- Text/select fields: draft state + explicit Save button
- Use `config.Settings.createFrom({...})` when constructing settings with nested objects
- Simple flat patches: `{ ...settings, ...patch } as config.Settings` is fine

### Wails bindings

- `wailsjs/go/models.ts` is auto-generated — never edit by hand (Wails regenerates on build)
- `wailsjs/go/desktop/App.js` + `App.d.ts` — update when adding new Go methods
- All Go calls via `import * as AppBridge from "../../wailsjs/go/desktop/App"`
- Always try/catch + `toast.error()` on failure

### External links

Use `<ExternalLink>` for any link that opens in the system browser. Never use `window.open()` (broken in Wails) or raw `<button>` styled as a link (wrong cursor, wrong semantics).

```tsx
import { ExternalLink } from "@/components/ui/external-link";

<ExternalLink href="https://example.com" className="text-xs">
  Link text
</ExternalLink>
```

## Don't

- Don't add `font-mono` — it's the global font
- Don't use `max-w-xl` or `max-w-3xl` for views
- Don't use `space-y-6` in view wrappers
- Don't use destructive buttons without ConfirmDialog
- Don't use `TriangleAlert` — use `AlertTriangle`
- Don't forget `shrink-0` on flex icons
- Don't use `gap-1.5` in CardTitle — always `gap-2`
- Don't use `window.open()` or `<button>` for external links — use `<ExternalLink>`
