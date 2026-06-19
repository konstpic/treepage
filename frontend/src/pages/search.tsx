import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { Loader2, Search } from "lucide-react";
import { ApiError, optionalAuthApi } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { SelectField } from "@/components/select-field";
import { useI18n } from "@/lib/i18n";
import { useTypewriterText } from "@/lib/use-typewriter";

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

export function SearchPage() {
  const { t } = useI18n();
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

  return (
    <div className="mx-auto max-w-3xl px-4 py-10 sm:px-6">
      <FadeIn>
        <h1 className="text-3xl font-bold gradient-text">{t("search.title")}</h1>
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
      </FadeIn>

      {(isLoading || isFetching) && submitted && (
        <div className="flex justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin text-primary" />
        </div>
      )}

      {isError && submitted && (
        <p className="mt-8 text-sm text-danger-soft">
          {error instanceof ApiError ? error.message : t("search.failed")}
        </p>
      )}

      {showResults && data && data.total === 0 && (
        <p className="mt-8 text-sm text-muted">{t("search.noResults", { query: submitted })}</p>
      )}

      {showResults && data && data.total > 0 && (
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
