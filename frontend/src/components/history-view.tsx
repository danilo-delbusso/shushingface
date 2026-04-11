import { History, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { ConfirmDialog } from "@/components/confirm-dialog";
import type { history } from "../../wailsjs/go/models";

interface HistoryViewProps {
  items: history.Record[];
  onClear: () => void;
}

export function HistoryView({ items, onClear }: HistoryViewProps) {
  return (
    <div className="flex flex-1 flex-col gap-4 overflow-hidden p-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">History</h2>
        {items.length > 0 && (
          <ConfirmDialog
            trigger={
              <Button variant="destructive" size="sm">
                <Trash2 className="size-3.5" />
                Clear All
              </Button>
            }
            title="Clear all history?"
            description="This will permanently delete all transcription records. This action cannot be undone."
            confirmLabel="Delete All"
            onConfirm={onClear}
          />
        )}
      </div>

      {items.length === 0 ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 text-muted-foreground">
          <History className="size-10 opacity-30" />
          <p className="text-sm">No transcriptions yet.</p>
        </div>
      ) : (
        <div className="flex-1 overflow-y-auto">
          <div className="space-y-2 pr-2">
            {items.map((item) => (
              <Card key={item.id}>
                <CardContent className="space-y-1 px-3 py-2">
                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span>{new Date(item.timestamp).toLocaleString()}</span>
                    {item.activeApp && (
                      <span className="font-medium text-primary">
                        {item.activeApp}
                      </span>
                    )}
                  </div>
                  <p className="text-sm font-medium">{item.refinedMessage}</p>
                  <p className="text-xs text-muted-foreground italic">
                    &ldquo;{item.rawTranscript}&rdquo;
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
