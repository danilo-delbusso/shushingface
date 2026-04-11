import { useState } from "react";
import {
  Play,
  Loader2,
  Key,
  Eye,
  EyeOff,
  AlertTriangle,
  Coffee,
  Briefcase,
  Zap,
  PenTool,
  Check,
  Trash2,
  Plus,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ConfirmDialog } from "@/components/confirm-dialog";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { config } from "../../wailsjs/go/models";

const iconMap: Record<string, React.FC<{ className?: string }>> = {
  coffee: Coffee,
  briefcase: Briefcase,
  zap: Zap,
  "pen-tool": PenTool,
};

interface AiViewProps {
  settings: config.Settings;
  configured: boolean;
  onSave: (settings: config.Settings) => void;
}

export function AiView({ settings, configured, onSave }: AiViewProps) {
  const [showKey, setShowKey] = useState(false);
  const [apiKey, setApiKey] = useState(
    settings.providers?.[settings.transcriptionProviderId]?.apiKey ?? "",
  );
  const [transModel, setTransModel] = useState(settings.transcriptionModel);
  const [sampleText, setSampleText] = useState("");
  const [testResult, setTestResult] = useState("");
  const [testing, setTesting] = useState(false);
  const [expandedProfile, setExpandedProfile] = useState<string | null>(null);

  const profiles = settings.refinementProfiles ?? [];
  const activeId = settings.activeProfileId;

  const saveAll = (
    newProfiles?: typeof profiles,
    newActiveId?: string,
  ) => {
    const providerId = settings.transcriptionProviderId;
    onSave(
      config.Settings.createFrom({
        ...settings,
        providers: {
          ...settings.providers,
          [providerId]: { ...settings.providers[providerId], apiKey },
        },
        transcriptionModel: transModel,
        refinementProfiles: newProfiles ?? profiles,
        activeProfileId: newActiveId ?? activeId,
      }),
    );
  };

  const updateProfile = (id: string, patch: Partial<config.RefinementProfile>) => {
    const updated = profiles.map((p) =>
      p.id === id ? config.RefinementProfile.createFrom({ ...p, ...patch }) : p,
    );
    saveAll(updated);
  };

  const deleteProfile = (id: string) => {
    const updated = profiles.filter((p) => p.id !== id);
    const newActive = activeId === id ? updated[0]?.id ?? "" : activeId;
    saveAll(updated, newActive);
  };

  const addProfile = () => {
    const id = `custom-${Date.now()}`;
    const newProfile = config.RefinementProfile.createFrom({
      id,
      name: "New Style",
      icon: "pen-tool",
      model: profiles[0]?.model ?? "llama-3.3-70b-versatile",
      prompt: "",
    });
    saveAll([...profiles, newProfile]);
    setExpandedProfile(id);
  };

  const setActive = (id: string) => {
    saveAll(undefined, id);
  };

  const placeholderText =
    "hey um so I was thinking we should probably move the meeting to thursday because like john cant make it on wednesday and I think it would be better if we all met together you know";

  const handleTest = async () => {
    const text = sampleText.trim() || placeholderText;
    const profile = profiles.find((p) => p.id === activeId);
    if (!profile?.prompt) {
      toast.error("Active profile has no prompt");
      return;
    }
    setTesting(true);
    setTestResult("");
    try {
      const res = await AppBridge.TestPrompt(text, profile.prompt);
      if (res.error) toast.error(res.error);
      else setTestResult(res.refined);
    } catch (err) {
      toast.error(`Test failed: ${err}`);
    } finally {
      setTesting(false);
    }
  };

  const presetIds = new Set(["casual", "professional", "concise"]);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        {!configured && (
          <div className="flex items-center gap-3 rounded-lg border border-amber-600/30 bg-amber-600/10 p-3 text-sm text-amber-500">
            <AlertTriangle className="size-4 shrink-0" />
            Add your API key to get started.
          </div>
        )}

        {/* API Key */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Key className="size-4" /> API Provider
            </CardTitle>
            <CardDescription>
              shushing face uses Groq for speech-to-text.{" "}
              <button
                type="button"
                className="text-primary underline underline-offset-2 hover:text-primary/80"
                onClick={() => window.open("https://console.groq.com/keys", "_blank")}
              >
                Get a free key
              </button>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex">
              <Input
                type={showKey ? "text" : "password"}
                value={apiKey}
                placeholder="gsk_..."
                onChange={(e) => setApiKey(e.target.value)}
                className="rounded-r-none"
              />
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={() => setShowKey(!showKey)}
                className="rounded-l-none border-l-0"
              >
                {showKey ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
              </Button>
            </div>
            <div className="mt-3 space-y-1">
              <Label htmlFor="trans-model">Transcription Model</Label>
              <Input
                id="trans-model"
                value={transModel}
                onChange={(e) => setTransModel(e.target.value)}
              />
            </div>
            <Button
              size="sm"
              className="mt-3"
              onClick={() => saveAll()}
            >
              Save
            </Button>
          </CardContent>
        </Card>

        <Separator />

        {/* Profiles */}
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold">Refinement Styles</h3>
          <Button variant="outline" size="sm" onClick={addProfile}>
            <Plus className="size-3.5" /> Add
          </Button>
        </div>

        {profiles.map((profile) => {
          const Icon = iconMap[profile.icon] || PenTool;
          const isActive = profile.id === activeId;
          const isExpanded = expandedProfile === profile.id;
          const isPreset = presetIds.has(profile.id);

          return (
            <Card
              key={profile.id}
              className={isActive ? "border-primary" : ""}
            >
              <CardHeader className="pb-2">
                <div className="flex items-center gap-3">
                  <div
                    className={`flex size-8 items-center justify-center rounded-md ${
                      isActive
                        ? "bg-primary text-primary-foreground"
                        : "bg-muted text-muted-foreground"
                    }`}
                  >
                    <Icon className="size-4" />
                  </div>
                  <div className="flex-1">
                    <CardTitle className="text-sm">{profile.name}</CardTitle>
                    <CardDescription className="text-xs">
                      {profile.model}
                    </CardDescription>
                  </div>
                  <div className="flex items-center gap-1">
                    {!isActive && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setActive(profile.id)}
                      >
                        <Check className="size-3.5" /> Use
                      </Button>
                    )}
                    {isActive && (
                      <span className="text-xs font-medium text-primary">active</span>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-7"
                      onClick={() =>
                        setExpandedProfile(isExpanded ? null : profile.id)
                      }
                    >
                      {isExpanded ? (
                        <ChevronUp className="size-3.5" />
                      ) : (
                        <ChevronDown className="size-3.5" />
                      )}
                    </Button>
                  </div>
                </div>
              </CardHeader>
              {isExpanded && (
                <CardContent className="space-y-3 pt-0">
                  <div className="space-y-1">
                    <Label>Name</Label>
                    <Input
                      value={profile.name}
                      onChange={(e) =>
                        updateProfile(profile.id, { name: e.target.value })
                      }
                    />
                  </div>
                  <div className="space-y-1">
                    <Label>Model</Label>
                    <Input
                      value={profile.model}
                      onChange={(e) =>
                        updateProfile(profile.id, { model: e.target.value })
                      }
                    />
                  </div>
                  <div className="space-y-1">
                    <Label>Prompt</Label>
                    <textarea
                      value={profile.prompt}
                      onChange={(e) =>
                        updateProfile(profile.id, { prompt: e.target.value })
                      }
                      rows={6}
                      className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                    />
                  </div>
                  {!isPreset && (
                    <ConfirmDialog
                      trigger={
                        <Button variant="destructive" size="sm">
                          <Trash2 className="size-3.5" /> Delete
                        </Button>
                      }
                      title={`Delete "${profile.name}"?`}
                      description="This style will be permanently removed."
                      confirmLabel="Delete"
                      onConfirm={() => deleteProfile(profile.id)}
                    />
                  )}
                </CardContent>
              )}
            </Card>
          );
        })}

        <Separator />

        {/* Test */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Test Playground</CardTitle>
            <CardDescription>
              Tests with the active style.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <textarea
              value={sampleText}
              onChange={(e) => setSampleText(e.target.value)}
              rows={3}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
              placeholder={placeholderText}
            />
            <Button onClick={handleTest} disabled={testing} className="w-full">
              {testing ? (
                <><Loader2 className="size-4 animate-spin" /> Running...</>
              ) : (
                <><Play className="size-4" /> Run Test</>
              )}
            </Button>
            {testResult && (
              <div className="rounded-lg bg-muted/50 p-3">
                <p className="text-xs text-muted-foreground mb-1">Result</p>
                <p className="text-sm leading-relaxed whitespace-pre-wrap">{testResult}</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
