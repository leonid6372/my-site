import { useReducedMotion, type Transition, type Variants } from 'motion/react';

// Spec §7. With prefers-reduced-motion: no transforms at all (not even an initial
// offset for one frame) and only opacity changes of at most 100 ms.

const REDUCED_FADE = { duration: 0.1 } as const;

/** Page enter: fade + 8 px rise over 300 ms; exit: fade out over 150 ms. */
const pageVariants: Variants = {
  initial: { opacity: 0, y: 8 },
  enter: { opacity: 1, y: 0, transition: { duration: 0.3, ease: 'easeOut' } },
  exit: { opacity: 0, transition: { duration: 0.15, ease: 'easeIn' } },
};

const reducedPageVariants: Variants = {
  initial: { opacity: 0 },
  enter: { opacity: 1, transition: REDUCED_FADE },
  exit: { opacity: 0, transition: REDUCED_FADE },
};

export function usePageVariants(): Variants {
  return useReducedMotion() ? reducedPageVariants : pageVariants;
}

/** Section containers: stagger children by 60 ms on first render. */
export const staggerContainer: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

const staggerItem: Variants = {
  hidden: { opacity: 0, y: 8 },
  show: { opacity: 1, y: 0, transition: { duration: 0.3, ease: 'easeOut' } },
};

const reducedStaggerItem: Variants = {
  hidden: { opacity: 0 },
  show: { opacity: 1, transition: REDUCED_FADE },
};

export function useStaggerItem(): Variants {
  return useReducedMotion() ? reducedStaggerItem : staggerItem;
}

export const panelTransition: Transition = { duration: 0.25, ease: 'easeOut' };
