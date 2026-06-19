import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { BookOpen, Loader2, Plus } from "lucide-react";
import { api, ApiError, optionalAuthApi } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { SelectField } from "@/components/select-field";
import { useAuthStore } from "@/lib/store";
import { useI18n } from "@/lib/i18n";
import { canManageBooksInSpace, type SpaceRole } from "@/lib/roles";

interface Candidate {
  id: string;
  title: string;
  description: string;
  root_path: string;
  doc_count: number;
  chapter_count: number;
}

interface SavedBook {
  id: string;
  slug: string;
  title: string;
  status: string;
  root_path: string;
  chapter_count: number;
  generated_at?: string;
}

const AUDIENCE_IDS = ["developer", "architect", "ops", "onboarding"] as const;

export function SpaceBooksPage() {
  const { slug } = useParams<{ slug: string }>();
  const qc = useQueryClient();
  const { t, statusLabel } = useI18n();
  const user = useAuthStore((s) => s.user);

  const { data: spaceMeta } = useQuery({
    queryKey: ["space", slug],
    queryFn: () =>
      optionalAuthApi<{ my_role?: SpaceRole; can_edit?: boolean }>(`/api/spaces/${slug}`),
    enabled: !!slug,
  });
  const canManage = canManageBooksInSpace(spaceMeta?.my_role, user, spaceMeta?.can_edit);
  const [rootPath, setRootPath] = useState("analytics");
  const [audience, setAudience] = useState<(typeof AUDIENCE_IDS)[number]>("developer");
  const [focus, setFocus] = useState("");
  const [error, setError] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["books", slug],
    queryFn: () =>
      optionalAuthApi<{
        saved: SavedBook[];
        candidates: Candidate[];
        llm_available: boolean;
      }>(`/api/spaces/${slug}/books`),
    enabled: !!slug,
  });

  const createBook = useMutation({
    mutationFn: () =>
      api(`/api/spaces/${slug}/books`, {
        method: "POST",
        body: JSON.stringify({ root_path: rootPath, audience, focus: focus.trim() || undefined }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["books", slug] });
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("books.createFailed")),
  });

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  const saved = data?.saved ?? [];
  const candidates = data?.candidates ?? [];

  return (
    <FadeIn>
      <div className="glass p-6">
        <h2 className="text-lg font-semibold text-fg">{t("books.title")}</h2>
        <p className="mt-2 text-sm text-muted">
          {canManage ? t("books.description") : t("books.readOnlyDescription")}
        </p>
        {canManage && (
          data?.llm_available ? (
            <p className="mt-3 text-sm text-success-soft">{t("books.llmConnected")}</p>
          ) : (
            <p className="mt-3 text-sm text-subtle">{t("books.noLlm")}</p>
          )
        )}
      </div>

      {canManage && (
        <div className="glass mt-4 p-6">
          <h3 className="text-sm font-semibold text-fg">{t("books.createTitle")}</h3>
          <div className="mt-4 grid gap-3 sm:grid-cols-2">
            <SelectField value={rootPath} onChange={(e) => setRootPath(e.target.value)}>
              {candidates.map((c) => (
                <option key={c.id} value={c.root_path}>
                  {c.title} ({t("books.pagesCount", { count: c.doc_count })})
                </option>
              ))}
            </SelectField>
            <SelectField
              value={audience}
              onChange={(e) => setAudience(e.target.value as (typeof AUDIENCE_IDS)[number])}
            >
              {AUDIENCE_IDS.map((id) => (
                <option key={id} value={id}>
                  {t(`books.audiences.${id}`)}
                </option>
              ))}
            </SelectField>
            <input
              className="input-field sm:col-span-2"
              placeholder={t("books.focusPlaceholder")}
              value={focus}
              onChange={(e) => setFocus(e.target.value)}
            />
          </div>
          {error && <p className="mt-3 text-sm text-danger-soft">{error}</p>}
          <button
            type="button"
            className="btn-primary mt-4"
            disabled={createBook.isPending || !rootPath}
            onClick={() => createBook.mutate()}
          >
            {createBook.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
            {t("books.create")}
          </button>
        </div>
      )}

      <div className="mt-4 space-y-2">
        {saved.map((book) => (
          <Link
            key={book.id}
            to={`/spaces/${slug}/books/${book.slug}`}
            className="glass block p-4 transition-colors hover:bg-surface-muted"
          >
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="font-medium text-fg">{book.title}</p>
                <p className="text-xs text-subtle">
                  {book.root_path}/ · {book.chapter_count} {t("common.chapters")}
                </p>
              </div>
              <span className="badge badge-neutral">{statusLabel(book.status)}</span>
            </div>
          </Link>
        ))}
        {saved.length === 0 && (
          <div className="glass flex flex-col items-center px-6 py-12 text-center">
            <BookOpen className="h-10 w-10 text-primary/50" />
            <p className="mt-4 text-sm text-muted">
              {canManage ? t("books.noSaved") : t("books.noSavedReader")}
            </p>
          </div>
        )}
      </div>
    </FadeIn>
  );
}
