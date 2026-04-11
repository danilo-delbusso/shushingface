import {
  Mic,
  History,
  Settings,
  SlidersHorizontal,
  LogOut,
  Bot,
  Palette,
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
import { ConfirmDialog } from "@/components/confirm-dialog";
import { Quit } from "../../wailsjs/runtime/runtime";
import { Collapsible } from "radix-ui";

export type View = "home" | "history" | "ai" | "appearance" | "general";

interface AppSidebarProps {
  view: View;
  onNavigate: (view: View) => void;
  configured: boolean;
  historyEnabled: boolean;
}

export function AppSidebar({
  view,
  onNavigate,
  configured,
  historyEnabled,
}: AppSidebarProps) {
  const settingsOpen = view === "ai" || view === "appearance" || view === "general";

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
                      onClick={() => onNavigate("ai")}
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
                          isActive={view === "ai"}
                          onClick={() => onNavigate("ai")}
                        >
                          <Bot className="size-3.5" />
                          <span>AI</span>
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
