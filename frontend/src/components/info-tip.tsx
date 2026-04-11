import { Info } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface InfoTipProps {
  text: string;
}

export function InfoTip({ text }: InfoTipProps) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Info className="inline size-3.5 text-muted-foreground/50 hover:text-muted-foreground cursor-help" />
      </TooltipTrigger>
      <TooltipContent side="top" className="max-w-xs text-xs">
        {text}
      </TooltipContent>
    </Tooltip>
  );
}
