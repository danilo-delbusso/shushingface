import {
  Mic,
  History,
  Settings,
  SlidersHorizontal,
  MessageSquareText,
  ChevronRight,
} from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarHeader,
  SidebarFooter,
} from "@/components/ui/sidebar";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Collapsible } from "radix-ui";

export type View = "home" | "history" | "settings" | "prompt";

interface AppSidebarProps {
  view: View;
  onNavigate: (view: View) => void;
  configured: boolean;
  historyEnabled: boolean;
  platform: { os: string; desktop: string } | null;
}

export function AppSidebar({
  view,
  onNavigate,
  configured,
  historyEnabled,
  platform,
}: AppSidebarProps) {
  const settingsOpen = view === "settings" || view === "prompt";

  return (
    <Sidebar collapsible="none" variant="sidebar">
      <SidebarHeader className="flex items-center justify-center gap-2 py-4">
        <img src="/appicon.png" alt="sussurro" className="size-8 rounded-md" />
        <span className="text-lg font-bold text-primary">sussurro</span>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {/* Record */}
              <SidebarMenuItem>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <SidebarMenuButton
                      isActive={view === "home"}
                      onClick={() => onNavigate("home")}
                    >
                      <Mic className="size-4" />
                      <span>Record</span>
                    </SidebarMenuButton>
                  </TooltipTrigger>
                  <TooltipContent side="right">Record</TooltipContent>
                </Tooltip>
              </SidebarMenuItem>

              {/* History (conditional) */}
              {historyEnabled && (
                <SidebarMenuItem>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <SidebarMenuButton
                        isActive={view === "history"}
                        onClick={() => onNavigate("history")}
                      >
                        <History className="size-4" />
                        <span>History</span>
                      </SidebarMenuButton>
                    </TooltipTrigger>
                    <TooltipContent side="right">History</TooltipContent>
                  </Tooltip>
                </SidebarMenuItem>
              )}

              {/* Settings group with sub-items */}
              <Collapsible.Root defaultOpen={settingsOpen} open={settingsOpen}>
                <SidebarMenuItem>
                  <Collapsible.Trigger asChild>
                    <SidebarMenuButton
                      onClick={() => onNavigate("settings")}
                      isActive={settingsOpen}
                    >
                      <Settings className="size-4" />
                      <span>Settings</span>
                      {!configured && (
                        <Badge
                          variant="destructive"
                          className="ml-auto size-5 justify-center rounded-full p-0 text-[10px]"
                        >
                          !
                        </Badge>
                      )}
                      <ChevronRight className="ml-auto size-4 transition-transform group-data-[state=open]/collapsible:rotate-90" />
                    </SidebarMenuButton>
                  </Collapsible.Trigger>
                  <Collapsible.Content>
                    <SidebarMenuSub>
                      <SidebarMenuSubItem>
                        <SidebarMenuSubButton
                          isActive={view === "settings"}
                          onClick={() => onNavigate("settings")}
                        >
                          <SlidersHorizontal className="size-3.5" />
                          <span>General</span>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                      <SidebarMenuSubItem>
                        <SidebarMenuSubButton
                          isActive={view === "prompt"}
                          onClick={() => onNavigate("prompt")}
                        >
                          <MessageSquareText className="size-3.5" />
                          <span>Refinement</span>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                    </SidebarMenuSub>
                  </Collapsible.Content>
                </SidebarMenuItem>
              </Collapsible.Root>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        {platform?.os === "linux" && (
          <div className="px-2 py-3 text-center">
            <p className="text-xs text-muted-foreground">
              <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] font-mono">
                Super+Ctrl+B
              </kbd>{" "}
              to toggle
            </p>
          </div>
        )}
      </SidebarFooter>
    </Sidebar>
  );
}
