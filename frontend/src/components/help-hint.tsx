import { HelpCircle } from "lucide-react";
import { cn } from "@/lib/utils";

type HelpHintProps = {
  text: string;
  className?: string;
};

/** Question-mark tooltip for settings cards and form labels. */
export function HelpHint({ text, className }: HelpHintProps) {
  return (
    <span className={cn("help-hint group relative inline-flex", className)}>
      <button
        type="button"
        className="inline-flex h-5 w-5 items-center justify-center rounded-full text-subtle transition-colors hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40"
        aria-label={text}
      >
        <HelpCircle className="h-4 w-4" aria-hidden />
      </button>
      <span
        role="tooltip"
        className="help-hint__popup pointer-events-none absolute left-1/2 top-full z-50 mt-2 w-64 -translate-x-1/2 rounded-xl border border-default bg-surface px-3 py-2 text-left text-xs leading-relaxed text-fg-secondary opacity-0 shadow-lg transition-opacity group-hover:opacity-100 group-focus-within:opacity-100"
      >
        {text}
      </span>
    </span>
  );
}

type SettingsCardProps = {
  title: string;
  hint?: string;
  help?: string;
  icon?: React.ReactNode;
  children: React.ReactNode;
};

export function SettingsCard({ title, hint, help, icon, children }: SettingsCardProps) {
  return (
    <div className="glass p-6">
      <div className="flex items-center gap-2">
        {icon}
        <h2 className="text-lg font-semibold text-fg">{title}</h2>
        {help ? <HelpHint text={help} /> : null}
      </div>
      {hint ? <p className="mt-1 text-sm text-muted">{hint}</p> : null}
      {children}
    </div>
  );
}
