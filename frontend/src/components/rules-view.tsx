import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { RotateCcw } from "lucide-react";
import { toast } from "sonner";
import { globalRulesSchema, type GlobalRulesFormData } from "@/lib/schemas";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { AdvancedToggle } from "@/components/ui/advanced-toggle";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { InfoTip } from "@/components/info-tip";
import { cn, textareaClass, textareaCompactClass } from "@/lib/utils";
import * as AppBridge from "../../wailsjs/go/desktop/App";
import { config } from "../../wailsjs/go/models";

interface RulesViewProps {
  settings: config.Settings;
  onSave: (settings: config.Settings) => void;
}

export function RulesView({ settings, onSave }: RulesViewProps) {
  const form = useForm<GlobalRulesFormData>({
    resolver: zodResolver(globalRulesSchema),
    defaultValues: {
      globalRules: settings.globalRules ?? "",
      builtInRules: settings.builtInRules ?? "",
    },
  });
  const [advancedOpen, setAdvancedOpen] = useState(false);

  useEffect(() => {
    if (!settings.builtInRules) {
      AppBridge.GetDefaultBuiltInRules().then((rules) =>
        form.setValue("builtInRules", rules),
      );
    }
  }, [settings.builtInRules, form]);

  const submit = (data: GlobalRulesFormData) => {
    onSave(
      config.Settings.createFrom({
        ...settings,
        globalRules: data.globalRules,
        builtInRules: data.builtInRules || undefined,
      }),
    );
    form.reset(data);
  };

  const restoreBuiltIn = async () => {
    const rules = await AppBridge.GetDefaultBuiltInRules();
    form.setValue("builtInRules", rules, { shouldDirty: true });
    toast.success("Built-in rules restored to default");
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="space-y-4 p-6 max-w-2xl">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-sm">
              Global Rules{" "}
              <InfoTip text="Rules applied to every refinement style. Use for preferences like 'don't use em dashes' or 'use British English'." />
            </CardTitle>
            <CardDescription>
              Applied to all styles, after the style prompt.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <textarea
              {...form.register("globalRules")}
              rows={4}
              className={textareaClass}
              placeholder={
                "- Don't use em dashes\n- Use British English spelling\n- Keep sentences under 20 words"
              }
            />

            <AdvancedToggle
              label="Built-in rules"
              open={advancedOpen}
              onToggle={setAdvancedOpen}
            >
              <p className="text-xs text-muted-foreground">
                Core rules always applied. Edit with care.
              </p>
              <textarea
                {...form.register("builtInRules")}
                rows={6}
                className={cn(textareaCompactClass, "font-mono")}
                placeholder="Built-in rules (leave empty for defaults)"
              />
              <ConfirmDialog
                trigger={
                  <Button variant="outline" size="sm" className="text-xs">
                    <RotateCcw className="size-3" /> Restore Default Rules
                  </Button>
                }
                title="Restore built-in rules?"
                description="This will reset the built-in rules to their factory defaults."
                confirmLabel="Restore"
                onConfirm={restoreBuiltIn}
              />
            </AdvancedToggle>

            <Button
              size="sm"
              onClick={form.handleSubmit(submit)}
              disabled={!form.formState.isDirty}
            >
              Save
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
