import type { ReactNode } from 'react';
import styles from './Chip.module.css';

/** A small rounded label (interests, skills). */
export function Chip({ children }: { children: ReactNode }) {
  return <span className={styles.chip}>{children}</span>;
}

/** A wrapping list of chips. */
export function ChipList({ items, label }: { items: string[]; label?: string }) {
  return (
    <ul className={styles.list} aria-label={label}>
      {items.map((item) => (
        <li key={item}>
          <Chip>{item}</Chip>
        </li>
      ))}
    </ul>
  );
}
