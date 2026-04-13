import { Coffee, Briefcase, Zap, PenTool } from "lucide-react";

const profileIconMap: Record<string, React.FC<{ className?: string }>> = {
  coffee: Coffee,
  briefcase: Briefcase,
  zap: Zap,
  "pen-tool": PenTool,
};

export function getProfileIcon(key: string | undefined) {
  return profileIconMap[key ?? ""] ?? PenTool;
}
