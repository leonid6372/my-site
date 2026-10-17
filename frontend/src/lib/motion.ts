import type { Transition, Variants } from 'motion/react';

// Spec §7. MotionConfig reducedMotion="user" (App) drops transforms for users who
// prefer reduced motion and keeps only opacity changes.

/** Page enter: fade + 8 px rise over 300 ms; exit: fade out over 150 ms. */
export const pageVariants: Variants = {
  initial: { opacity: 0, y: 8 },
  enter: { opacity: 1, y: 0, transition: { duration: 0.3, ease: 'easeOut' } },
  exit: { opacity: 0, transition: { duration: 0.15, ease: 'easeIn' } },
};

/** Section containers: stagger children by 60 ms on first render. */
export const staggerContainer: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

export const staggerItem: Variants = {
  hidden: { opacity: 0, y: 8 },
  show: { opacity: 1, y: 0, transition: { duration: 0.3, ease: 'easeOut' } },
};

export const panelTransition: Transition = { duration: 0.25, ease: 'easeOut' };
