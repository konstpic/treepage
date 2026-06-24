import { useCallback, useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";

export interface MentionUser {
  id: string;
  email: string;
  display_name: string;
}

interface MentionTextareaProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  minRows?: number;
}

export function MentionTextarea({
  value,
  onChange,
  placeholder,
  className,
  minRows = 4,
}: MentionTextareaProps) {
  const { t } = useI18n();
  const ref = useRef<HTMLTextAreaElement>(null);
  const [mentionQuery, setMentionQuery] = useState<string | null>(null);
  const [mentionStart, setMentionStart] = useState(-1);
  const [highlight, setHighlight] = useState(0);

  const { data: suggestions } = useQuery({
    queryKey: ["mention-users", mentionQuery],
    queryFn: () =>
      api<{ items: MentionUser[] }>(
        `/api/users/mention-suggest?q=${encodeURIComponent(mentionQuery ?? "")}`,
      ),
    enabled: mentionQuery !== null,
    staleTime: 30_000,
  });

  const items = suggestions?.items ?? [];

  const detectMention = useCallback((text: string, cursor: number) => {
    const before = text.slice(0, cursor);
    const at = before.lastIndexOf("@");
    if (at < 0) {
      setMentionQuery(null);
      setMentionStart(-1);
      return;
    }
    const fragment = before.slice(at + 1);
    if (/\s/.test(fragment)) {
      setMentionQuery(null);
      setMentionStart(-1);
      return;
    }
    setMentionStart(at);
    setMentionQuery(fragment);
    setHighlight(0);
  }, []);

  function insertMention(user: MentionUser) {
    if (mentionStart < 0 || !ref.current) return;
    const cursor = ref.current.selectionStart;
    const before = value.slice(0, mentionStart);
    const after = value.slice(cursor);
    const mention = `@${user.email} `;
    const next = before + mention + after;
    onChange(next);
    setMentionQuery(null);
    setMentionStart(-1);
    requestAnimationFrame(() => {
      const pos = before.length + mention.length;
      ref.current?.setSelectionRange(pos, pos);
      ref.current?.focus();
    });
  }

  useEffect(() => {
    if (highlight >= items.length) setHighlight(0);
  }, [items.length, highlight]);

  return (
    <div className="relative">
      <textarea
        ref={ref}
        className={cn("input-field w-full resize-none", className)}
        style={{ minHeight: `${minRows * 1.5}rem` }}
        placeholder={placeholder}
        value={value}
        onChange={(e) => {
          onChange(e.target.value);
          detectMention(e.target.value, e.target.selectionStart);
        }}
        onClick={(e) => detectMention(value, e.currentTarget.selectionStart)}
        onKeyUp={(e) => detectMention(value, e.currentTarget.selectionStart)}
        onKeyDown={(e) => {
          if (mentionQuery === null || items.length === 0) return;
          if (e.key === "ArrowDown") {
            e.preventDefault();
            setHighlight((h) => (h + 1) % items.length);
          } else if (e.key === "ArrowUp") {
            e.preventDefault();
            setHighlight((h) => (h - 1 + items.length) % items.length);
          } else if (e.key === "Enter" || e.key === "Tab") {
            e.preventDefault();
            insertMention(items[highlight]);
          } else if (e.key === "Escape") {
            setMentionQuery(null);
          }
        }}
      />
      {mentionQuery !== null && (
        <ul
          className="absolute z-50 mt-1 max-h-48 w-full overflow-auto rounded-xl border border-default bg-surface py-1 shadow-lg"
          role="listbox"
        >
          {items.length === 0 ? (
            <li className="px-3 py-2 text-xs text-muted">{t("comments.mentionNoUsers")}</li>
          ) : (
            items.map((user, i) => (
              <li key={user.id}>
                <button
                  type="button"
                  role="option"
                  aria-selected={i === highlight}
                  className={cn(
                    "flex w-full flex-col px-3 py-2 text-left text-sm hover:bg-surface-muted",
                    i === highlight && "bg-surface-muted",
                  )}
                  onMouseDown={(e) => {
                    e.preventDefault();
                    insertMention(user);
                  }}
                >
                  <span className="font-medium text-fg">{user.display_name || user.email}</span>
                  <span className="text-xs text-subtle">{user.email}</span>
                </button>
              </li>
            ))
          )}
        </ul>
      )}
    </div>
  );
}
