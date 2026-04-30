import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { eventsAPI, adminAPI } from '../lib/api';
import { toast } from 'sonner';
import { motion, AnimatePresence } from 'framer-motion';
import { useState } from 'react';

interface Event { id: string; home_team: string; away_team: string; status: string; sport_name: string; }

const MARKET_PRESETS: Record<string, { name: string; outcomes: string[] }> = {
    'Red Card': { name: 'Red Card', outcomes: ['Yes', 'No'] },
    'First Goal Scorer': { name: 'First Goal Scorer', outcomes: ['Home', 'Away'] },
    'Clean Sheet': { name: 'Clean Sheet', outcomes: ['Yes', 'No'] },
    'Penalty': { name: 'Penalty in Match', outcomes: ['Yes', 'No'] },
    'Corner Kick Over/Under': { name: 'Corner Kicks Over/Under 9.5', outcomes: ['Over', 'Under'] },
    'Half-Time Result': { name: 'Half Time Result', outcomes: ['Home', 'Draw', 'Away'] },
};

export default function AdminBet() {
    const queryClient = useQueryClient();
    const [selectedEvent, setSelectedEvent] = useState('');
    const [marketName, setMarketName] = useState('');
    const marketType = 'custom';
    const [outcomes, setOutcomes] = useState<{ outcome: string; odds: string }[]>([
        { outcome: '', odds: '2.00' },
        { outcome: '', odds: '2.00' },
    ]);

    const { data: eventsData } = useQuery({
        queryKey: ['admin-market-events'],
        queryFn: () => eventsAPI.list({ status: 'upcoming', limit: '50' }),
    });

    const { data: liveEventsData } = useQuery({
        queryKey: ['admin-market-live-events'],
        queryFn: () => eventsAPI.list({ status: 'live', limit: '50' }),
    });

    const upcomingEvents: Event[] = eventsData?.data?.events || [];
    const liveEvents: Event[] = liveEventsData?.data?.events || [];
    const allEvents = [...liveEvents, ...upcomingEvents];

    const createMutation = useMutation({
        mutationFn: () =>
            adminAPI.createMarket({
                event_id: selectedEvent,
                market_type: marketType,
                name: marketName,
                initial_odds: outcomes.map((o) => ({
                    outcome: o.outcome,
                    odds: parseFloat(o.odds),
                })),
            }),
        onSuccess: () => {
            toast.success(`Market "${marketName}" created successfully!`);
            queryClient.invalidateQueries({ queryKey: ['admin-market-events'] });
            setMarketName('');
            setOutcomes([{ outcome: '', odds: '2.00' }, { outcome: '', odds: '2.00' }]);
        },
        onError: (err: any) => toast.error(err.response?.data?.error || 'Failed to create market'),
    });

    const applyPreset = (presetKey: string) => {
        const preset = MARKET_PRESETS[presetKey];
        if (preset) {
            setMarketName(preset.name);
            setOutcomes(preset.outcomes.map((o) => ({ outcome: o, odds: '2.00' })));
        }
    };

    const addOutcome = () => setOutcomes([...outcomes, { outcome: '', odds: '2.00' }]);
    const removeOutcome = (idx: number) => {
        if (outcomes.length > 2) setOutcomes(outcomes.filter((_, i) => i !== idx));
    };
    const updateOutcome = (idx: number, field: 'outcome' | 'odds', value: string) => {
        const updated = [...outcomes];
        updated[idx] = { ...updated[idx], [field]: value };
        setOutcomes(updated);
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!selectedEvent) { toast.error('Select an event'); return; }
        if (!marketName.trim()) { toast.error('Market name is required'); return; }
        for (const o of outcomes) {
            if (!o.outcome.trim()) { toast.error('All outcomes must have a name'); return; }
            const odds = parseFloat(o.odds);
            if (!odds || odds < 1.01) { toast.error(`Odds for "${o.outcome}" must be ≥ 1.01`); return; }
        }
        const names = outcomes.map((o) => o.outcome.trim().toLowerCase());
        if (new Set(names).size !== names.length) { toast.error('Outcome names must be unique'); return; }
        createMutation.mutate();
    };

    return (
        <div>
            <motion.div initial={{ y: 10, opacity: 0 }} animate={{ y: 0, opacity: 1 }}>
                <h1 className="text-2xl font-bold mb-2">🎲 Create Custom Market</h1>
                <p className="text-text-muted text-sm mb-6">Add a new betting market to an event (e.g., Red Card, First Goal Scorer)</p>
            </motion.div>

            <form onSubmit={handleSubmit} className="space-y-4 max-w-2xl">
                {/* Event Selector */}
                <motion.div initial={{ y: 15, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: 0.05 }} className="card">
                    <label className="block text-sm text-text-secondary mb-2">Select Event</label>
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
                </motion.div>

                {/* Presets */}
                <motion.div initial={{ y: 15, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: 0.1 }} className="card">
                    <label className="block text-sm text-text-secondary mb-2">Quick Presets</label>
                    <div className="flex gap-2 flex-wrap">
                        {Object.keys(MARKET_PRESETS).map((key) => (
                            <button
                                key={key}
                                type="button"
                                onClick={() => applyPreset(key)}
                                className={`text-xs px-3 py-1.5 rounded-lg border transition-all ${marketName === MARKET_PRESETS[key].name
                                    ? 'bg-accent/20 border-accent text-accent'
                                    : 'bg-bg-secondary border-border-light text-text-secondary hover:border-accent/30'
                                    }`}
                            >
                                {key}
                            </button>
                        ))}
                    </div>
                </motion.div>

                {/* Market Details */}
                <motion.div initial={{ y: 15, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: 0.15 }} className="card">
                    <label className="block text-sm text-text-secondary mb-2">Market Name</label>
                    <input
                        type="text"
                        value={marketName}
                        onChange={(e) => setMarketName(e.target.value)}
                        placeholder="e.g., Red Card, First Goal Scorer"
                        className="input-field"
                        required
                    />
                </motion.div>

                {/* Outcomes */}
                <motion.div initial={{ y: 15, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: 0.2 }} className="card">
                    <div className="flex items-center justify-between mb-3">
                        <label className="text-sm text-text-secondary">Outcomes & Odds</label>
                        <button
                            type="button"
                            onClick={addOutcome}
                            className="text-xs text-accent hover:text-accent-hover transition-colors font-medium"
                        >
                            + Add Outcome
                        </button>
                    </div>

                    <AnimatePresence>
                        <div className="space-y-2">
                            {outcomes.map((o, idx) => (
                                <motion.div
                                    key={idx}
                                    initial={{ opacity: 0, x: -10 }}
                                    animate={{ opacity: 1, x: 0 }}
                                    exit={{ opacity: 0, x: 10 }}
                                    className="flex items-center gap-2"
                                >
                                    <input
                                        type="text"
                                        value={o.outcome}
                                        onChange={(e) => updateOutcome(idx, 'outcome', e.target.value)}
                                        placeholder={`Outcome ${idx + 1} (e.g., Yes, No, Home)`}
                                        className="input-field flex-1"
                                        required
                                    />
                                    <div className="relative w-28">
                                        <input
                                            type="number"
                                            value={o.odds}
                                            onChange={(e) => updateOutcome(idx, 'odds', e.target.value)}
                                            className="input-field text-right text-accent font-semibold"
                                            min="1.01"
                                            step="0.01"
                                            required
                                        />
                                    </div>
                                    {outcomes.length > 2 && (
                                        <button
                                            type="button"
                                            onClick={() => removeOutcome(idx)}
                                            className="text-text-muted hover:text-danger transition-colors p-1"
                                        >
                                            ✕
                                        </button>
                                    )}
                                </motion.div>
                            ))}
                        </div>
                    </AnimatePresence>
                </motion.div>

                {/* Summary & Submit */}
                {selectedEvent && marketName && (
                    <motion.div initial={{ y: 15, opacity: 0 }} animate={{ y: 0, opacity: 1 }} className="card">
                        <div className="flex items-center justify-between">
                            <div className="text-sm">
                                <p className="text-text-muted mb-1">Creating market:</p>
                                <p className="font-semibold text-text-primary">{marketName}</p>
                                <p className="text-text-muted text-xs mt-1">
                                    {outcomes.filter((o) => o.outcome).map((o) => `${o.outcome} (${o.odds})`).join(' · ')}
                                </p>
                            </div>
                            <button
                                type="submit"
                                disabled={createMutation.isPending}
                                className="btn-primary !py-2.5 !px-8 disabled:opacity-50"
                            >
                                {createMutation.isPending ? 'Creating...' : 'Create Market'}
                            </button>
                        </div>
                    </motion.div>
                )}
            </form>
        </div>
    );
}
