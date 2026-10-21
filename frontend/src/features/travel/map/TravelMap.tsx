import 'leaflet/dist/leaflet.css';
import type { Feature, Geometry, Polygon } from 'geojson';
import L, {
  type GeoJSON as GeoJSONLayer,
  type LatLngBoundsExpression,
  type Map as LeafletMap,
} from 'leaflet';
import { useReducedMotion } from 'motion/react';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  CircleMarker,
  GeoJSON,
  MapContainer,
  Pane,
  Polyline,
  Tooltip,
  useMap,
} from 'react-leaflet';
import { useTranslation } from 'react-i18next';
import type { Country, Place } from '../../../api/types';
import { useLanguage } from '../../../app/providers/language-context';
import { useTheme } from '../../../app/providers/theme-context';
import type { Selection } from '../useSelection';
import { ABKHAZIA_INNER_BORDER, ABKHAZIA_OUTER_BORDER } from './abkhazia';
import { readMapColors, type MapColors } from './colors';
import styles from './TravelMap.module.css';
import { featureKey, useWorldGeo, type WorldGeo } from './useWorldGeo';

const FLY_DURATION = 0.8; // seconds (spec §7)
const WORLD_CENTER: L.LatLngExpression = [25, 10];
const MIN_ZOOM = 1.5;
const PADDING: L.PointExpression = [40, 40];

interface Props {
  countries: Country[];
  places: Place[];
  selection: Selection;
  onSelect: (s: Selection) => void;
}

type CountryFeature = Feature<Geometry, { name?: string; mask?: boolean }>;

// [lat, lng] for Leaflet.
const toLatLngs = (line: typeof ABKHAZIA_INNER_BORDER): L.LatLngTuple[] =>
  line.map(([lng = 0, lat = 0]) => [lat, lng]);
const ABKHAZIA_INNER = toLatLngs(ABKHAZIA_INNER_BORDER);
const ABKHAZIA_OUTER = toLatLngs(ABKHAZIA_OUTER_BORDER);

/**
 * Vector world map without tiles (spec §5.2, plan §5.2): visited countries filled,
 * a pin per place, fly-to on selection. The list below the map is its accessible
 * equivalent.
 */
export default function TravelMap({ countries, places, selection, onSelect }: Props) {
  const { t } = useTranslation();
  const { data: world } = useWorldGeo();
  const [map, setMap] = useState<LeafletMap | null>(null);

  return (
    <div className={styles.wrap} role="region" aria-label={t('travel.mapLabel')}>
      <MapContainer
        ref={setMap}
        className={styles.map}
        center={WORLD_CENTER}
        zoom={MIN_ZOOM}
        minZoom={MIN_ZOOM}
        maxZoom={10}
        zoomSnap={0.25}
        worldCopyJump={false}
        maxBounds={[
          [-85, -200],
          [85, 200],
        ]}
        maxBoundsViscosity={0.8}
        scrollWheelZoom={false}
        attributionControl={false}
      >
        {world && <Countries world={world} countries={countries} onSelect={onSelect} />}
        {world && <ExtraBorders />}
        {/* Own pane above the overlay pane: the countries load asynchronously and would
            otherwise be drawn on top of the pins and swallow their clicks. */}
        <Pane name="pins" style={{ zIndex: 450 }}>
          {places.map((p) => (
            <Pin
              key={p.id}
              place={p}
              selected={selection?.type === 'place' && selection.id === p.id}
              onSelect={onSelect}
            />
          ))}
        </Pane>
        <Controller world={world} countries={countries} places={places} selection={selection} />
      </MapContainer>
      {map && <ScrollWheelGate map={map} />}
    </div>
  );
}

function useMapColors(): MapColors {
  const { theme } = useTheme();
  // Re-read the tokens whenever the theme changes (M-01: re-colour without reload).
  return useMemo(() => readMapColors(theme), [theme]);
}

