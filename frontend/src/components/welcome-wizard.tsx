import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  wizardConnectionSchema,
  type WizardConnectionFormData,
} from "@/lib/schemas";
import { FormField } from "@/components/ui/form-field";
import {
  Eye,
  EyeOff,
  ArrowRight,
  Check,
  Plug,
  SkipForward,
  Loader2,
  RefreshCw,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getProfileIcon } from "@/lib/icons";
import { providerPresets } from "@/lib/providers";
import { ExternalLink } from "@/components/ui/external-link";
import { InfoTip } from "@/components/info-tip";
import { ShortcutGuide } from "@/components/shortcut-guide";
import { usePlatform } from "@/lib/hooks";
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
  const [showKey, setShowKey] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testOk, setTestOk] = useState(false);
  const [selectedProfile, setSelectedProfile] = useState("professional");
  const platformInfo = usePlatform();

  const {
    register,
    watch,
    setValue,
    trigger,
    getValues,
    formState: { errors },
  } = useForm<WizardConnectionFormData>({
    resolver: zodResolver(wizardConnectionSchema),
    defaultValues: { providerId: "groq", apiKey: "", baseUrl: "" },
  });

  const providerId = watch("providerId");
  const apiKey = watch("apiKey");
  const baseUrl = watch("baseUrl");

  useEffect(() => {
    AppBridge.ListProviders().then(setProviders);
  }, []);

  const preset = providerPresets[providerId];
  const needsBaseUrl = preset?.requiresBaseUrl ?? false;
  const canProceed = needsBaseUrl ? !!baseUrl?.trim() : !!apiKey?.trim();

  const advanceStep1 = async () => {
    const valid = await trigger();
    if (valid) setStep(2);
  };

  const testConnection = async () => {
    const { providerId: pId, apiKey: key, baseUrl: url } = getValues();
    const connId = `test_${Date.now()}`;
    const conn = config.Connection.createFrom({
      id: connId,
      name: "test",
      providerId: pId,
      apiKey: key,
      baseUrl: url || undefined,
    });
    await AppBridge.SaveSettings(
      config.Settings.createFrom({ ...settings, connections: [conn] }),
    );
    setTesting(true);
    setTestOk(false);
    try {
      const models = await AppBridge.ListModelsForConnection(connId);
      setTestOk(true);
      toast.success(`Connected — ${models?.length ?? 0} models available`);
    } catch (err) {
      toast.error(`Connection failed: ${err}`);
    } finally {
      setTesting(false);
    }
  };

  const finishWithConnection = () => {
    const { providerId: pId, apiKey: key, baseUrl: url } = getValues();
    const connId = `conn_${Date.now()}`;
    const connName = providerPresets[pId]?.name ?? "Default";
    const conn = config.Connection.createFrom({
      id: connId,
      name: connName,
      providerId: pId,
      apiKey: key,
      baseUrl: url || undefined,
    });
    onComplete(
      config.Settings.createFrom({
        ...settings,
        connections: [conn],
        transcriptionConnectionId: connId,
        refinementConnectionId: connId,
        activeProfileId: selectedProfile,
        setupComplete: true,
      }),
    );
  };

  const finishSkip = () => {
    onComplete(
      config.Settings.createFrom({
        ...settings,
        activeProfileId: selectedProfile,
        setupComplete: true,
      }),
    );
  };

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

        {/* Step 1: Add connection or skip */}
        {step === 1 && (
          <div className="space-y-6">
            <div className="space-y-2 text-center">
              <h2 className="text-xl font-bold">connect an AI provider</h2>
              <p className="text-sm text-muted-foreground">
                add a connection for transcription and refinement. you can add
                more later.
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
                    onClick={() => {
                      setValue("providerId", p.id);
                      setTestOk(false);
                    }}
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
                        <p className="text-xs text-muted-foreground">
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

            {/* Connection fields — adapts to provider type */}
            <div className="space-y-3">
              {/* Base URL — shown for providers that need it */}
              {needsBaseUrl && (
                <FormField
                  label={<>Base URL <InfoTip text="The API endpoint, e.g. http://localhost:11434/v1 for Ollama or https://api.openai.com/v1 for OpenAI." /></>}
                  error={errors.baseUrl?.message}
                >
                  <Input
                    {...register("baseUrl", { onChange: () => setTestOk(false) })}
                    placeholder="http://localhost:11434/v1"
                  />
                </FormField>
              )}

              {/* API key */}
              <FormField
                label={
                  <span className="flex items-center gap-2">
                    API Key
                    {needsBaseUrl && (
                      <span className="text-xs text-muted-foreground font-normal">
                        optional for local
                      </span>
                    )}
                    {preset?.keyUrl && (
                      <ExternalLink href={preset.keyUrl} className="text-xs font-normal">
                        {preset.keyUrlLabel}
                      </ExternalLink>
                    )}
                  </span>
                }
                error={errors.apiKey?.message}
              >
                <div className="flex">
                  <Input
                    type={showKey ? "text" : "password"}
                    {...register("apiKey", { onChange: () => setTestOk(false) })}
                    placeholder={preset?.keyPlaceholder ?? "API key..."}
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
              </FormField>

              {/* Test connection */}
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                disabled={testing || !canProceed}
                onClick={testConnection}
              >
                {testing ? (
                  <>
                    <Loader2 className="size-3.5 animate-spin" /> Testing...
                  </>
                ) : testOk ? (
                  <>
                    <Check className="size-3.5" /> Connected
                  </>
                ) : (
                  <>
                    <RefreshCw className="size-3.5" /> Test Connection
                  </>
                )}
              </Button>
            </div>

            <div className="flex gap-2">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => setStep(2)}
              >
                <SkipForward className="size-4" /> skip for now
              </Button>
              <Button
                className="flex-1"
                onClick={advanceStep1}
                disabled={!canProceed}
              >
                next <ArrowRight className="size-4" />
              </Button>
            </div>
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
            <Button className="w-full" onClick={() => setStep(3)}>
              next <ArrowRight className="size-4" />
            </Button>
          </div>
        )}

        {/* Step 3: Shortcuts */}
        {step === 3 && (
          <div className="space-y-6">
            <div className="space-y-2 text-center">
              <h2 className="text-xl font-bold">set up a shortcut</h2>
              <p className="text-sm text-muted-foreground">
                bind a key to toggle recording from anywhere, without opening
                the app window.
              </p>
            </div>

            <div className="rounded-lg border bg-card p-4">
              <ShortcutGuide platform={platformInfo} compact />
            </div>

            <Button
              className="w-full"
              onClick={canProceed ? finishWithConnection : finishSkip}
            >
              finish <Check className="size-4" />
            </Button>
          </div>
        )}

        {/* Step indicators */}
        <div className="flex justify-center gap-2">
          {[0, 1, 2, 3].map((i) => (
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
