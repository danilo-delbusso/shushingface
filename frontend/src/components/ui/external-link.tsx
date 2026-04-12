import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime";
import { cn } from "@/lib/utils";

interface ExternalLinkProps {
  href: string;
  className?: string;
  children: React.ReactNode;
}

export function ExternalLink({ href, className, children }: ExternalLinkProps) {
  return (
    <a
      href={href}
      onClick={(e) => {
        e.preventDefault();
        BrowserOpenURL(href);
      }}
      className={cn(
        "cursor-pointer text-primary underline underline-offset-2 hover:text-primary/80",
        className,
      )}
    >
      {children}
    </a>
  );
}