function Countries({
  world,
  countries,
  onSelect,
}: {
  world: WorldGeo;
  countries: Country[];
  onSelect: (s: Selection) => void;
}) {
  const { lang } = useLanguage();
  const colors = useMapColors();
  const layerRef = useRef<GeoJSONLayer>(null);
  const byKey = useMemo(() => new Map(countries.map((c) => [featureKey(c), c])), [countries]);

  const styleFor = useMemo(
    () =>
      (feature?: CountryFeature): L.PathOptions => {
        if (feature?.properties.mask) {
          return { fillColor: colors.land, fillOpacity: 1, weight: 0, className: styles.mask };
        }
        const visited = feature?.id !== undefined && byKey.has(String(feature.id));
        // Places drawn on top of a country (Abkhazia): their borders come from
        // ExtraBorders, so the inner one can be dashed.
        const extra = typeof feature?.id === 'string' && !/^\d+$/.test(feature.id);
        return {
          fillColor: visited ? colors.accent : colors.land,
          fillOpacity: visited ? 0.55 : 1,
          color: colors.border,
          weight: extra ? 0 : 0.6,
          // Applied when the SVG path is created (pointer cursor on visited countries only).
          className: visited ? styles.visited : styles.land,
        };
      },
    [byKey, colors],
  );

  // Theme change: restyle in place instead of rebuilding the layer.
  useEffect(() => {
    layerRef.current?.setStyle(styleFor as L.StyleFunction);
  }, [styleFor]);

  const onEachFeature = (feature: CountryFeature, layer: L.Layer) => {
    const country = feature.id !== undefined ? byKey.get(String(feature.id)) : undefined;
    if (!country || !(layer instanceof L.Path)) return; // non-visited: no tooltip, no action (M-04)
    layer.bindTooltip(country.name, { sticky: true, direction: 'top', className: styles.tooltip });
    layer.on({
      mouseover: () => {
        layer.setStyle({ fillOpacity: 0.8 });
      },
      mouseout: () => {
        layer.setStyle({ fillOpacity: 0.55 });
      },
      click: () => {
        onSelect({ type: 'country', iso2: country.iso2 });
      },
      // The SVG path exists once the layer is on the map.
      add: () => {
        makeKeyboardButton(layer.getElement(), country.name, () => {
          onSelect({ type: 'country', iso2: country.iso2 });
        });
      },
    });
  };

  return (
    <GeoJSON
      // Tooltips and handlers are bound once per layer: rebuild on language or data change.
      key={`${lang}:${[...byKey.keys()].join(',')}`}
      ref={layerRef}
      data={world}
      style={styleFor as L.StyleFunction}
      onEachFeature={onEachFeature}
    />
  );
}

/**
 * Borders of places drawn on top of a country (Abkhazia), as Google Maps shows them:
 * the disputed inner boundary dashed, the coast and international border solid.
 */
function ExtraBorders() {
  const colors = useMapColors();
  const solid = { color: colors.border, weight: 0.6, interactive: false };
  return (
    // Above the countries, below the pins.
    <Pane name="extra-borders" style={{ zIndex: 420 }}>
      <Polyline positions={ABKHAZIA_OUTER} pathOptions={solid} />
      <Polyline
        positions={ABKHAZIA_INNER}
        pathOptions={{ ...solid, weight: 1, dashArray: '4 3' }}
      />
    </Pane>
  );
}

function Pin({
  place,
  selected,
  onSelect,
}: {
  place: Place;
  selected: boolean;
  onSelect: (s: Selection) => void;
}) {
  const colors = useMapColors();
  const marker = useRef<L.CircleMarker>(null);

  // Keyboard and screen-reader access to the pin; re-run when the name changes (language).
  useEffect(() => {
    return makeKeyboardButton(marker.current?.getElement(), place.name, () => {
      onSelect({ type: 'place', id: place.id });
    });
  }, [place.id, place.name, onSelect]);

  return (
    <CircleMarker
      ref={marker}
      center={[place.lat, place.lng]}
      radius={selected ? 9 : 6}
      pathOptions={{
        color: '#ffffff',
        weight: 2,
        fillColor: colors.accent,
        fillOpacity: 1,
        className: styles.pin,
      }}
      eventHandlers={{
        click: () => {
          onSelect({ type: 'place', id: place.id });
        },
      }}
    >
      <Tooltip direction="top" offset={[0, -6]} className={styles.tooltip}>
        {place.name}
      </Tooltip>
    </CircleMarker>
  );
}

/**
 * Clickable SVG paths are focusable, so give them what a button needs: a role, a
 * name and Enter/Space activation (the list below the map remains the full
 * alternative). Returns a cleanup function.
 */
