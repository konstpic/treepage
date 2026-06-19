import { Link } from "react-router-dom";
import { useI18n } from "@/lib/i18n";

export function Footer() {
  const { t } = useI18n();

  return (
    <footer className="border-t border-default py-8">
      <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 px-4 sm:flex-row sm:px-6">
        <p className="text-sm text-subtle">{t("footer.tagline")}</p>
        <div className="flex gap-6 text-sm text-subtle">
          <Link to="/spaces" className="hover:text-fg-secondary">
            {t("footer.documentation")}
          </Link>
        </div>
      </div>
    </footer>
  );
}
