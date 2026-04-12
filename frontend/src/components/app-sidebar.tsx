import {
  Mic,
  History,
  Settings,
  SlidersHorizontal,
  LogOut,
  Bot,
  Palette,
  Plug,
  ChevronRight,
  AlertTriangle,
  ArrowUpCircle,
} from "lucide-react";
import { ExternalLink } from "@/components/ui/external-link";
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
import { ConfirmDialog } from "@/components/confirm-dialog";
import { Quit } from "../../wailsjs/runtime/runtime";
import { Collapsible } from "radix-ui";

export type View = "home" | "history" | "connections" | "ai" | "appearance" | "general";

interface AppSidebarProps {
  view: View;
  onNavigate: (view: View) => void;
  configured: boolean;
  historyEnabled: boolean;
  hasWarnings?: boolean;
  version?: string;
  updateAvailable?: { version: string; url: string } | null;
}

export function AppSidebar({
  view,
  onNavigate,
  configured,
  historyEnabled,
  hasWarnings,
  version,
  updateAvailable,
}: AppSidebarProps) {
  const settingsOpen = view === "connections" || view === "ai" || view === "appearance" || view === "general";

  return (
    <Sidebar collapsible="none" variant="sidebar">
      <SidebarHeader className="flex items-center justify-center gap-2 py-4">
        <img src="/appicon.png" alt="shushingface" className="size-8 rounded-md" />
        <span className="text-lg font-bold text-primary">shushing face</span>
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
                      onClick={() => onNavigate("connections")}
                      isActive={settingsOpen}
                    >
                      <Settings className="size-4" />
                      <span>Settings</span>
                      {(!configured || hasWarnings) && (
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
                          isActive={view === "connections"}
                          onClick={() => onNavigate("connections")}
                        >
                          <Plug className="size-3.5" />
                          <span>Connections</span>
                          {!configured && (
                            <AlertTriangle className="size-3 text-amber-500 ml-auto" />
                          )}
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                      <SidebarMenuSubItem>
                        <SidebarMenuSubButton
                          isActive={view === "ai"}
                          onClick={() => onNavigate("ai")}
                        >
                          <Bot className="size-3.5" />
                          <span>AI</span>
                          {hasWarnings && (
                            <AlertTriangle className="size-3 text-amber-500 ml-auto" />
                          )}
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                      <SidebarMenuSubItem>
                        <SidebarMenuSubButton
                          isActive={view === "appearance"}
                          onClick={() => onNavigate("appearance")}
                        >
                          <Palette className="size-3.5" />
                          <span>Appearance</span>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                      <SidebarMenuSubItem>
                        <SidebarMenuSubButton
                          isActive={view === "general"}
                          onClick={() => onNavigate("general")}
                        >
                          <SlidersHorizontal className="size-3.5" />
                          <span>General</span>
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
        {updateAvailable && (
          <ExternalLink
            href={updateAvailable.url}
            className="flex items-center gap-1.5 rounded-md border border-primary/30 bg-primary/5 px-3 py-2 text-xs text-primary no-underline hover:bg-primary/10"
          >
            <ArrowUpCircle className="size-3.5 shrink-0" />
            {updateAvailable.version} available
          </ExternalLink>
        )}
        {version && (
          <p className="px-3 pb-1 text-[10px] text-muted-foreground/50">{version}</p>
        )}
        <SidebarMenu>
          <SidebarMenuItem>
            <ConfirmDialog
              trigger={
                <SidebarMenuButton className="text-muted-foreground">
                  <LogOut className="size-4" />
                  <span>Quit</span>
                </SidebarMenuButton>
              }
              title="Quit shushing face?"
              description="The app will close completely and stop running in the background."
              confirmLabel="Quit"
              onConfirm={() => Quit()}
            />
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
}
