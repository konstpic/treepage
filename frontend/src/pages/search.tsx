import { useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { Loader2, Search, Sparkles, ThumbsDown, ThumbsUp } from "lucide-react";
import { ApiError, api, optionalAuthApi } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { SelectField } from "@/components/select-field";
import { useI18n } from "@/lib/i18n";
import { useTypewriterText } from "@/lib/use-typewriter";
import { useAuthStore } from "@/lib/store";
import { cn, pageShellClass } from "@/lib/utils";

interface SearchResult {
  id: string;
  space_id: string;
  space_slug: string;
  title: string;
  slug: string;
  snippet: string;
  author_name?: string;
  tags?: string[];
}

interface SpaceOption {
  id: string;
  slug: string;
  name: string;
}

function buildSearchUrl(params: {
  q: string;
  space_slug?: string;
  author?: string;
  tags?: string;
}) {
  const sp = new URLSearchParams();
  if (params.q) sp.set("q", params.q);
  if (params.space_slug) sp.set("space_slug", params.space_slug);
  if (params.author) sp.set("author", params.author);
  if (params.tags) sp.set("tags", params.tags);
  const qs = sp.toString();
  return qs ? `/api/search?${qs}` : "/api/search";
}

interface RagSource {
  document_id: string;
  space_slug: string;
  doc_slug: string;
  title: string;
  snippet: string;
  score?: number;
}

interface RagCitation {
  document_id: string;
  space_slug: string;
  doc_slug: string;
  title: string;
  path: string;
  quote: string;
}

interface RagAnswer {
  answer: string;
  sources: RagSource[];
  citations?: RagCitation[];
  confidence?: number;
  low_confidence?: boolean;
  follow_up_questions?: string[];
}

export function SearchPage() {
  const { t } = useI18n();
  const { isAuthenticated } = useAuthStore();
  const [mode, setMode] = useState<"search" | "ask">("search");
  const [askQuestion, setAskQuestion] = useState("");
  const [askError, setAskError] = useState("");
  const [searchParams, setSearchParams] = useSearchParams();
  const initialQuery = searchParams.get("q") ?? "";
  const [query, setQuery] = useState(initialQuery);
  const [submitted, setSubmitted] = useState(initialQuery.trim());
  const [spaceSlug, setSpaceSlug] = useState(searchParams.get("space_slug") ?? "");
  const [author, setAuthor] = useState(searchParams.get("author") ?? "");
  const [tags, setTags] = useState(searchParams.get("tags") ?? "");
  const [focused, setFocused] = useState(false);

  const placeholder = t("search.placeholder");
  const showTypewriter = !focused && query.length === 0;
  const typedPlaceholder = useTypewriterText(placeholder, showTypewriter);

  const { data: spacesData } = useQuery({
    queryKey: ["spaces-search-filter"],
    queryFn: () => optionalAuthApi<{ items: SpaceOption[] }>("/api/spaces"),
  });

  useEffect(() => {
    const q = searchParams.get("q")?.trim() ?? "";
    setQuery(q);
    setSubmitted(q);
    setSpaceSlug(searchParams.get("space_slug") ?? "");
    setAuthor(searchParams.get("author") ?? "");
    setTags(searchParams.get("tags") ?? "");
  }, [searchParams]);

  const filterKey = `${submitted}|${spaceSlug}|${author}|${tags}`;

  const { data, isLoading, isFetching, isError, error } = useQuery({
    queryKey: ["search", filterKey],
    queryFn: () =>
      optionalAuthApi<{ items: SearchResult[]; total: number }>(
        buildSearchUrl({ q: submitted, space_slug: spaceSlug || undefined, author: author || undefined, tags: tags || undefined }),
      ),
    enabled: submitted.length > 0,
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const next = query.trim();
    setSubmitted(next);
    const sp = new URLSearchParams();
    if (next) sp.set("q", next);
    if (spaceSlug) sp.set("space_slug", spaceSlug);
    if (author) sp.set("author", author);
    if (tags) sp.set("tags", tags);
    setSearchParams(sp);
  }

  const showResults = submitted.length > 0 && !isLoading && !isFetching && !isError;

  const [feedbackSent, setFeedbackSent] = useState(false);

  const ragAsk = useMutation({
    mutationFn: (question: string) =>
      api<RagAnswer>("/api/rag/ask", { method: "POST", body: JSON.stringify({ question }) }),
    onSuccess: () => setFeedbackSent(false),
    onError: (e) => setAskError(e instanceof ApiError ? e.message : t("common.failed")),
  });

  const ragFeedback = useMutation({
    mutationFn: (payload: { helpful: boolean; data: RagAnswer; question: string }) =>
      api("/api/rag/feedback", {
        method: "POST",
        body: JSON.stringify({
          question: payload.question,
          answer: payload.data.answer,
          helpful: payload.helpful,
          confidence: payload.data.confidence ?? 0,
          sources: payload.data.sources ?? [],
          citations: payload.data.citations ?? [],
        }),
      }),
    onSuccess: () => setFeedbackSent(true),
  });

  return (
    <div className={cn(pageShellClass, "py-10")}>
      <FadeIn>
        <div className="flex flex-wrap gap-2" data-tour="search-main">
          <button type="button" className={cn("btn-secondary", mode === "search" && "!bg-primary !text-on-primary")} onClick={() => setMode("search")}>
            <Search className="h-4 w-4" />
            {t("search.title")}
          </button>
          {isAuthenticated && (
            <button type="button" className={cn("btn-secondary", mode === "ask" && "!bg-primary !text-on-primary")} onClick={() => setMode("ask")}>
              <Sparkles className="h-4 w-4" />
              {t("rag.title")}
            </button>
          )}
        </div>
        {mode === "search" ? (
          <>
            <h1 className="mt-4 text-3xl font-bold gradient-text">{t("search.title")}</h1>
            <form onSubmit={handleSubmit} className="mt-6 space-y-3">
              <div className="flex gap-2">
                <div className="relative min-w-0 flex-1">
                  <Search className="pointer-events-none absolute left-3 top-1/2 z-10 h-4 w-4 -translate-y-1/2 text-subtle" />
                  <input
                    className="input-field pl-10"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    onFocus={() => setFocused(true)}
                    onBlur={() => setFocused(false)}
                    aria-label={placeholder}
                    placeholder={showTypewriter ? "" : placeholder}
                  />
                  {showTypewriter && (
                    <span
                      className="pointer-events-none absolute left-10 top-1/2 z-10 flex max-w-[calc(100%-3rem)] -translate-y-1/2 items-center text-sm text-subtle"
                      aria-hidden
                    >
                      <span className="truncate">{typedPlaceholder}</span>
                      <span className="typewriter-cursor ml-px inline-block h-[1.1em] w-px shrink-0 bg-subtle" />
                    </span>
                  )}
                </div>
                <button type="submit" className="btn-primary !px-5">
                  {t("search.submit")}
                </button>
              </div>
              <div className="flex flex-wrap gap-2">
                <SelectField
                  className="min-w-[10rem] flex-1"
                  value={spaceSlug}
                  onChange={(e) => setSpaceSlug(e.target.value)}
                  aria-label={t("search.filterSpace")}
                >
                  <option value="">{t("search.allSpaces")}</option>
                  {spacesData?.items.map((s) => (
                    <option key={s.id} value={s.slug}>
                      {s.name}
                    </option>
                  ))}
                </SelectField>
                <input
                  className="input-field min-w-[8rem] flex-1"
                  placeholder={t("search.filterAuthor")}
                  value={author}
                  onChange={(e) => setAuthor(e.target.value)}
                />
                <input
                  className="input-field min-w-[8rem] flex-1"
                  placeholder={t("search.filterTags")}
                  value={tags}
                  onChange={(e) => setTags(e.target.value)}
                />
              </div>
            </form>
          </>
        ) : (
          <>
            <h1 className="mt-4 text-3xl font-bold gradient-text">{t("rag.title")}</h1>
            <p className="mt-2 text-sm text-muted">{t("rag.subtitle")}</p>
            <textarea
              className="input-field mt-4 min-h-[100px] w-full"
              placeholder={t("rag.placeholder")}
              value={askQuestion}
              onChange={(e) => setAskQuestion(e.target.value)}
            />
            {askError && <p className="mt-2 text-sm text-danger-soft">{askError}</p>}
            <button
              type="button"
              className="btn-primary mt-3"
              disabled={!askQuestion.trim() || ragAsk.isPending}
              onClick={() => {
                setAskError("");
                ragAsk.mutate(askQuestion.trim());
              }}
            >
              {ragAsk.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : t("rag.ask")}
            </button>
            {ragAsk.data && (
              <div className="mt-6 space-y-4">
                <div className="glass p-4 text-sm text-fg whitespace-pre-wrap">{ragAsk.data.answer}</div>

                {ragAsk.data.citations && ragAsk.data.citations.length > 0 && (
                  <div className="space-y-2">
                    <p className="text-sm font-medium text-fg">{t("rag.citations")}</p>
                    {ragAsk.data.citations.map((c) => (
                      <blockquote
                        key={`${c.document_id}-${c.quote.slice(0, 40)}`}
                        className="border-l-2 border-primary/60 bg-surface/50 px-3 py-2 text-sm"
                      >
                        <p className="text-fg italic">&ldquo;{c.quote}&rdquo;</p>
                        <Link
                          to={`/spaces/${c.space_slug}/docs/${c.doc_slug}`}
                          className="mt-1 inline-block text-xs text-primary hover:underline"
                        >
                          {c.title}
                        </Link>
                      </blockquote>
                    ))}
                  </div>
                )}

                {ragAsk.data.low_confidence && ragAsk.data.follow_up_questions && ragAsk.data.follow_up_questions.length > 0 && (
                  <div className="glass border border-amber-500/30 p-4">
                    <p className="text-sm font-medium text-fg">{t("rag.lowConfidence")}</p>
                    <p className="mt-1 text-xs text-muted">{t("rag.followUp")}</p>
                    <ul className="mt-2 space-y-2">
                      {ragAsk.data.follow_up_questions.map((q) => (
                        <li key={q}>
                          <button
                            type="button"
                            className="text-left text-sm text-primary hover:underline"
                            onClick={() => {
                              setAskQuestion(q);
                              setAskError("");
                              ragAsk.mutate(q);
                            }}
                          >
                            {q}
                          </button>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {ragAsk.data.sources?.length > 0 && (
                  <div>
                    <p className="text-sm font-medium text-fg">{t("rag.sources")}</p>
                    <ul className="mt-2 space-y-2 text-sm">
                      {ragAsk.data.sources.map((s, i) => (
                        <li key={`${s.space_slug}-${s.doc_slug}`}>
                          <Link to={`/spaces/${s.space_slug}/docs/${s.doc_slug}`} className="text-primary hover:underline">
                            {s.title}
                          </Link>
                          {i === 0 && (
                            <span className="ml-2 text-xs font-medium text-primary/80">{t("rag.bestMatch")}</span>
                          )}
                          <p className="text-xs text-muted">{s.snippet}</p>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                <div className="flex flex-wrap items-center gap-2 border-t border-border pt-3">
                  {feedbackSent ? (
                    <span className="text-xs text-muted">{t("rag.feedbackThanks")}</span>
                  ) : (
                    <>
                      <button
                        type="button"
                        className="btn-secondary text-xs"
                        disabled={ragFeedback.isPending}
                        onClick={() =>
                          ragFeedback.mutate({ helpful: true, data: ragAsk.data!, question: askQuestion.trim() })
                        }
                      >
                        <ThumbsUp className="h-3.5 w-3.5" />
                        {t("rag.feedbackHelpful")}
                      </button>
                      <button
                        type="button"
                        className="btn-secondary text-xs"
                        disabled={ragFeedback.isPending}
                        onClick={() =>
                          ragFeedback.mutate({ helpful: false, data: ragAsk.data!, question: askQuestion.trim() })
                        }
                      >
                        <ThumbsDown className="h-3.5 w-3.5" />
                        {t("rag.feedbackNotHelpful")}
                      </button>
                    </>
                  )}
                </div>
              </div>
            )}
          </>
        )}
      </FadeIn>

      {mode === "search" && (isLoading || isFetching) && submitted && (
        <div className="flex justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin text-primary" />
        </div>
      )}

      {mode === "search" && isError && submitted && (
        <p className="mt-8 text-sm text-danger-soft">
          {error instanceof ApiError ? error.message : t("search.failed")}
        </p>
      )}

      {mode === "search" && showResults && data && data.total === 0 && (
        <p className="mt-8 text-sm text-muted">{t("search.noResults", { query: submitted })}</p>
      )}

      {mode === "search" && showResults && data && data.total > 0 && (
        <div className="mt-8 space-y-3">
          <p className="text-sm text-subtle">{t("search.results", { count: data.total })}</p>
          {data.items.map((item, i) => (
            <FadeIn key={item.id} delay={i * 0.03}>
              <div className="glass p-4">
                <Link
                  to={`/spaces/${item.space_slug}/docs/${item.slug}`}
                  className="font-medium text-primary hover:text-primary-hover"
                >
                  {item.title}
                </Link>
                <p className="mt-1 text-xs text-subtle">/{item.space_slug}</p>
                <p className="mt-1 text-sm text-muted">{item.snippet}</p>
                {item.author_name && (
                  <p className="mt-2 text-xs text-subtle">
                    {t("common.by")} {item.author_name}
                  </p>
                )}
                {item.tags && item.tags.length > 0 && (
                  <div className="mt-2 flex flex-wrap gap-1">
                    {item.tags.map((tag) => (
                      <span key={tag} className="badge badge-primary text-xs">
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </FadeIn>
          ))}
        </div>
      )}
    </div>
  );
}
