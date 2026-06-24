import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/** Shared max-width shell aligned with the main navigation. */
export const pageShellClass =
  "mx-auto w-full max-w-[min(100%,100rem)] px-4 sm:px-6 lg:px-8 2xl:px-10";

export function formatDate(iso: string | null | undefined): string {
  if (!iso) return "—";
  return new Date(iso).toLocaleDateString("en-US", {
    day: "numeric",
    month: "long",
    year: "numeric",
  });
}
