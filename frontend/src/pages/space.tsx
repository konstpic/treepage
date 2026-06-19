import { FileText } from "lucide-react";
import { FadeIn } from "@/components/motion-wrapper";
import { useI18n } from "@/lib/i18n";

export function SpaceIndexPage() {
  const { t } = useI18n();

  return (
    <FadeIn>
      <div className="glass flex flex-col items-center justify-center px-8 py-16 text-center">
        <FileText className="h-10 w-10 text-primary/60" />
        <h2 className="mt-4 text-lg font-medium text-fg-secondary">{t("space.selectDocument")}</h2>
        <p className="mt-2 max-w-sm text-sm text-subtle">{t("space.selectDocumentHint")}</p>
      </div>
    </FadeIn>
  );
}
