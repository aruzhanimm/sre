import { useQuery } from '@tanstack/react-query';
import { betsAPI } from '../lib/api';
import { motion } from 'framer-motion';
import { useState } from 'react';

interface BetLeg {
    outcome: string; locked_odd_value: number; result: string;
    event_name: string; market_type: string;
}
interface Bet {
    id: string; bet_type: string; stake: number; potential_return: number;
    actual_return: number; status: string; placed_at: string; legs: BetLeg[];
}

const statusColors: Record<string, string> = {
    pending: 'bg-warning/20 text-warning',
    won: 'bg-accent/20 text-accent',
    lost: 'bg-danger/20 text-danger',
    cancelled: 'bg-text-muted/20 text-text-muted',
};

export default function BetHistoryPage() {
    const [filter, setFilter] = useState<string>('');

    const { data, isLoading } = useQuery({
        queryKey: ['bets', filter],
        queryFn: () => betsAPI.list({ ...(filter ? { status: filter } : {}), limit: '50' }),
    });

    const bets: Bet[] = data?.data?.bets || [];

    return (
        <div>
            <h1 className="text-2xl font-bold mb-6">My Bets</h1>

            <div className="flex gap-2 mb-6">
                {['', 'pending', 'won', 'lost'].map((s) => (
                    <button
                        key={s}
                        onClick={() => setFilter(s)}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${filter === s ? 'bg-accent text-bg-primary' : 'bg-bg-card text-text-secondary hover:text-text-primary border border-border-light'
                            }`}
                    >
                        {s === '' ? 'All' : s.charAt(0).toUpperCase() + s.slice(1)}
                    </button>
                ))}
            </div>

            {isLoading ? (
                <div className="space-y-3">{[...Array(4)].map((_, i) => <div key={i} className="card animate-pulse h-28" />)}</div>
            ) : bets.length === 0 ? (
                <div className="card text-center py-12">
                    <p className="text-text-muted text-lg">No bets found</p>
                    <p className="text-text-muted text-sm mt-1">Place your first bet to see it here</p>
                </div>
            ) : (
                <div className="space-y-3">
                    {bets.map((bet, idx) => (
                        <motion.div key={bet.id} initial={{ y: 10, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: idx * 0.03 }}>
                            <BetCard bet={bet} />
                        </motion.div>
                    ))}
                </div>
            )}
        </div>
    );
}

function BetCard({ bet }: { bet: Bet }) {
    const [expanded, setExpanded] = useState(false);

    return (
        <div className="card cursor-pointer" onClick={() => setExpanded(!expanded)}>
            <div className="flex items-center justify-between">
                <div>
                    <div className="flex items-center gap-2">
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${statusColors[bet.status]}`}>
                            {bet.status.toUpperCase()}
                        </span>
                        <span className="text-xs text-text-muted uppercase font-medium">{bet.bet_type}</span>
                    </div>
                    <p className="text-sm text-text-muted mt-1">
                        {new Date(bet.placed_at).toLocaleString('en', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                    </p>
                </div>
                <div className="text-right">
                    <p className="text-sm text-text-secondary">Stake: <span className="font-semibold text-text-primary">${bet.stake.toFixed(2)}</span></p>
                    <p className="text-sm">
                        {bet.status === 'won' ? (
                            <span className="text-accent font-bold">+${bet.actual_return.toFixed(2)}</span>
                        ) : bet.status === 'pending' ? (
                            <span className="text-text-secondary">Returns: <span className="text-warning font-semibold">${bet.potential_return.toFixed(2)}</span></span>
                        ) : (
                            <span className="text-danger font-semibold">-${bet.stake.toFixed(2)}</span>
                        )}
                    </p>
                </div>
            </div>

            {expanded && bet.legs && (
                <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: 'auto', opacity: 1 }} className="mt-4 pt-3 border-t border-border-light space-y-2">
                    {bet.legs.map((leg, i) => (
                        <div key={i} className="flex items-center justify-between text-sm bg-bg-secondary rounded-lg px-3 py-2">
                            <div>
                                <p className="text-text-muted text-xs">{leg.event_name}</p>
                                <p className="font-medium">{leg.outcome}</p>
                            </div>
                            <div className="text-right">
                                <p className="font-semibold">{leg.locked_odd_value.toFixed(2)}</p>
                                <span className={`text-xs px-1.5 py-0.5 rounded ${statusColors[leg.result] || 'text-text-muted'}`}>
                                    {leg.result}
                                </span>
                            </div>
                        </div>
                    ))}
                </motion.div>
            )}
        </div>
    );
}
