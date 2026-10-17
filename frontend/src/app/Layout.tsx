import { AnimatePresence, motion } from 'motion/react';
import { useEffect, type ReactNode } from 'react';
import { useLocation, useOutlet } from 'react-router';
import { pageVariants } from '../lib/motion';
import styles from './Layout.module.css';

/**
 * App shell: header, the animated page outlet and the footer. The outgoing page fades
 * out before the incoming one fades in (AnimatePresence mode="wait", spec §7).
 */
export function Layout({ header, footer }: { header?: ReactNode; footer?: ReactNode }) {
  const location = useLocation();
  const outlet = useOutlet();

  // A new page starts at the top (query-only changes like ?place= keep the scroll).
  useEffect(() => {
    window.scrollTo({ top: 0 });
  }, [location.pathname]);

  return (
    <div className={styles.shell}>
      {header}
      <main id="content" tabIndex={-1} className={styles.main}>
        <AnimatePresence mode="wait" initial={false}>
          <motion.div
            key={location.pathname}
            variants={pageVariants}
            initial="initial"
            animate="enter"
            exit="exit"
          >
            {outlet}
          </motion.div>
        </AnimatePresence>
      </main>
      {footer}
    </div>
  );
}
