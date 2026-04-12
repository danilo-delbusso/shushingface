import { useState, useEffect } from "react";
import { Eye, EyeOff, ArrowRight, Check, Plug } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getProfileIcon } from "@/lib/icons";
import { providerPresets } from "@/lib/providers";
import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { config } from "../../wailsjs/go/models";
import type { ai } from "../../wailsjs/go/models";

interface WelcomeWizardProps {
  settings: config.Settings;
  onComplete: (settings: config.Settings) => void;
}

export function WelcomeWizard({ settings, onComplete }: WelcomeWizardProps) {
  const [step, setStep] = useState(0);
  const [providers, setProviders] = useState<ai.ProviderInfo[]>([]);
  const [providerId, setProviderId] = useState(settings.providerId || "groq");
  const [apiKey, setApiKey] = useState(settings.providerApiKey ?? "");
  const [showKey, setShowKey] = useState(false);
  const [selectedProfile, setSelectedProfile] = useState("professional");

  useEffect(() => {
    AppBridge.ListProviders().then(setProviders);
  }, []);

  const preset = providerPresets[providerId];

  const finish = () => {
    onComplete(
      config.Settings.createFrom({
        ...settings,
        providerId,
        providerApiKey: apiKey,
        activeProfileId: selectedProfile,
        setupComplete: true,
      }),
    );
  };

  const totalSteps = 3;

  return (
    <div className="flex h-screen w-screen flex-col justify-center bg-background">
      <div className="mx-auto w-full max-w-md space-y-8 px-6">
        {/* Step 0: Welcome */}
        {step === 0 && (
          <div className="space-y-6 text-center">
            <img src="/appicon.png" alt="" className="mx-auto size-20" />
            <div className="space-y-2">
              <h1 className="text-2xl font-bold">welcome to shushing face</h1>
              <p className="text-sm text-muted-foreground">
                speak naturally, get polished text. let's get you set up.
              </p>
            </div>
            <Button className="w-full" onClick={() => setStep(1)}>
              get started <ArrowRight className="size-4" />
            </Button>
          </div>
        )}

        {/* Step 1: Choose provider + API key */}
        {step === 1 && (
          <div className="space-y-6">
            <div className="space-y-2 text-center">
              <h2 className="text-xl font-bold">connect an AI provider</h2>
              <p className="text-sm text-muted-foreground">
                choose a provider for transcription and refinement.
              </p>
            </div>

            {/* Provider picker */}
            <div className="grid gap-2">
              {providers.map((p) => {
                const meta = providerPresets[p.id];
                const active = p.id === providerId;
                return (
                  <button
                    key={p.id}
                    type="button"
                    onClick={() => setProviderId(p.id)}
                    className={`flex items-center gap-3 overflow-hidden rounded-lg border-2 p-3 text-left transition-colors ${
                      active
                        ? "border-primary bg-primary/5"
                        : "border-border hover:border-muted-foreground/30"
                    }`}
                  >
                    <div
                      className={`flex size-9 shrink-0 items-center justify-center rounded-md ${
                        active
                          ? "bg-primary text-primary-foreground"
                          : "bg-muted text-muted-foreground"
                      }`}
                    >
                      {meta?.icon ? (
                        <img src={meta.icon} alt="" className="size-4" />
                      ) : (
                        <Plug className="size-4" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium">{p.displayName}</p>
                      {meta && (
                        <p className="text-xs text-muted-foreground truncate">
                          {meta.description}
                        </p>
                      )}
                    </div>
                    {active && (
                      <Check className="size-4 text-primary shrink-0" />
                    )}
                  </button>
                );
              })}
            </div>

            {/* API key */}
            <div className="space-y-2">
              {preset && (
                <button
                  type="button"
                  className="text-primary underline underline-offset-2 text-xs"
                  onClick={() => BrowserOpenURL(preset.keyUrl)}
                >
                  {preset.keyUrlLabel}
                </button>
              )}
              <div className="flex">
                <Input
                  type={showKey ? "text" : "password"}
                  value={apiKey}
                  placeholder={preset?.keyPlaceholder ?? "API key..."}
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
                  {showKey ? (
                    <EyeOff className="size-4" />
                  ) : (
                    <Eye className="size-4" />
                  )}
                </Button>
              </div>
            </div>

            <Button
              className="w-full"
              onClick={() => setStep(2)}
              disabled={!apiKey.trim()}
            >
              next <ArrowRight className="size-4" />
            </Button>
          </div>
        )}

        {/* Step 2: Choose style */}
        {step === 2 && (
          <div className="space-y-6">
            <div className="space-y-2 text-center">
              <h2 className="text-xl font-bold">choose your style</h2>
              <p className="text-sm text-muted-foreground">
                how should your speech be refined? you can change this anytime.
              </p>
            </div>
            <div className="grid gap-3">
              {settings.refinementProfiles?.map((profile) => {
                const Icon = getProfileIcon(profile.icon);
                const active = selectedProfile === profile.id;
                return (
                  <button
                    key={profile.id}
                    type="button"
                    onClick={() => setSelectedProfile(profile.id)}
                    className={`flex items-center gap-4 rounded-lg border-2 p-4 text-left transition-colors ${
                      active
                        ? "border-primary bg-primary/5"
                        : "border-border hover:border-muted-foreground/30"
                    }`}
                  >
                    <div
                      className={`flex size-10 items-center justify-center rounded-lg ${
                        active
                          ? "bg-primary text-primary-foreground"
                          : "bg-muted text-muted-foreground"
                      }`}
                    >
                      <Icon className="size-5" />
                    </div>
                    <div className="flex-1">
                      <p className="font-medium">{profile.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {profile.id === "casual" && "friendly and relaxed"}
                        {profile.id === "professional" && "clear and polished"}
                        {profile.id === "concise" && "brief and to-the-point"}
                      </p>
                    </div>
                    {active && <Check className="size-5 text-primary" />}
                  </button>
                );
              })}
            </div>
            <Button className="w-full" onClick={finish}>
              finish <Check className="size-4" />
            </Button>
          </div>
        )}

        {/* Step indicators */}
        <div className="flex justify-center gap-2">
          {Array.from({ length: totalSteps + 1 }, (_, i) => (
            <div
              key={i}
              className={`size-1.5 rounded-full transition-colors ${
                i === step ? "bg-primary" : "bg-muted"
              }`}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
