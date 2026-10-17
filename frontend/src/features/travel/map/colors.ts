// Leaflet writes colours into SVG attributes, which cannot use CSS custom properties.
// Resolve the design tokens to concrete rgb() strings (re-run when the theme changes).

export interface MapColors {
  bg: string;
  land: string;
  border: string;
  accent: string;
  visited: string;
}

function resolve(token: string, probe: HTMLElement): string {
  probe.style.color = `var(${token})`;
  return getComputedStyle(probe).color;
}

/** `_theme` is only a cache key: the values come from the tokens currently on <html>. */
export function readMapColors(_theme: string): MapColors {
  const probe = document.createElement('span');
  probe.style.display = 'none';
  document.body.appendChild(probe);
  try {
    return {
      bg: resolve('--bg', probe),
      land: resolve('--land', probe),
      border: resolve('--border', probe),
      accent: resolve('--accent', probe),
      visited: resolve('--visited', probe),
    };
  } finally {
    probe.remove();
  }
}
