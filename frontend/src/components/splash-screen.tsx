import { motion } from "framer-motion";
import { TreePageLogo, TREE_LOGO_DRAW_MS } from "@/components/treepage-logo";
import { useBrandingStore } from "@/lib/store";

export function SplashScreen() {
  const projectName = useBrandingStore((s) => s.projectName);
  const textDelay = TREE_LOGO_DRAW_MS / 1000 - 0.55;

  return (
    <motion.div
      className="splash-screen"
      role="status"
      aria-live="polite"
      aria-label="Loading"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.5, ease: [0.4, 0, 0.2, 1] }}
    >
      <div className="splash-screen__glow" aria-hidden />

      <div className="relative flex flex-col items-center gap-6">
        <TreePageLogo size={128} />

        <motion.div
          className="flex flex-col items-center gap-1 overflow-visible pb-1"
          initial={{ opacity: 0, y: 14 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: textDelay, duration: 0.65, ease: [0.25, 0.46, 0.45, 0.94] }}
        >
          <h1 className="splash-screen__title">{projectName}</h1>
          <motion.div
            className="h-0.5 rounded-full bg-primary/60"
            initial={{ width: 0, opacity: 0 }}
            animate={{ width: 72, opacity: 1 }}
            transition={{ delay: textDelay + 0.2, duration: 0.5, ease: "easeOut" }}
          />
        </motion.div>
      </div>
    </motion.div>
  );
}
