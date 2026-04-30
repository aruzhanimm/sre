import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { eventsAPI, adminAPI } from '../lib/api';
import { toast } from 'sonner';
import { motion } from 'framer-motion';
import { useState } from 'react';

interface Odd { id: string; outcome: string; current_odds: number; total_stake: number; bet_count: number; }
interface Market { id: string; event_id: string; market_type: string; name: string; status: string; odds: Odd[]; }
interface Event { id: string; home_team: string; away_team: string; status: string; sport_name: string; }

export default function AdminSettle() {
    const queryClient = useQueryClient();
    const [selectedEvent, setSelectedEvent] = useState<string>('');

    const { data: eventsData } = useQuery({
        queryKey: ['settle-events'],
        queryFn: () => eventsAPI.list({ status: 'finished', limit: '50' }),
    });

    const { data: liveEventsData } = useQuery({
        queryKey: ['settle-live-events'],
        queryFn: () => eventsAPI.list({ status: 'live', limit: '50' }),
    });

    const { data: marketsData } = useQuery({
        queryKey: ['settle-markets', selectedEvent],
        queryFn: () => eventsAPI.markets(selectedEvent),
        enabled: !!selectedEvent,
    });

    const settleMutation = useMutation({
        mutationFn: (data: { market_id: string; winning_outcome: string }) => adminAPI.settle(data),
        onSuccess: (res) => {
            toast.success(`Settled ${res.data.settled_count} bets`);
            queryClient.invalidateQueries({ queryKey: ['settle-markets'] });
            queryClient.invalidateQueries({ queryKey: ['admin-stats'] });
        },
        onError: (err: any) => toast.error(err.response?.data?.error || 'Settlement failed'),
    });

    const finishedEvents: Event[] = eventsData?.data?.events || [];
    const liveEvents: Event[] = liveEventsData?.data?.events || [];
    const allEvents = [...liveEvents, ...finishedEvents];
    const markets: Market[] = marketsData?.data?.markets || [];
    const unsettledMarkets = markets.filter((m) => m.status !== 'settled');

    return (
        <div>
            <h1 className="text-2xl font-bold mb-6">Bet Settlement</h1>

            {/* Event selector */}
            <div className="card mb-6">
                <label className="block text-sm text-text-secondary mb-2">Select Event to Settle</label>
                <select
                    value={selectedEvent}
                    onChange={(e) => setSelectedEvent(e.target.value)}
                    className="input-field"
                >
                    <option value="">— Choose an event —</option>
                    {allEvents.map((ev) => (
                        <option key={ev.id} value={ev.id}>
                            [{ev.status.toUpperCase()}] {ev.home_team} vs {ev.away_team}
                        </option>
                    ))}
                </select>
            </div>

            {/* Markets */}
            {selectedEvent && unsettledMarkets.length === 0 && (
                <div className="card text-center py-8">
                    <p className="text-text-muted">All markets are settled for this event ✅</p>
                </div>
            )}

            {unsettledMarkets.map((market, idx) => (
                <motion.div key={market.id} initial={{ y: 10, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: idx * 0.05 }}>
                    <MarketSettleCard market={market} onSettle={(outcome) => settleMutation.mutate({ market_id: market.id, winning_outcome: outcome })} loading={settleMutation.isPending} />
                </motion.div>
            ))}
        </div>
    );
}

function MarketSettleCard({ market, onSettle, loading }: { market: Market; onSettle: (outcome: string) => void; loading: boolean }) {
    const [selected, setSelected] = useState<string>('');

    return (
        <div className="card mb-4">
            <div className="flex items-center justify-between mb-3">
                <div>
                    <h3 className="font-medium">{market.name}</h3>
                    <span className="text-xs text-text-muted">{market.market_type} • {market.status}</span>
                </div>
            </div>

            <p className="text-sm text-text-secondary mb-3">Select the winning outcome:</p>

            <div className="grid gap-2" style={{ gridTemplateColumns: `repeat(${market.odds?.length || 3}, 1fr)` }}>
                {market.odds?.map((odd) => (
                    <button
                        key={odd.id}
                        onClick={() => setSelected(odd.outcome)}
                        className={`rounded-lg p-3 text-center transition-all border ${selected === odd.outcome
                                ? 'bg-accent/20 border-accent text-accent'
                                : 'bg-bg-secondary border-border-light text-text-primary hover:border-accent/30'
                            }`}
                    >
                        <p className="text-xs text-text-muted mb-1">{odd.outcome}</p>
                        <p className="font-bold">{odd.current_odds.toFixed(2)}</p>
                        <p className="text-xs text-text-muted mt-1">{odd.bet_count} bets · ${odd.total_stake.toFixed(2)}</p>
                    </button>
                ))}
            </div>

            {selected && (
                <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="mt-4 pt-3 border-t border-border-light flex items-center justify-between">
                    <p className="text-sm">
                        Winning: <span className="font-bold text-accent">{selected}</span>
                    </p>
                    <button
                        onClick={() => { if (confirm(`Settle "${market.name}" with winner "${selected}"?`)) onSettle(selected); }}
                        disabled={loading}
                        className="btn-primary !py-2 !px-6 text-sm"
                    >
                        {loading ? 'Settling...' : 'Confirm Settlement'}
                    </button>
                </motion.div>
            )}
        </div>
    );
}