function makeKeyboardButton(el: Element | undefined, name: string, activate: () => void) {
  if (!el) return undefined;
  el.setAttribute('role', 'button');
  el.setAttribute('tabindex', '0');
  el.setAttribute('aria-label', name);
  const onKey = (e: Event) => {
    const key = (e as KeyboardEvent).key;
    if (key === 'Enter' || key === ' ') {
      e.preventDefault();
      activate();
    }
  };
  el.addEventListener('keydown', onKey);
  return () => {
    el.removeEventListener('keydown', onKey);
  };
}

/** Initial fit and fly-to on selection changes (instant under reduced motion). */
function Controller({
  world,
  countries,
  places,
  selection,
}: {
  world: WorldGeo | undefined;
  countries: Country[];
  places: Place[];
  selection: Selection;
}) {
  const map = useMap();
  const reduced = useReducedMotion() ?? false;

  // Initial view: all places with padding; with none, the whole world (M-06).
  const fitted = useRef(false);
  useEffect(() => {
    if (fitted.current) return;
    fitted.current = true;
    if (places.length === 0) {
      map.setView(WORLD_CENTER, MIN_ZOOM, { animate: false });
      return;
    }
    const bounds = L.latLngBounds(places.map((p) => [p.lat, p.lng] as [number, number]));
    map.fitBounds(bounds, { padding: PADDING, maxZoom: 5, animate: false });
  }, [map, places]);

  const placeId = selection?.type === 'place' ? selection.id : null;
  const iso2 = selection?.type === 'country' ? selection.iso2 : null;

  useEffect(() => {
    if (!placeId) return;
    const p = places.find((x) => x.id === placeId);
    if (!p) return;
    const zoom = Math.max(map.getZoom(), 5);
    if (reduced) map.setView([p.lat, p.lng], zoom, { animate: false });
    else map.flyTo([p.lat, p.lng], zoom, { duration: FLY_DURATION });
  }, [map, placeId, places, reduced]);

  useEffect(() => {
    if (!iso2 || !world) return;
    const country = countries.find((c) => c.iso2 === iso2);
    const key = country && featureKey(country);
    const f = key ? world.features.find((x) => String(x.id) === key) : undefined;
    let bounds: LatLngBoundsExpression | null = f ? mainlandBounds(f) : null;
    if (!bounds) {
      // No polygon (e.g. Kosovo): fit the country's places instead.
      const ids = countries.find((c) => c.iso2 === iso2)?.placeIds ?? [];
      const pts = places
        .filter((p) => ids.includes(p.id))
        .map((p) => [p.lat, p.lng] as [number, number]);
      bounds = pts.length ? L.latLngBounds(pts) : null;
    }
    if (!bounds) return;
    if (reduced) map.fitBounds(bounds, { padding: PADDING, maxZoom: 6, animate: false });
    else map.flyToBounds(bounds, { padding: PADDING, maxZoom: 6, duration: FLY_DURATION });
  }, [map, iso2, world, countries, places, reduced]);

  return null;
}

/**
 * Bounds of a country's largest polygon. Many countries have far-away parts (the
 * Netherlands' Caribbean islands, French Guiana, Alaska): fitting all of them would
 * barely zoom in, so fly to the mainland instead.
 */
function mainlandBounds(f: CountryFeature): L.LatLngBounds {
  const g = f.geometry;
  if (g.type !== 'MultiPolygon') return L.geoJSON(f).getBounds();
  let best: L.LatLngBounds | null = null;
  let bestArea = -1;
  for (const coords of g.coordinates) {
    const polygon: Polygon = { type: 'Polygon', coordinates: coords };
    const b = L.geoJSON(polygon).getBounds();
    const area = (b.getNorth() - b.getSouth()) * (b.getEast() - b.getWest());
    if (area > bestArea) {
      best = b;
      bestArea = area;
    }
  }
  return best ?? L.geoJSON(f).getBounds();
}

/** Scroll-wheel zoom only while the map is focused or clicked (no page scroll-trap, M-03). */
function ScrollWheelGate({ map }: { map: LeafletMap }) {
  useEffect(() => {
    const el = map.getContainer();
    const enable = () => {
      map.scrollWheelZoom.enable();
    };
    const disable = () => {
      map.scrollWheelZoom.disable();
    };
    el.addEventListener('focus', enable);
    el.addEventListener('click', enable);
    el.addEventListener('blur', disable);
    el.addEventListener('mouseleave', disable);
    return () => {
      el.removeEventListener('focus', enable);
      el.removeEventListener('click', enable);
      el.removeEventListener('blur', disable);
      el.removeEventListener('mouseleave', disable);
    };
  }, [map]);
  return null;
}
