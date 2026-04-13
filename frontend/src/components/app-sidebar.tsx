import {
  Mic,
  History,
  Settings,
  LogOut,
  ArrowUpCircle,
  PanelLeft,
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
  SidebarHeader,
  SidebarFooter,
  useSidebar,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { Quit } from "../../wailsjs/runtime/runtime";

export type View = "home" | "history";

interface AppSidebarProps {
  view: View;
  onNavigate: (view: View) => void;
  onOpenSettings: () => void;
  historyEnabled: boolean;
  hasWarnings?: boolean;
  version?: string;
  updateAvailable?: { version: string; url: string } | null;
}

export function AppSidebar({
  view,
  onNavigate,
  onOpenSettings,
  historyEnabled,
  hasWarnings,
  version,
  updateAvailable,
}: AppSidebarProps) {
  const { toggleSidebar } = useSidebar();

  return (
    <Sidebar collapsible="icon" variant="sidebar">
      <SidebarHeader className="p-2">
        <div className="flex items-center gap-2 group-data-[collapsible=icon]:justify-center">
          <Button
            variant="ghost"
            size="icon"
            className="size-7 shrink-0"
            onClick={toggleSidebar}
            aria-label="Toggle sidebar"
          >
            <PanelLeft className="size-4" />
          </Button>
          <div className="flex items-center gap-2 overflow-hidden group-data-[collapsible=icon]:hidden">
            <img
              src="/appicon.png"
              alt="shushingface"
              className="size-6 rounded"
            />
            <span className="text-sm font-bold text-primary truncate">
              shushing face
            </span>
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  isActive={view === "home"}
                  onClick={() => onNavigate("home")}
                  tooltip="Record"
                >
                  <Mic className="size-4" />
                  <span>Record</span>
                </SidebarMenuButton>
              </SidebarMenuItem>

              {historyEnabled && (
                <SidebarMenuItem>
                  <SidebarMenuButton
                    isActive={view === "history"}
                    onClick={() => onNavigate("history")}
                    tooltip="History"
                  >
                    <History className="size-4" />
                    <span>History</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        {updateAvailable && (
          <ExternalLink
            href={updateAvailable.url}
            className="flex items-center gap-1.5 rounded-md border border-primary/30 bg-primary/5 px-3 py-2 text-xs text-primary no-underline hover:bg-primary/10 group-data-[collapsible=icon]:hidden"
          >
            <ArrowUpCircle className="size-3.5 shrink-0" />
            {updateAvailable.version} available
          </ExternalLink>
        )}
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton onClick={onOpenSettings} tooltip="Settings">
              <Settings className="size-4" />
              <span>Settings</span>
              {hasWarnings && (
                <span className="ml-auto size-1.5 rounded-full bg-amber-500" />
              )}
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <ConfirmDialog
              trigger={
                <SidebarMenuButton
                  className="text-muted-foreground"
                  tooltip="Quit"
                >
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
        {version && (
          <p className="px-3 text-[10px] text-muted-foreground/50 group-data-[collapsible=icon]:hidden">
            {version}
          </p>
        )}
      </SidebarFooter>
    </Sidebar>
  );
}
