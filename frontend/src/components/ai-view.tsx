import { useState, useEffect } from "react";
import { Play, RotateCcw, Loader2, Key, Bot, Eye, EyeOff, AlertTriangle } from "lucide-react";
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
import * as AppBridge from "../../wailsjs/go/desktop/App";
import type { config } from "../../wailsjs/go/models";

interface AiViewProps {
  settings: config.Settings;
  configured: boolean;
  onSave: (settings: config.Settings) => void;
}

export function AiView({ settings, configured, onSave }: AiViewProps) {
  const [showKey, setShowKey] = useState(false);
  const [defaultPrompt, setDefaultPrompt] = useState("");
  const [draft, setDraft] = useState({
    apiKey: settings.providers?.[settings.transcriptionProviderId]?.apiKey ?? "",
    transcriptionModel: settings.transcriptionModel,
    refinementModel: settings.refinementModel,
    systemPrompt: "",
  });
  const [sampleText, setSampleText] = useState("");
  const [testResult, setTestResult] = useState("");
  const [testing, setTesting] = useState(false);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    AppBridge.GetDefaultPrompt().then((p) => {
      setDefaultPrompt(p);
      setDraft((d) => ({ ...d, systemPrompt: settings.systemPrompt || p }));
      setLoaded(true);
    });
  }, [settings.systemPrompt]);

  const handleSave = () => {
    const providerId = settings.transcriptionProviderId;
    onSave({
      ...settings,
      providers: {
        ...settings.providers,
        [providerId]: { ...settings.providers[providerId], apiKey: draft.apiKey },
      },
      transcriptionModel: draft.transcriptionModel,
      refinementModel: draft.refinementModel,
      systemPrompt: draft.systemPrompt,
    } as config.Settings);
  };

  const placeholderText =
    "hey um so I was thinking we should probably move the meeting to thursday because like john cant make it on wednesday and I think it would be better if we all met together you know";

  const handleTest = async () => {
    const text = sampleText.trim() || placeholderText;
    setTesting(true);
    setTestResult("");
    try {
      const res = await AppBridge.TestPrompt(text, draft.systemPrompt);
      if (res.error) {
        toast.error(res.error);
      } else {
        setTestResult(res.refined);
      }
    } catch (err) {
      toast.error(`Test failed: ${err}`);
    } finally {
      setTesting(false);
    }
  };

  const currentPrompt = settings.systemPrompt || defaultPrompt;
  const hasChanges =
    loaded &&
    (draft.apiKey !== (settings.providers?.[settings.transcriptionProviderId]?.apiKey ?? "") ||
      draft.transcriptionModel !== settings.transcriptionModel ||
      draft.refinementModel !== settings.refinementModel ||
      draft.systemPrompt !== currentPrompt);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        {!configured && (
          <div className="flex items-center gap-3 rounded-lg border border-amber-600/30 bg-amber-600/10 p-3 text-sm text-amber-500">
            <AlertTriangle className="size-4 shrink-0" />
            Add your API key to get started.
          </div>
        )}

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
                value={draft.apiKey}
                placeholder="gsk_..."
                onChange={(e) => setDraft({ ...draft, apiKey: e.target.value })}
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
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Bot className="size-4" /> Models
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1">
              <Label htmlFor="trans-model">Transcription</Label>
              <Input
                id="trans-model"
                value={draft.transcriptionModel}
                onChange={(e) => setDraft({ ...draft, transcriptionModel: e.target.value })}
              />
            </div>
            <Separator />
            <div className="space-y-1">
              <Label htmlFor="refine-model">Refinement</Label>
              <Input
                id="refine-model"
                value={draft.refinementModel}
                onChange={(e) => setDraft({ ...draft, refinementModel: e.target.value })}
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Refinement Prompt</CardTitle>
            <CardDescription>
              Instructs the AI how to rewrite your transcripts.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <textarea
              value={draft.systemPrompt}
              onChange={(e) => setDraft({ ...draft, systemPrompt: e.target.value })}
              rows={8}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
            />
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setDraft({ ...draft, systemPrompt: defaultPrompt })}
            >
              <RotateCcw className="size-3.5" /> Reset to default
            </Button>
          </CardContent>
        </Card>

        <Button className="w-full" onClick={handleSave} disabled={!hasChanges}>
          {hasChanges ? "Save" : "No changes"}
        </Button>

        <Separator />

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Test Playground</CardTitle>
            <CardDescription>
              Try your prompt against sample text.
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
