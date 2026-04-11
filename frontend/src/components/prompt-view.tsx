import { useState } from "react";
import { Play, RotateCcw, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ScrollArea } from "@/components/ui/scroll-area";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import type { config } from "../../wailsjs/go/models";

interface PromptViewProps {
  settings: config.Settings;
  onSave: (settings: config.Settings) => void;
}

export function PromptView({ settings, onSave }: PromptViewProps) {
  const [draft, setDraft] = useState(settings.systemPrompt || "");
  const [sampleText, setSampleText] = useState("");
  const [testResult, setTestResult] = useState("");
  const [testing, setTesting] = useState(false);

  const handleSave = () => {
    const updated = { ...settings, systemPrompt: draft };
    // Use the config class to ensure proper serialization
    onSave(updated as config.Settings);
  };

  const handleReset = async () => {
    const defaultPrompt = await AppBridge.GetDefaultPrompt();
    setDraft(defaultPrompt);
  };

  const handleTest = async () => {
    if (!sampleText.trim()) {
      toast.error("Enter some sample text to test");
      return;
    }
    setTesting(true);
    setTestResult("");
    try {
      const res = await AppBridge.TestPrompt(sampleText, draft);
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

  const hasChanges = draft !== (settings.systemPrompt || "");

  return (
    <ScrollArea className="flex-1">
      <div className="space-y-4 p-6 max-w-2xl">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Refinement Prompt</CardTitle>
            <CardDescription>
              This prompt instructs the AI how to rewrite your speech
              transcripts. Customize it for your use case — formal emails,
              casual messages, technical docs, etc.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="system-prompt">System Prompt</Label>
              <textarea
                id="system-prompt"
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                rows={10}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                placeholder="Enter your system prompt..."
              />
            </div>
            <div className="flex items-center justify-between">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleReset}
              >
                <RotateCcw className="size-3.5" />
                Reset to Default
              </Button>
              <Button
                type="button"
                size="sm"
                onClick={handleSave}
                disabled={!hasChanges}
              >
                {hasChanges ? "Save Prompt" : "No Changes"}
              </Button>
            </div>
          </CardContent>
        </Card>

        <Separator />

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Test Playground</CardTitle>
            <CardDescription>
              Enter sample text to see how your prompt transforms it. Uses the
              prompt above (saved or unsaved).
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="sample-text">Sample Transcript</Label>
              <textarea
                id="sample-text"
                value={sampleText}
                onChange={(e) => setSampleText(e.target.value)}
                rows={4}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                placeholder="hey um so I was thinking we should probably move the meeting to thursday because like john cant make it on wednesday..."
              />
            </div>
            <Button
              type="button"
              onClick={handleTest}
              disabled={testing || !sampleText.trim()}
              className="w-full"
            >
              {testing ? (
                <>
                  <Loader2 className="size-4 animate-spin" />
                  Running...
                </>
              ) : (
                <>
                  <Play className="size-4" />
                  Run Test
                </>
              )}
            </Button>
            {testResult && (
              <Card className="bg-muted/50">
                <CardContent className="p-4">
                  <Label className="text-xs text-muted-foreground">
                    Result
                  </Label>
                  <p className="mt-1 text-sm leading-relaxed whitespace-pre-wrap">
                    {testResult}
                  </p>
                </CardContent>
              </Card>
            )}
          </CardContent>
        </Card>
      </div>
    </ScrollArea>
  );
}
