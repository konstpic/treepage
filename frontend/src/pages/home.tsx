import { Link } from "react-router-dom";
import { BookOpen, GitBranch, Search, Shield } from "lucide-react";
import { FadeIn } from "@/components/motion-wrapper";
import { useI18n } from "@/lib/i18n";
import { pageShellClass } from "@/lib/utils";

export function HomePage() {
  const { t } = useI18n();

  const features = [
    { icon: BookOpen, title: t("home.features.markdown.title"), desc: t("home.features.markdown.desc") },
    { icon: GitBranch, title: t("home.features.git.title"), desc: t("home.features.git.desc") },
    { icon: Search, title: t("home.features.search.title"), desc: t("home.features.search.desc") },
    { icon: Shield, title: t("home.features.rbac.title"), desc: t("home.features.rbac.desc") },
  ];

  return (
    <div className={`${pageShellClass} py-16`}>
      <FadeIn className="text-center">
        <h1 className="text-4xl font-bold tracking-tight sm:text-6xl">
          {t("home.title")}{" "}
          <span className="gradient-text">{t("home.titleAccent")}</span>
        </h1>
        <p className="mx-auto mt-6 max-w-2xl text-lg text-muted">{t("home.subtitle")}</p>
        <div className="mt-10 flex flex-wrap justify-center gap-4">
          <Link to="/spaces" className="btn-primary">
            {t("home.browseSpaces")}
          </Link>
        </div>
      </FadeIn>

      <div className="mt-24 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {features.map((f, i) => (
          <FadeIn key={f.title} delay={i * 0.1}>
            <div className="glass-hover h-full p-6">
              <f.icon className="h-8 w-8 text-primary" />
              <h3 className="mt-4 text-lg font-semibold text-fg">{f.title}</h3>
              <p className="mt-2 text-sm text-muted">{f.desc}</p>
            </div>
          </FadeIn>
        ))}
      </div>
    </div>
  );
}
