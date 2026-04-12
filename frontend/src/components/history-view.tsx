import { History, Trash2, ChevronDown, ChevronUp, AlertTriangle } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
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
              <HistoryItem key={item.id} item={item} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function HistoryItem({ item }: { item: history.Record }) {
  const [open, setOpen] = useState(false);
  const isError = !!item.error;

  return (
    <div
      className={`space-y-0.5 rounded-lg border text-card-foreground px-3 py-2 ${
        isError ? "border-destructive/30 bg-destructive/5" : "bg-card"
      }`}
    >
      <div className="flex items-center justify-between text-[11px] text-muted-foreground">
        <span className="flex items-center gap-1">
          {isError && <AlertTriangle className="size-3 text-destructive shrink-0" />}
          {new Date(item.timestamp).toLocaleString()}
        </span>
        {item.activeApp && (
          <span className="font-medium text-primary">{item.activeApp}</span>
        )}
      </div>
      {isError ? (
        <p className="text-xs text-destructive">{item.error}</p>
      ) : (
        <>
          <p className="text-sm">{item.refinedMessage}</p>
          <button
            type="button"
            onClick={() => setOpen(!open)}
            className="flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground transition-colors"
          >
            {open ? (
              <ChevronUp className="size-3" />
            ) : (
              <ChevronDown className="size-3" />
            )}
            transcript
          </button>
          {open && (
            <p className="text-xs text-muted-foreground italic">
              &ldquo;{item.rawTranscript}&rdquo;
            </p>
          )}
        </>
      )}
    </div>
  );
}
