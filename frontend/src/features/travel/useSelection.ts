import { useCallback, useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router';
import type { Place } from '../../api/types';

export type Selection = { type: 'place'; id: string } | { type: 'country'; iso2: string } | null;

/**
 * Travel page selection (plan §5.3). A place selection is mirrored to `?place=<id>`
 * (replace, so it doesn't add history entries); a country selection is local state.
 * An unknown `?place=` id is ignored and removed from the URL once data is loaded.
 */
export function useSelection(places: Place[] | undefined) {
  const [params, setParams] = useSearchParams();
  const [country, setCountry] = useState<string | null>(null);
  const placeParam = params.get('place');

  const knownPlace =
    placeParam !== null && places?.some((p) => p.id === placeParam) ? placeParam : null;

  useEffect(() => {
    if (places && placeParam !== null && knownPlace === null) {
      setParams(
        (p) => {
          p.delete('place');
          return p;
        },
        { replace: true },
      );
    }
  }, [places, placeParam, knownPlace, setParams]);

  const selection: Selection = useMemo(() => {
    if (knownPlace) return { type: 'place', id: knownPlace };
    if (country) return { type: 'country', iso2: country };
    return null;
  }, [knownPlace, country]);

  const select = useCallback(
    (next: Selection) => {
      setCountry(next?.type === 'country' ? next.iso2 : null);
      setParams(
        (p) => {
          if (next?.type === 'place') p.set('place', next.id);
          else p.delete('place');
          return p;
        },
        { replace: true },
      );
    },
    [setParams],
  );

  return { selection, select };
}
