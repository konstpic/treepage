import { Link } from "react-router-dom";
import { useI18n } from "@/lib/i18n";
import { pageShellClass } from "@/lib/utils";

export function Footer() {
  const { t } = useI18n();

  return (
    <footer className="border-t border-default py-8">
      <div className={`${pageShellClass} flex flex-col items-center justify-between gap-4 sm:flex-row`}>
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
