import { Mic, History, Settings } from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarHeader,
  SidebarFooter,
} from "@/components/ui/sidebar";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

export type View = "home" | "history" | "settings";

interface AppSidebarProps {
  view: View;
  onNavigate: (view: View) => void;
  configured: boolean;
  historyEnabled: boolean;
  hotkey?: string;
}

export function AppSidebar({
  view,
  onNavigate,
  configured,
  historyEnabled,
  hotkey,
}: AppSidebarProps) {
  const navItems = [
    { id: "home" as View, label: "Record", icon: Mic },
    ...(historyEnabled
      ? [{ id: "history" as View, label: "History", icon: History }]
      : []),
    { id: "settings" as View, label: "Settings", icon: Settings },
  ];

  return (
    <Sidebar collapsible="none" variant="sidebar">
      <SidebarHeader className="flex items-center justify-center py-4">
        <span className="text-xl font-bold text-primary">Sussurro</span>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.id}>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <SidebarMenuButton
                        isActive={view === item.id}
                        onClick={() => onNavigate(item.id)}
                      >
                        <item.icon className="size-4" />
                        <span>{item.label}</span>
                        {item.id === "settings" && !configured && (
                          <Badge
                            variant="destructive"
                            className="ml-auto size-5 justify-center rounded-full p-0 text-[10px]"
                          >
                            !
                          </Badge>
                        )}
                      </SidebarMenuButton>
                    </TooltipTrigger>
                    <TooltipContent side="right">{item.label}</TooltipContent>
                  </Tooltip>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        {hotkey && (
          <div className="px-2 py-3 text-center">
            <p className="text-xs text-muted-foreground">
              <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] font-mono">
                {hotkey}
              </kbd>{" "}
              to record
            </p>
          </div>
        )}
      </SidebarFooter>
    </Sidebar>
  );
}
