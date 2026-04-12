import { useState, useEffect, useCallback } from "react";
import {
  Info,
  FileText,
  RefreshCw,
  Copy,
  Shield,
  ShieldOff,
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
import { ExternalLink } from "@/components/ui/external-link";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import type { platform } from "../../wailsjs/go/models";

interface AboutViewProps {
  version: string;
  platform: platform.Info | null;
}

export function AboutView({ version, platform }: AboutViewProps) {
  const [logPath, setLogPath] = useState("");
  const [logs, setLogs] = useState("");
  const [loading, setLoading] = useState(false);
  const [secure, setSecure] = useState(false);

  useEffect(() => {
    AppBridge.GetLogPath().then(setLogPath);
    AppBridge.IsSecretStorageSecure().then(setSecure);
  }, []);

  const loadLogs = useCallback(async () => {
    setLoading(true);
    try {
      const text = await AppBridge.GetRecentLogs(100);
      setLogs(text);
    } catch (err) {
      toast.error(`Failed to load logs: ${err}`);
    } finally {
      setLoading(false);
    }
  }, []);

  const copyLogs = () => {
    navigator.clipboard.writeText(logs);
    toast.success("Logs copied to clipboard");
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        {/* App info */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Info className="size-4" /> About
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Version</span>
                <span>{version || "dev"}</span>
              </div>
              {platform && (
                <>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">OS</span>
                    <span>{platform.os}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Desktop</span>
                    <span>{platform.desktop || "unknown"}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Display</span>
                    <span>{platform.displayServer}</span>
                  </div>
                </>
              )}
              <div className="flex justify-between">
                <span className="text-muted-foreground">Secret storage</span>
                <span className="flex items-center gap-1">
                  {secure ? (
                    <>
                      <Shield className="size-3 text-green-500" /> OS keyring
                    </>
                  ) : (
                    <>
                      <ShieldOff className="size-3 text-amber-500" /> Config
                      file
                    </>
                  )}
                </span>
              </div>
            </div>
            <div className="pt-2">
              <ExternalLink
                href="https://codeberg.org/dbus/shushingface"
                className="text-xs"
              >
                Source code on Codeberg
              </ExternalLink>
            </div>
          </CardContent>
        </Card>

        {/* Logs */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              <FileText className="size-4" /> Logs
            </CardTitle>
            <CardDescription>
              {logPath && (
                <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
                  {logPath}
                </code>
              )}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={loadLogs}
                disabled={loading}
              >
                {loading ? (
                  <>
                    <RefreshCw className="size-3.5 animate-spin" /> Loading...
                  </>
                ) : (
                  <>
                    <RefreshCw className="size-3.5" />{" "}
                    {logs ? "Refresh" : "Load Logs"}
                  </>
                )}
              </Button>
              {logs && (
                <Button variant="outline" size="sm" onClick={copyLogs}>
                  <Copy className="size-3.5" /> Copy
                </Button>
              )}
            </div>
            {logs && (
              <pre className="max-h-80 overflow-auto rounded-md border bg-muted/30 p-3 text-[11px] leading-relaxed">
                {logs}
              </pre>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
