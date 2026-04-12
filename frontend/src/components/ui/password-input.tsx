import { useState, forwardRef } from "react";
import { Eye, EyeOff } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const PasswordInput = forwardRef<
  HTMLInputElement,
  React.ComponentProps<"input">
>(({ className, ...props }, ref) => {
  const [show, setShow] = useState(false);

  return (
    <div className="flex">
      <Input
        ref={ref}
        type={show ? "text" : "password"}
        className={cn("rounded-r-none", className)}
        {...props}
      />
      <Button
        type="button"
        variant="outline"
        size="icon"
        onClick={() => setShow(!show)}
        className="rounded-l-none border-l-0"
      >
        {show ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
      </Button>
    </div>
  );
});
PasswordInput.displayName = "PasswordInput";

export { PasswordInput };
