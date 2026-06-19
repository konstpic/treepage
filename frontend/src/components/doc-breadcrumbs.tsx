import { Link } from "react-router-dom";
import { ChevronRight } from "lucide-react";
import { useI18n } from "@/lib/i18n";

interface DocBreadcrumbsProps {
  spaceSlug: string;
  spaceName?: string;
  docPath: string;
  docTitle: string;
}

export function DocBreadcrumbs({ spaceSlug, spaceName, docPath, docTitle }: DocBreadcrumbsProps) {
  const { t } = useI18n();
  const segments = docPath.replace(/^\//, "").split("/").filter(Boolean);
  const fileName = segments.pop() ?? docTitle;

  return (
    <nav aria-label={t("document.breadcrumbs")} className="mb-4 flex flex-wrap items-center gap-1 text-sm text-subtle">
      <Link to="/spaces" className="hover:text-primary">
        {t("space.backToSpaces")}
      </Link>
      <ChevronRight className="h-3.5 w-3.5 shrink-0" />
      <Link to={`/spaces/${spaceSlug}`} className="hover:text-primary">
        {spaceName || spaceSlug}
      </Link>
      {segments.map((seg, i) => {
        const partial = segments.slice(0, i + 1).join("/");
        return (
          <span key={partial} className="flex items-center gap-1">
            <ChevronRight className="h-3.5 w-3.5 shrink-0" />
            <span>{seg}</span>
          </span>
        );
      })}
      <ChevronRight className="h-3.5 w-3.5 shrink-0" />
      <span className="font-medium text-fg">{fileName}</span>
    </nav>
  );
}
