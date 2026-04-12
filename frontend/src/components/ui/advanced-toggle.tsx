import { Settings2, ChevronDown, ChevronUp } from "lucide-react";

interface AdvancedToggleProps {
  label?: string;
  open: boolean;
  onToggle: (open: boolean) => void;
  children: React.ReactNode;
}

export function AdvancedToggle({
  label = "Advanced",
  open,
  onToggle,
  children,
}: AdvancedToggleProps) {
  return (
    <>
      <button
        type="button"
        className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
        onClick={() => onToggle(!open)}
      >
        <Settings2 className="size-3" />
        {label}
        {open ? (
          <ChevronUp className="size-3" />
        ) : (
          <ChevronDown className="size-3" />
        )}
      </button>
      {open && (
        <div className="space-y-2 rounded-md border border-border bg-muted/30 p-3">
          {children}
        </div>
      )}
    </>
  );
}
