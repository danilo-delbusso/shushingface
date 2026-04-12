import { Component } from "react";
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";

interface Props {
  children: React.ReactNode;
}

interface State {
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex flex-1 flex-col items-center justify-center gap-4 p-6 text-center">
          <AlertTriangle className="size-10 text-destructive opacity-50" />
          <div className="space-y-1">
            <h2 className="text-sm font-semibold">Something went wrong</h2>
            <p className="max-w-sm text-xs text-muted-foreground">
              {this.state.error.message}
            </p>
          </div>
          <Button
            size="sm"
            variant="outline"
            onClick={() => this.setState({ error: null })}
          >
            Try again
          </Button>
        </div>
      );
    }
    return this.props.children;
  }
}
