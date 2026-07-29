export interface NewsEvent {
  id: string;
  title: string;
  summary: string;
  country: string;
  countryCode: string;
  lat: number;
  lng: number;
  sentiment: 'positive' | 'negative' | 'neutral';
  score: number;
  topic: string;
  timestamp: string;
}

export const mockEvents: NewsEvent[] = [
  {
    id: '1',
    title: 'Cyberattack on Critical Infrastructure',
    summary: 'A major energy provider reported a sophisticated ransomware attack affecting multiple regional grids.',
    country: 'United States',
    countryCode: 'US',
    lat: 38.0,
    lng: -97.0,
    sentiment: 'negative',
    score: 0.95,
    topic: 'Cybersecurity',
    timestamp: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
  },
  {
    id: '2',
    title: 'New Trade Agreement Signed',
    summary: 'European nations have finalized a comprehensive digital trade agreement aiming to boost tech sector cooperation.',
    country: 'Germany',
    countryCode: 'DE',
    lat: 51.1657,
    lng: 10.4515,
    sentiment: 'positive',
    score: 0.82,
    topic: 'Economy',
    timestamp: new Date(Date.now() - 1000 * 60 * 25).toISOString(),
  },
  {
    id: '3',
    title: 'Semiconductor Supply Chain Disruption',
    summary: 'Logistical challenges in the APAC region are causing delays in semiconductor manufacturing and exports.',
    country: 'Taiwan',
    countryCode: 'TW',
    lat: 23.6978,
    lng: 120.9605,
    sentiment: 'negative',
    score: 0.88,
    topic: 'Technology',
    timestamp: new Date(Date.now() - 1000 * 60 * 45).toISOString(),
  },
  {
    id: '4',
    title: 'Breakthrough in Quantum Computing',
    summary: 'Researchers announce a major milestone in quantum error correction, paving the way for stable quantum computers.',
    country: 'Japan',
    countryCode: 'JP',
    lat: 36.2048,
    lng: 138.2529,
    sentiment: 'positive',
    score: 0.75,
    topic: 'Technology',
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
  },
  {
    id: '5',
    title: 'Geopolitical Tensions Escalate at Border',
    summary: 'Increased military presence reported near the border region, raising concerns among international observers.',
    country: 'Ukraine',
    countryCode: 'UA',
    lat: 48.3794,
    lng: 31.1656,
    sentiment: 'negative',
    score: 0.92,
    topic: 'Geopolitics',
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 3).toISOString(),
  },
];

export const mockArcs = [
  { startLat: 38.0, startLng: -97.0, endLat: 51.1657, endLng: 10.4515, color: ['#ef4444', '#10b981'] },
  { startLat: 23.6978, startLng: 120.9605, endLat: 38.0, endLng: -97.0, color: ['#ef4444', '#ef4444'] },
  { startLat: 36.2048, startLng: 138.2529, endLat: 51.1657, endLng: 10.4515, color: ['#10b981', '#10b981'] },
  { startLat: 48.3794, startLng: 31.1656, endLat: 55.7558, endLng: 37.6173, color: ['#ef4444', '#ef4444'] },
  { startLat: -23.5505, startLng: -46.6333, endLat: 40.7128, endLng: -74.0060, color: ['#3b82f6', '#10b981'] },
  { startLat: 1.3521, startLng: 103.8198, endLat: -33.8688, endLng: 151.2093, color: ['#10b981', '#3b82f6'] },
  { startLat: 28.6139, startLng: 77.2090, endLat: 51.5074, endLng: -0.1278, color: ['#f59e0b', '#3b82f6'] },
  { startLat: 39.9042, startLng: 116.4074, endLat: -1.2921, endLng: 36.8219, color: ['#f59e0b', '#10b981'] },
];
