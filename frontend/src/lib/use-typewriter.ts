import { useEffect, useState } from "react";

const TYPE_MS = 58;
const PAUSE_MS = 2200;
const RESTART_MS = 350;

export function useTypewriterText(text: string, enabled: boolean) {
  const [value, setValue] = useState("");

  useEffect(() => {
    if (!enabled || !text) {
      setValue("");
      return;
    }

    let cancelled = false;
    let index = 0;
    let timeout: ReturnType<typeof setTimeout>;

    const tick = () => {
      if (cancelled) return;

      if (index < text.length) {
        index += 1;
        setValue(text.slice(0, index));
        timeout = setTimeout(tick, TYPE_MS);
        return;
      }

      timeout = setTimeout(() => {
        if (cancelled) return;
        index = 0;
        setValue("");
        timeout = setTimeout(tick, RESTART_MS);
      }, PAUSE_MS);
    };

    setValue("");
    index = 0;
    timeout = setTimeout(tick, RESTART_MS);

    return () => {
      cancelled = true;
      clearTimeout(timeout);
    };
  }, [text, enabled]);

  return value;
}
