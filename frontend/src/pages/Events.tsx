import { useQuery } from '@tanstack/react-query';
import { eventsAPI, sportsAPI } from '../lib/api';
import { Link, useSearchParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useState } from 'react';

interface Sport { id: number; name: string; slug: string; icon: string; }
interface Event {
    id: string; sport_id: number; home_team: string; away_team: string;
    start_time: string; status: string; sport_name: string; sport_icon: string; markets_count: number;
}

export default function EventsPage() {
    const [searchParams, setSearchParams] = useSearchParams();
    const [search, setSearch] = useState('');
    const activeSport = searchParams.get('sport_id');

    const { data: sportsData } = useQuery({ queryKey: ['sports'], queryFn: () => sportsAPI.list() });
    const { data: eventsData, isLoading } = useQuery({
        queryKey: ['events', activeSport, search],
        queryFn: () => eventsAPI.list({
            ...(activeSport ? { sport_id: activeSport } : {}),
            ...(search ? { search } : {}),
            status: 'upcoming',
            limit: '50',
        }),
    });

    const sports: Sport[] = sportsData?.data?.sports || [];
    const events: Event[] = eventsData?.data?.events || [];

    return (
        <div>
            {/* Hero */}
            <div className="mb-6">
                <h1 className="text-3xl font-bold mb-1">
                    <span className="bg-gradient-to-r from-accent to-emerald-300 bg-clip-text text-transparent">Live & Upcoming</span> Events
                </h1>
                <p className="text-text-muted">Place your bets on the biggest matches</p>
            </div>

            {/* Sports filter */}
            <div className="flex gap-2 mb-4 overflow-x-auto pb-2">
                <button
                    onClick={() => setSearchParams({})}
                    className={`px-4 py-2 rounded-lg text-sm font-medium transition-all whitespace-nowrap ${!activeSport ? 'bg-accent text-bg-primary' : 'bg-bg-card text-text-secondary hover:text-text-primary border border-border-light'
                        }`}
                >All Sports</button>
                {sports.map((s) => (
                    <button
                        key={s.id}
                        onClick={() => setSearchParams({ sport_id: String(s.id) })}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-all whitespace-nowrap flex items-center gap-1.5 ${activeSport === String(s.id) ? 'bg-accent text-bg-primary' : 'bg-bg-card text-text-secondary hover:text-text-primary border border-border-light'
                            }`}
                    >
                        <span>{s.icon}</span> {s.name}
                    </button>
                ))}
            </div>

            {/* Search */}
            <div className="mb-6">
                <input
                    type="text"
                    placeholder="Search teams..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="input-field max-w-sm"
                />
            </div>

            {/* Events grid */}
            {isLoading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {[...Array(6)].map((_, i) => (
                        <div key={i} className="card animate-pulse h-36" />
                    ))}
                </div>
            ) : events.length === 0 ? (
                <div className="card text-center py-12">
                    <p className="text-text-muted text-lg">No events found</p>
                    <p className="text-text-muted text-sm mt-1">Check back later for upcoming matches</p>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {events.map((event, idx) => (
                        <motion.div
                            key={event.id}
                            initial={{ y: 20, opacity: 0 }}
                            animate={{ y: 0, opacity: 1 }}
                            transition={{ delay: idx * 0.05 }}
                        >
                            <EventCard event={event} />
                        </motion.div>
                    ))}
                </div>
            )}
        </div>
    );
}

function EventCard({ event }: { event: Event }) {
    const startTime = new Date(event.start_time);
    const isLive = event.status === 'live';

    return (
        <Link to={`/events/${event.id}`} className="card group hover:border-accent/30 transition-all duration-300 block">
            <div className="flex items-center justify-between mb-3">
                <span className="text-xs text-text-muted flex items-center gap-1">
                    <span>{event.sport_icon}</span> {event.sport_name}
                </span>
                {isLive ? (
                    <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-danger/20 text-danger animate-pulse">● LIVE</span>
                ) : (
                    <span className="text-xs text-text-muted">
                        {startTime.toLocaleDateString('en', { month: 'short', day: 'numeric' })} • {startTime.toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' })}
                    </span>
                )}
            </div>

            <div className="flex items-center justify-between">
                <div className="flex-1">
                    <p className="font-semibold text-text-primary group-hover:text-accent transition-colors">{event.home_team}</p>
                    <p className="text-text-muted text-xs my-1">vs</p>
                    <p className="font-semibold text-text-primary group-hover:text-accent transition-colors">{event.away_team}</p>
                </div>
                <div className="text-right">
                    <span className="text-xs text-text-muted">{event.markets_count} markets</span>
                    <div className="mt-2 text-accent text-sm font-medium group-hover:translate-x-1 transition-transform">View →</div>
                </div>
            </div>
        </Link>
    );
}
