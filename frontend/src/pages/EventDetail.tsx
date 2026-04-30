import { useParams } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { eventsAPI } from '../lib/api';
import { useBetslipStore } from '../store';
import { useWebSocket } from '../hooks/useWebSocket';
import { motion, AnimatePresence } from 'framer-motion';
import { useCallback, useState } from 'react';

interface Odd { id: string; market_id: string; outcome: string; current_odds: number; initial_odds: number; total_stake: number; bet_count: number; }
interface Market { id: string; event_id: string; market_type: string; name: string; line: number | null; status: string; odds: Odd[]; }
interface Event { id: string; home_team: string; away_team: string; start_time: string; status: string; sport_name: string; sport_icon: string; }

export default function EventDetailPage() {
    const { id } = useParams<{ id: string }>();
    const queryClient = useQueryClient();
    const { addSelection, isSelected } = useBetslipStore();
    const [oddsChanges, setOddsChanges] = useState<Record<string, 'up' | 'down'>>({});

    const { data: eventData } = useQuery({
        queryKey: ['event', id],
        queryFn: () => eventsAPI.get(id!),
        enabled: !!id,
    });

    const { data: marketsData } = useQuery({
        queryKey: ['markets', id],
        queryFn: () => eventsAPI.markets(id!),
        enabled: !!id,
        refetchInterval: 30000,
    });

    const event: Event | null = eventData?.data || null;
    const markets: Market[] = marketsData?.data?.markets || [];

    // WebSocket for live odds
    const handleOddsUpdate = useCallback((data: any) => {
        if (data.type === 'odds_update' && data.data) {
            const { marketId, outcome, direction } = data.data;
            setOddsChanges((prev) => ({ ...prev, [`${marketId}_${outcome}`]: direction }));
            setTimeout(() => setOddsChanges((prev) => {
                const next = { ...prev };
                delete next[`${marketId}_${outcome}`];
                return next;
            }), 2000);
            queryClient.invalidateQueries({ queryKey: ['markets', id] });
        }
    }, [queryClient, id]);

    useWebSocket(id || null, handleOddsUpdate);

    if (!event) {
        return <div className="card animate-pulse h-64" />;
    }

    const startTime = new Date(event.start_time);

    return (
        <div>
            {/* Event header */}
            <div className="card mb-6">
                <div className="flex items-center gap-2 text-text-muted text-sm mb-4">
                    <span>{event.sport_icon}</span> {event.sport_name}
                    {event.status === 'live' && (
                        <span className="ml-auto text-xs font-medium px-2 py-0.5 rounded-full bg-danger/20 text-danger animate-pulse">● LIVE</span>
                    )}
                </div>

                <div className="flex items-center justify-center gap-8">
                    <div className="text-center flex-1">
                        <div className="w-16 h-16 rounded-2xl bg-bg-secondary flex items-center justify-center mx-auto mb-3">
                            <span className="text-2xl font-bold text-accent">{event.home_team[0]}</span>
                        </div>
                        <h2 className="font-bold text-lg">{event.home_team}</h2>
                    </div>

                    <div className="text-center px-4">
                        <p className="text-3xl font-bold text-text-muted">VS</p>
                        <p className="text-xs text-text-muted mt-2">
                            {startTime.toLocaleDateString('en', { weekday: 'short', month: 'short', day: 'numeric' })}
                        </p>
                        <p className="text-xs text-text-muted">
                            {startTime.toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' })}
                        </p>
                    </div>

                    <div className="text-center flex-1">
                        <div className="w-16 h-16 rounded-2xl bg-bg-secondary flex items-center justify-center mx-auto mb-3">
                            <span className="text-2xl font-bold text-accent">{event.away_team[0]}</span>
                        </div>
                        <h2 className="font-bold text-lg">{event.away_team}</h2>
                    </div>
                </div>
            </div>

            {/* Markets */}
            {markets.length === 0 ? (
                <div className="card text-center py-8">
                    <p className="text-text-muted">No markets available yet</p>
                </div>
            ) : (
                <div className="space-y-4">
                    {markets.map((market) => (
                        <MarketCard
                            key={market.id}
                            market={market}
                            event={event}
                            addSelection={addSelection}
                            isSelected={isSelected}
                            oddsChanges={oddsChanges}
                        />
                    ))}
                </div>
            )}
        </div>
    );
}

function MarketCard({
    market, event, addSelection, isSelected, oddsChanges,
}: {
    market: Market; event: Event; addSelection: any; isSelected: (id: string) => boolean;
    oddsChanges: Record<string, 'up' | 'down'>;
}) {
    const outcomeLabels: Record<string, Record<string, string>> = {
        '1x2': { home: event.home_team, draw: 'Draw', away: event.away_team },
        'over_under': { over: `Over ${market.line}`, under: `Under ${market.line}` },
        'both_teams_score': { yes: 'Yes', no: 'No' },
    };

    const getLabel = (outcome: string) => outcomeLabels[market.market_type]?.[outcome] || outcome;

    return (
        <div className="card">
            <div className="flex items-center justify-between mb-3">
                <h3 className="font-medium text-sm">{market.name}</h3>
                {market.status === 'locked' && (
                    <span className="text-xs text-warning bg-warning/10 px-2 py-0.5 rounded-full">Locked</span>
                )}
            </div>

            <div className="grid gap-2" style={{ gridTemplateColumns: `repeat(${market.odds?.length || 3}, 1fr)` }}>
                <AnimatePresence>
                    {market.odds?.map((odd) => {
                        const selected = isSelected(odd.id);
                        const changeKey = `${market.id}_${odd.outcome}`;
                        const change = oddsChanges[changeKey];
                        const disabled = market.status !== 'open';

                        return (
                            <motion.button
                                key={odd.id}
                                layout
                                onClick={() => {
                                    if (disabled) return;
                                    addSelection({
                                        marketId: market.id,
                                        oddId: odd.id,
                                        outcome: odd.outcome,
                                        odds: odd.current_odds,
                                        eventName: `${event.home_team} vs ${event.away_team}`,
                                        marketType: market.market_type,
                                    });
                                }}
                                disabled={disabled}
                                className={`${selected ? 'odds-badge-selected' : 'odds-badge-default'} ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
                  ${change === 'up' ? '!border-odds-up !text-odds-up' : ''}
                  ${change === 'down' ? '!border-odds-down !text-odds-down' : ''}
                  flex flex-col items-center py-3`}
                            >
                                <span className="text-xs text-text-muted mb-1">{getLabel(odd.outcome)}</span>
                                <span className="font-bold text-base">
                                    {change === 'up' && '▲ '}
                                    {change === 'down' && '▼ '}
                                    {odd.current_odds.toFixed(2)}
                                </span>
                            </motion.button>
                        );
                    })}
                </AnimatePresence>
            </div>
        </div>
    );
}
