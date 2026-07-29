import React, { useEffect, useRef, useState, useMemo } from 'react';
import Globe, { GlobeMethods } from 'react-globe.gl';
import * as THREE from 'three';
import { mockEvents, mockArcs } from '../lib/data';

interface GlobeViewProps {
  onSelectCountry?: (country: string) => void;
}

export default function GlobeView({ onSelectCountry }: GlobeViewProps) {
  const globeEl = useRef<GlobeMethods | undefined>(undefined);
  const [countries, setCountries] = useState({ features: [] });
  const [hoverD, setHoverD] = useState<any>();
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 });
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Load country polygons
    fetch('https://unpkg.com/world-atlas/countries-110m.json')
      .then(res => res.json())
      .then(topoData => {
        // We need topojson-client to convert topojson to geojson, but maybe we can just fetch geojson directly
        // Let's use a geojson url directly to avoid adding topojson-client
      });
  }, []);

  useEffect(() => {
    fetch('https://raw.githubusercontent.com/nvkelso/natural-earth-vector/master/geojson/ne_110m_admin_0_countries.geojson')
      .then(res => res.json())
      .then(data => setCountries(data));
  }, []);

  useEffect(() => {
    if (!containerRef.current) return;
    const updateSize = () => {
      if (containerRef.current) {
        setDimensions({
          width: containerRef.current.clientWidth,
          height: containerRef.current.clientHeight
        });
      }
    };
    updateSize();
    window.addEventListener('resize', updateSize);
    return () => window.removeEventListener('resize', updateSize);
  }, []);

  useEffect(() => {
    // Auto-rotate
    if (globeEl.current) {
      globeEl.current.controls().autoRotate = true;
      globeEl.current.controls().autoRotateSpeed = 0.5;
      
      // Set initial point of view
      globeEl.current.pointOfView({ lat: 20, lng: 0, altitude: 2 });
    }
  }, [globeEl.current]);

  const colorScale = (sentiment: string) => {
    if (sentiment === 'negative') return '#ef4444'; // red
    if (sentiment === 'positive') return '#10b981'; // green
    return '#3b82f6'; // blue
  };

  return (
    <div ref={containerRef} className="w-full h-full relative cursor-crosshair">
      <Globe
        ref={globeEl}
        width={dimensions.width}
        height={dimensions.height}
        globeImageUrl="//unpkg.com/three-globe/example/img/earth-night.jpg"
        backgroundImageUrl="//unpkg.com/three-globe/example/img/night-sky.png"
        
        // Polygons
        polygonsData={countries.features}
        polygonAltitude={d => d === hoverD ? 0.02 : 0.01}
        polygonCapColor={d => d === hoverD ? 'rgba(59, 130, 246, 0.4)' : 'rgba(10, 10, 10, 0.6)'}
        polygonSideColor={() => 'rgba(0, 100, 100, 0.15)'}
        polygonStrokeColor={() => '#111'}
        onPolygonHover={setHoverD}
        onPolygonClick={(d: any) => {
          if (onSelectCountry && d.properties.ADMIN) {
            onSelectCountry(d.properties.ADMIN);
            
            // Focus on country
            // Note: simplistic centering, doesn't calculate actual country centroid, just zooming slightly
          }
        }}
        
        // Rings for events
        ringsData={mockEvents}
        ringColor={(d: any) => colorScale(d.sentiment)}
        ringMaxRadius="score"
        ringPropagationSpeed={2}
        ringRepeatPeriod={1000}
        
        // Labels for events
        labelsData={mockEvents}
        labelLat={(d: any) => d.lat}
        labelLng={(d: any) => d.lng}
        labelText={(d: any) => d.title}
        labelSize={1.5}
        labelDotRadius={0.5}
        labelColor={(d: any) => colorScale(d.sentiment)}
        labelResolution={2}
        labelAltitude={0.05}
        
        // Arcs
        arcsData={mockArcs}
        arcColor="color"
        arcDashLength={0.4}
        arcDashGap={0.2}
        arcDashAnimateTime={1500}
        arcAltitude={0.2}
        arcStroke={0.5}
      />
    </div>
  );
}
