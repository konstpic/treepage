import { type ReactNode } from "react";
import { motion } from "framer-motion";
import { useSplashStore, SCATTER_ANIM_MS } from "@/lib/splash-store";

const scatterEase: [number, number, number, number] = [0.55, 0, 0.15, 1];
const returnEase: [number, number, number, number] = [0.25, 0.46, 0.45, 0.94];

export interface ScatterExit {
  x: string | number;
  y: string | number;
  rotate?: number;
  delay?: number;
  scale?: number;
}

interface AuthScatterPieceProps {
  children: ReactNode;
  exit: ScatterExit;
  className?: string;
}

export function AuthScatterPiece({ children, exit, className }: AuthScatterPieceProps) {
  const scattering = useSplashStore((s) => s.phase === "scatter");

  return (
    <motion.div
      className={className}
      animate={
        scattering
          ? {
              opacity: 0,
              x: exit.x,
              y: exit.y,
              rotate: exit.rotate ?? 0,
              scale: exit.scale ?? 0.82,
              filter: "blur(10px)",
            }
          : { opacity: 1, x: 0, y: 0, rotate: 0, scale: 1, filter: "blur(0px)" }
      }
      transition={{
        duration: SCATTER_ANIM_MS / 1000,
        delay: exit.delay ?? 0,
        ease: scatterEase,
      }}
    >
      {children}
    </motion.div>
  );
}

interface AppShellProps {
  header: ReactNode;
  main: ReactNode;
  footer: ReactNode;
}

export function AppShell({ header, main, footer }: AppShellProps) {
  const phase = useSplashStore((s) => s.phase);
  const scatterMode = useSplashStore((s) => s.scatterMode);
  const scattering = phase === "scatter";
  const hidden = phase === "splash";
  const authFormScatter = scattering && scatterMode === "auth-form";

  const scatterTransition = (delay = 0) => ({
    duration: SCATTER_ANIM_MS / 1000,
    ease: scatterEase,
    delay,
  });

  const returnTransition = {
    duration: 0.45,
    ease: returnEase,
  };

  return (
    <>
      <motion.div
        className="relative z-10"
        animate={
          authFormScatter
            ? { opacity: 0, y: "-110vh", scale: 0.9, rotate: -3, filter: "blur(10px)" }
            : scattering
              ? { opacity: 0, y: "-110vh", scale: 0.82, rotate: -4, filter: "blur(12px)" }
              : hidden
                ? { opacity: 0 }
                : { opacity: 1, y: 0, scale: 1, rotate: 0, filter: "blur(0px)" }
        }
        transition={scattering ? scatterTransition(0) : returnTransition}
        style={{ pointerEvents: hidden || scattering ? "none" : "auto" }}
      >
        {header}
      </motion.div>

      <motion.main
        className="relative z-0 flex-1"
        animate={
          authFormScatter
            ? { opacity: 1, scale: 1, y: 0, x: 0, rotate: 0, filter: "blur(0px)" }
            : scattering
              ? { opacity: 0, y: "85vh", x: "-55vw", scale: 0.72, rotate: 5, filter: "blur(14px)" }
              : hidden
                ? { opacity: 0 }
                : { opacity: 1, scale: 1, y: 0, x: 0, rotate: 0, filter: "blur(0px)" }
        }
        transition={scattering && !authFormScatter ? scatterTransition(0.08) : returnTransition}
        style={{ pointerEvents: hidden || scattering ? "none" : "auto" }}
      >
        {main}
      </motion.main>

      <motion.div
        className="relative z-10"
        animate={
          authFormScatter
            ? { opacity: 0, y: "110vh", scale: 0.9, rotate: 3, filter: "blur(10px)" }
            : scattering
              ? { opacity: 0, y: "110vh", x: "30vw", scale: 0.82, rotate: 3, filter: "blur(12px)" }
              : hidden
                ? { opacity: 0 }
                : { opacity: 1, y: 0, x: 0, scale: 1, rotate: 0, filter: "blur(0px)" }
        }
        transition={scattering ? scatterTransition(0.14) : returnTransition}
        style={{ pointerEvents: hidden || scattering ? "none" : "auto" }}
      >
        {footer}
      </motion.div>
    </>
  );
}
