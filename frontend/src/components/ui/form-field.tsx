import { Label } from "@/components/ui/label";

interface FormFieldProps {
  label: React.ReactNode;
  error?: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}

export function FormField({
  label,
  error,
  description,
  children,
  className,
}: FormFieldProps) {
  return (
    <div className={className ?? "space-y-1"}>
      <Label>{label}</Label>
      {children}
      {error && <p className="text-xs text-destructive">{error}</p>}
      {description && (
        <p className="text-xs text-muted-foreground">{description}</p>
      )}
    </div>
  );
}
