import { AnimatePresence, motion } from 'motion/react';
import { useEffect } from 'react';
import { useLocation, useOutlet } from 'react-router';
import { Footer } from '../components/Footer';
import { Header } from '../components/Header';
import { SkipLink } from '../components/SkipLink';
import { usePageVariants } from '../lib/motion';
import styles from './Layout.module.css';

/**
 * App shell (FR-F01): skip link, header, the animated page outlet and the footer.
 * The outgoing page fades out before the incoming one fades in (spec §7).
 */
export function Layout() {
  const location = useLocation();
  const outlet = useOutlet();
  const pageVariants = usePageVariants();

  // A new page starts at the top (query-only changes like ?place= keep the scroll).
  useEffect(() => {
    window.scrollTo({ top: 0 });
  }, [location.pathname]);

  return (
    <div className={styles.shell}>
      <SkipLink />
      <Header />
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
      <Footer />
    </div>
  );
}
