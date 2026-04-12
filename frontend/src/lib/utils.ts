import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/** Standard textarea classes (matches Input styling). */
export const textareaClass =
  "w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y";

/** Compact textarea for nested contexts (examples, code). */
export const textareaCompactClass =
  "w-full rounded border border-input bg-background px-2 py-1 text-xs leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y";
