import type { SelectHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

export function SelectField({
  className,
  children,
  ...props
}: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select className={cn("select-field", className)} {...props}>
      {children}
    </select>
  );
}
