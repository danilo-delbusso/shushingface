import { useState } from "react";
import { Plus, RotateCcw, MoreHorizontal } from "lucide-react";
import { Popover } from "radix-ui";
import { toast } from "sonner";
import type { ProfileFormData } from "@/lib/schemas";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { InfoTip } from "@/components/info-tip";
import { ProfileCard } from "@/components/profile-card";
import { getProfileIcon } from "@/lib/icons";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { config } from "../../wailsjs/go/models";

interface StylesViewProps {
  settings: config.Settings;
  onSave: (settings: config.Settings) => void;
}

const PRESET_IDS = new Set(["casual", "professional", "concise"]);

export function StylesView({ settings, onSave }: StylesViewProps) {
  const connections = settings.connections ?? [];
  const refConnId = settings.refinementConnectionId;
  const refModel = settings.refinementModel;

  const [draftProfiles, setDraftProfiles] = useState(
    settings.refinementProfiles ?? [],
  );
  const [activeId, setActiveId] = useState(settings.activeProfileId);
  const [expandedProfile, setExpandedProfile] = useState<string | null>(null);
  const [advancedOpen, setAdvancedOpen] = useState<string | null>(null);

  const persist = (
    profiles?: typeof draftProfiles,
    newActiveId?: string,
  ) => {
    const p = profiles ?? draftProfiles;
    const a = newActiveId ?? activeId;
    onSave(
      config.Settings.createFrom({
        ...settings,
        refinementProfiles: p,
        activeProfileId: a,
      }),
    );
  };

  const saveProfile = (id: string, data: ProfileFormData) => {
    const updated = draftProfiles.map((p) =>
      p.id === id
        ? config.RefinementProfile.createFrom({
            ...p,
            ...data,
            connectionId: data.connectionId || undefined,
          })
        : p,
    );
    setDraftProfiles(updated);
    persist(updated);
  };

  const deleteProfile = (id: string) => {
    const updated = draftProfiles.filter((p) => p.id !== id);
    const newActive = activeId === id ? (updated[0]?.id ?? "") : activeId;
    setDraftProfiles(updated);
    setActiveId(newActive);
    persist(updated, newActive);
  };

  const addProfile = () => {
    const id = `custom-${Date.now()}`;
    const newProfile = config.RefinementProfile.createFrom({
      id,
      name: "New Style",
      icon: "pen-tool",
      model: "",
      prompt: "",
    });
    const updated = [...draftProfiles, newProfile];
    setDraftProfiles(updated);
    setExpandedProfile(id);
  };

  const setActive = (id: string) => {
    setActiveId(id);
    persist(undefined, id);
  };

  const applyDefaultsToAll = () => {
    const updated = draftProfiles.map((p) =>
      config.RefinementProfile.createFrom({
        ...p,
        connectionId: undefined,
        model: "",
      }),
    );
    setDraftProfiles(updated);
    persist(updated);
    toast.success("All styles now use global defaults");
  };

  const restoreDefaults = async () => {
    const defaults = await AppBridge.GetDefaultProfiles();
    const defaultIds = new Set(defaults.map((d) => d.id));
    const custom = draftProfiles.filter((p) => !defaultIds.has(p.id));
    const updated = [...defaults, ...custom];
    setDraftProfiles(updated);
    const newActive = defaultIds.has(activeId)
      ? activeId
      : (updated[0]?.id ?? "");
    setActiveId(newActive);
    persist(updated, newActive);
    toast.success("Default styles restored");
  };

  const restoreProfile = async (id: string) => {
    const defaults = await AppBridge.GetDefaultProfiles();
    const def = defaults.find((d) => d.id === id);
    if (!def) return;
    const updated = draftProfiles.map((p) => (p.id === id ? def : p));
    setDraftProfiles(updated);
    persist(updated);
    toast.success(`"${def.name}" restored to default`);
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-3 p-6 max-w-2xl">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold flex items-center gap-2">
            Refinement Styles{" "}
            <InfoTip text="Each style defines how your speech gets rewritten. Styles can override the connection and model." />
          </h3>
          <div className="flex items-center gap-1.5">
            <Button variant="outline" size="sm" onClick={addProfile}>
              <Plus className="size-3.5" /> Add
            </Button>
            <StylesOverflowMenu
              onApplyDefaults={applyDefaultsToAll}
              onRestoreDefaults={restoreDefaults}
            />
          </div>
        </div>

        {draftProfiles.map((profile) => {
          const Icon = getProfileIcon(profile.icon);
          const isActive = profile.id === activeId;
          const isExpanded = expandedProfile === profile.id;
          const isPreset = PRESET_IDS.has(profile.id);

          return (
            <ProfileCard
              key={profile.id}
              profile={profile}
              Icon={Icon}
              isActive={isActive}
              isExpanded={isExpanded}
              isPreset={isPreset}
              connections={connections}
              defaultRefConnId={refConnId}
              defaultRefModel={refModel}
              advancedOpen={advancedOpen === profile.id}
              onToggleExpand={() =>
                setExpandedProfile(isExpanded ? null : profile.id)
              }
              onToggleAdvanced={(v) => setAdvancedOpen(v ? profile.id : null)}
              onActivate={() => setActive(profile.id)}
              onSave={(data) => saveProfile(profile.id, data)}
              onRestore={() => restoreProfile(profile.id)}
              onDelete={() => deleteProfile(profile.id)}
            />
          );
        })}
      </div>
    </div>
  );
}

function StylesOverflowMenu({
  onApplyDefaults,
  onRestoreDefaults,
}: {
  onApplyDefaults: () => void;
  onRestoreDefaults: () => void;
}) {
  const [open, setOpen] = useState(false);
  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      <Popover.Trigger asChild>
        <Button variant="ghost" size="icon" className="size-7">
          <MoreHorizontal className="size-4" />
        </Button>
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          align="end"
          sideOffset={4}
          className="z-50 w-56 rounded-md border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95"
        >
          <ConfirmDialog
            trigger={
              <button
                type="button"
                className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-xs hover:bg-accent hover:text-accent-foreground"
              >
                Apply defaults to all
              </button>
            }
            title="Apply defaults to all styles?"
            description="This clears connection and model overrides on every style so they all use the global defaults from Models."
            confirmLabel="Apply"
            variant="default"
            onConfirm={() => {
              onApplyDefaults();
              setOpen(false);
            }}
          />
          <ConfirmDialog
            trigger={
              <button
                type="button"
                className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-xs hover:bg-accent hover:text-accent-foreground"
              >
                <RotateCcw className="size-3.5" /> Restore default styles
              </button>
            }
            title="Restore default styles?"
            description="This will replace the built-in styles with their defaults. Custom styles will be kept."
            confirmLabel="Restore"
            onConfirm={() => {
              onRestoreDefaults();
              setOpen(false);
            }}
          />
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
