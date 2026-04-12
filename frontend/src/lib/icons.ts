import { Coffee, Briefcase, Zap, PenTool } from "lucide-react";

/** Maps profile icon string keys to lucide components. */
export const profileIconMap: Record<
  string,
  React.FC<{ className?: string }>
> = {
  coffee: Coffee,
  briefcase: Briefcase,
  zap: Zap,
  "pen-tool": PenTool,
};

/** Returns the icon component for a profile, falling back to PenTool. */
export function getProfileIcon(key: string | undefined) {
  return profileIconMap[key ?? ""] ?? PenTool;
}
