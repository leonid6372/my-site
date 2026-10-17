import { useReducedMotion } from 'motion/react';
import { useEffect, useState, type RefObject } from 'react';

const easeOut = (t: number) => 1 - (1 - t) ** 3;

/**
 * Counts from 0 to `target` over `duration` ms once `ref` scrolls into view
 * (plan §5.5). Returns the target directly under reduced motion or when
 * IntersectionObserver is unavailable.
 */
export function useCountUp(target: number, ref: RefObject<Element | null>, duration = 800): number {
  const reduced = useReducedMotion() ?? false;
  const canAnimate = !reduced && typeof IntersectionObserver !== 'undefined';
  const [value, setValue] = useState(0);

  useEffect(() => {
    const el = ref.current;
    if (!canAnimate || !el) return;
    let frame = 0;
    const io = new IntersectionObserver((entries) => {
      if (!entries.some((e) => e.isIntersecting)) return;
      io.disconnect();
      const start = performance.now();
      const tick = (now: number) => {
        const t = Math.min(1, (now - start) / duration);
        setValue(t === 1 ? target : target * easeOut(t));
        if (t < 1) frame = requestAnimationFrame(tick);
      };
      frame = requestAnimationFrame(tick);
    });
    io.observe(el);
    return () => {
      io.disconnect();
      cancelAnimationFrame(frame);
    };
  }, [canAnimate, target, duration, ref]);

  return canAnimate ? value : target;
}
