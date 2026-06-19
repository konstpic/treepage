import { useId } from "react";
import { motion } from "framer-motion";

/**
 * Real outline of the "tree-pine" icon (viewBox 0 0 24 24), split into the
 * two sub-paths it's actually made of:
 *  - PINE_CANOPY: the rounded zig-zag silhouette (closed shape, traced as a line)
 *  - PINE_TRUNK: the short stem under the canopy
 * These are the SAME d-strings used in the static icon, so the "drawn" shape
 * always matches what's rendered when not animating.
 */
export const PINE_CANOPY =
  "m17 14 3 3.3a1 1 0 0 1-.7 1.7H4.7a1 1 0 0 1-.7-1.7L7 14h-.3a1 1 0 0 1-.7-1.7L9 9h-.2A1 1 0 0 1 8 7.3L12 3l4 4.3a1 1 0 0 1-.8 1.7H15l3 3.3a1 1 0 0 1-.7 1.7H17Z";
export const PINE_TRUNK = "M 12 23 L 12 19";
/** Drawn top → bottom so the stroke grows down from the canopy, not up from a base dot. */
export const PINE_TRUNK_DRAW = "M 12 19 L 12 23";

const CANOPY_DURATION = 1.5;
const TRUNK_DURATION = 0.3;
const CANOPY_EASE: [number, number, number, number] = [0.42, 0, 0.2, 1];
const STROKE = 2;

interface TreePageLogoProps {
  size?: number;
  className?: string;
  animate?: boolean;
  /** White on purple nav badge; brand colors on splash. */
  variant?: "brand" | "onPrimary";
}

export function TreePageLogo({
  size = 120,
  className,
  animate = true,
  variant = "brand",
}: TreePageLogoProps) {
  const uid = useId().replace(/:/g, "");
  const onPrimary = variant === "onPrimary";

  const stroke = onPrimary ? "#ffffff" : "rgb(var(--primary))";
  const strokeGhost = onPrimary ? "rgb(255 255 255 / 0.25)" : "rgb(var(--primary) / 0.2)";

  const canopyTransition = animate
    ? { duration: CANOPY_DURATION, ease: CANOPY_EASE }
    : { duration: 0 };
  const trunkTransition = animate
    ? {
        pathLength: { delay: CANOPY_DURATION, duration: TRUNK_DURATION, ease: CANOPY_EASE },
        opacity: { delay: CANOPY_DURATION, duration: 0.01 },
      }
    : { duration: 0 };

  return (
    <svg
      viewBox="0 0 24 24"
      width={size}
      height={size}
      className={className}
      aria-hidden
    >
      <defs>
        <filter id={`pine-glow-${uid}`} x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="0.8" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        {!onPrimary && (
          <radialGradient id={`pine-aura-${uid}`} cx="50%" cy="45%" r="50%">
            <stop offset="0%" stopColor="rgb(var(--primary))" stopOpacity="0.35" />
            <stop offset="100%" stopColor="rgb(var(--primary))" stopOpacity="0" />
          </radialGradient>
        )}
      </defs>

      {!onPrimary && (
        <motion.circle
          cx={12}
          cy={11}
          r={10}
          fill={`url(#pine-aura-${uid})`}
          initial={{ opacity: 0, scale: 0.6 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 1.2, ease: "easeOut" }}
        />
      )}

      {!animate && (
        <>
          <path
            d={PINE_CANOPY}
            fill="none"
            stroke={stroke}
            strokeWidth={STROKE}
            strokeLinecap="round"
            strokeLinejoin="round"
          />
          <path
            d={PINE_TRUNK}
            fill="none"
            stroke={stroke}
            strokeWidth={STROKE}
            strokeLinecap="round"
          />
        </>
      )}

      {animate && (
        <>
          {/* Ghost underlay — canopy only; trunk uses a single solid stroke */}
          <motion.path
            d={PINE_CANOPY}
            fill="none"
            stroke={strokeGhost}
            strokeWidth={STROKE}
            strokeLinecap="round"
            strokeLinejoin="round"
            initial={{ pathLength: 0, opacity: 1 }}
            animate={{ pathLength: 1, opacity: 0 }}
            transition={{
              pathLength: canopyTransition,
              opacity: { delay: CANOPY_DURATION * 0.85, duration: 0.2, ease: "easeOut" },
            }}
          />

          <motion.path
            d={PINE_CANOPY}
            fill="none"
            stroke={stroke}
            strokeWidth={STROKE}
            strokeLinecap="round"
            strokeLinejoin="round"
            filter={onPrimary ? undefined : `url(#pine-glow-${uid})`}
            initial={{ pathLength: 0 }}
            animate={{ pathLength: 1 }}
            transition={canopyTransition}
          />
          <motion.path
            d={PINE_TRUNK_DRAW}
            fill="none"
            stroke={stroke}
            strokeWidth={STROKE}
            strokeLinecap="round"
            initial={{ pathLength: 0, opacity: 0 }}
            animate={{ pathLength: 1, opacity: 1 }}
            transition={trunkTransition}
          />
        </>
      )}
    </svg>
  );
}

export const TREE_LOGO_DRAW_MS =
  Math.ceil((CANOPY_DURATION + TRUNK_DURATION) * 1000) + 200;
