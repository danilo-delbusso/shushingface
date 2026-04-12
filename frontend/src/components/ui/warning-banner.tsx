import { AlertTriangle } from "lucide-react";

interface WarningBannerProps {
  children: React.ReactNode;
}

export function WarningBanner({ children }: WarningBannerProps) {
  return (
    <div className="flex items-center gap-3 rounded-md border border-amber-600/30 bg-amber-600/10 p-3 text-sm text-amber-500">
      <AlertTriangle className="size-4 shrink-0" />
      {children}
    </div>
  );
}
