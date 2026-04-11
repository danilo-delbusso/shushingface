import { Sun, Moon, Monitor } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { config } from "../../wailsjs/go/models";

interface AppearanceViewProps {
  settings: config.Settings;
  onSave: (settings: config.Settings) => void;
}

export function AppearanceView({ settings, onSave }: AppearanceViewProps) {
  const setTheme = (theme: string) => {
    onSave({ ...settings, theme } as config.Settings);
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Theme</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              {([
                { value: "light", icon: Sun, label: "Light" },
                { value: "dark", icon: Moon, label: "Dark" },
                { value: "system", icon: Monitor, label: "System" },
              ] as const).map(({ value, icon: Icon, label }) => (
                <Button
                  key={value}
                  variant={settings.theme === value ? "default" : "outline"}
                  size="sm"
                  className="flex-1"
                  onClick={() => setTheme(value)}
                >
                  <Icon className="size-3.5" />
                  {label}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
