import { useState } from "react";
import { Coffee, Briefcase, Zap, Eye, EyeOff, ArrowRight, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { config } from "../../wailsjs/go/models";

const icons: Record<string, React.FC<{ className?: string }>> = {
  coffee: Coffee,
  briefcase: Briefcase,
  zap: Zap,
};

interface WelcomeWizardProps {
  settings: config.Settings;
  onComplete: (settings: config.Settings) => void;
}

export function WelcomeWizard({ settings, onComplete }: WelcomeWizardProps) {
  const [step, setStep] = useState(0);
  const [apiKey, setApiKey] = useState(settings.providerApiKey ?? "");
  const [showKey, setShowKey] = useState(false);
  const [selectedProfile, setSelectedProfile] = useState("professional");

  const finish = () => {
    onComplete(
      config.Settings.createFrom({
        ...settings,
        providerApiKey: apiKey,
        activeProfileId: selectedProfile,
        setupComplete: true,
      }),
    );
  };

  return (
    <div className="flex h-screen w-screen items-center justify-center bg-background">
      <div className="w-full max-w-md space-y-8 px-6">
        {step === 0 && (
          <div className="flex flex-col items-center gap-6 text-center">
            <img src="/appicon.png" alt="" className="size-20" />
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

        {step === 1 && (
          <div className="flex flex-col gap-6">
            <div className="space-y-2 text-center">
              <h2 className="text-xl font-bold">connect to groq</h2>
              <p className="text-sm text-muted-foreground">
                shushing face uses groq for fast transcription.{" "}
                <button
                  type="button"
                  className="text-primary underline underline-offset-2"
                  onClick={() =>
                    window.open("https://console.groq.com/keys", "_blank")
                  }
                >
                  get a free key
                </button>
              </p>
            </div>
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
                {showKey ? (
                  <EyeOff className="size-4" />
                ) : (
                  <Eye className="size-4" />
                )}
              </Button>
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

        {step === 2 && (
          <div className="flex flex-col gap-6">
            <div className="space-y-2 text-center">
              <h2 className="text-xl font-bold">choose your style</h2>
              <p className="text-sm text-muted-foreground">
                how should your speech be refined? you can change this anytime.
              </p>
            </div>
            <div className="grid gap-3">
              {settings.refinementProfiles?.map((profile) => {
                const Icon = icons[profile.icon] || Zap;
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
                        active ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground"
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

        <div className="flex justify-center gap-2">
          {[0, 1, 2].map((i) => (
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
