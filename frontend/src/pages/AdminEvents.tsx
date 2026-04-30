import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { eventsAPI, adminAPI, sportsAPI } from '../lib/api';
import { toast } from 'sonner';
import { motion } from 'framer-motion';
import { useState } from 'react';

interface Event {
    id: string; sport_id: number; home_team: string; away_team: string;
    start_time: string; status: string; sport_name: string;
}

export default function AdminEvents() {
    const queryClient = useQueryClient();
    const [showCreate, setShowCreate] = useState(false);

    const { data, isLoading } = useQuery({
        queryKey: ['admin-events'],
        queryFn: () => eventsAPI.list({ limit: '100' }),
    });

    const events: Event[] = data?.data?.events || [];

    const statusMutation = useMutation({
        mutationFn: ({ id, status }: { id: string; status: string }) => adminAPI.updateEventStatus(id, status),
        onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['admin-events'] }); toast.success('Status updated'); },
        onError: (err: any) => toast.error(err.response?.data?.error || 'Failed to update status'),
    });

    const deleteMutation = useMutation({
        mutationFn: (id: string) => adminAPI.deleteEvent(id),
        onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['admin-events'] }); toast.success('Event deleted'); },
        onError: (err: any) => toast.error(err.response?.data?.error || 'Failed to delete'),
    });

    const nextStatus: Record<string, string> = { upcoming: 'live', live: 'finished' };

    return (
        <div>
            <div className="flex items-center justify-between mb-6">
                <h1 className="text-2xl font-bold">Event Management</h1>
                <button onClick={() => setShowCreate(!showCreate)} className="btn-primary !py-2 !px-4 text-sm">
                    {showCreate ? 'Close' : '+ New Event'}
                </button>
            </div>

            {showCreate && <CreateEventForm onClose={() => setShowCreate(false)} />}

            {isLoading ? (
                <div className="space-y-3">{[...Array(5)].map((_, i) => <div key={i} className="card animate-pulse h-20" />)}</div>
            ) : events.length === 0 ? (
                <div className="card text-center py-12">
                    <p className="text-text-muted">No events found</p>
                </div>
            ) : (
                <div className="space-y-3">
                    {events.map((event, idx) => (
                        <motion.div key={event.id} initial={{ y: 10, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: idx * 0.03 }}>
                            <div className="card flex items-center justify-between">
                                <div className="flex-1">
                                    <div className="flex items-center gap-2 mb-1">
                                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${event.status === 'live' ? 'bg-danger/20 text-danger' :
                                            event.status === 'finished' ? 'bg-text-muted/20 text-text-muted' :
                                                'bg-accent/20 text-accent'
                                            }`}>{event.status.toUpperCase()}</span>
                                        <span className="text-xs text-text-muted">{event.sport_name}</span>
                                    </div>
                                    <p className="font-medium">{event.home_team} <span className="text-text-muted">vs</span> {event.away_team}</p>
                                    <p className="text-xs text-text-muted mt-0.5">
                                        {new Date(event.start_time).toLocaleString('en', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                                    </p>
                                </div>
                                <div className="flex items-center gap-2">
                                    {nextStatus[event.status] && (
                                        <button
                                            onClick={() => statusMutation.mutate({ id: event.id, status: nextStatus[event.status] })}
                                            className="btn-secondary !py-1.5 !px-3 text-xs"
                                        >
                                            → {nextStatus[event.status]}
                                        </button>
                                    )}
                                    {event.status === 'upcoming' && (
                                        <button
                                            onClick={() => { if (confirm('Delete this event?')) deleteMutation.mutate(event.id); }}
                                            className="btn-danger !py-1.5 !px-3 text-xs"
                                        >Delete</button>
                                    )}
                                </div>
                            </div>
                        </motion.div>
                    ))}
                </div>
            )}
        </div>
    );
}

function CreateEventForm({ onClose }: { onClose: () => void }) {
    const queryClient = useQueryClient();
    const { data: sportsData } = useQuery({
        queryKey: ['sports'],
        queryFn: sportsAPI.list,
    });
    const sports = (sportsData as any)?.data?.sports || [];

    const [form, setForm] = useState({
        sport_id: 1, home_team: '', away_team: '', start_time: '',
    });

    // Set default sport_id when sports are loaded
    if (form.sport_id === 1 && sports.length > 0 && !sports.find((s: any) => s.id === 1)) {
        setForm(prev => ({ ...prev, sport_id: sports[0].id }));
    }

    const mutation = useMutation({
        mutationFn: (data: typeof form) => adminAPI.createEvent({
            ...data,
            start_time: new Date(data.start_time).toISOString(),
        }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['admin-events'] });
            toast.success('Event created!');
            onClose();
        },
        onError: (err: any) => toast.error(err.response?.data?.error || 'Failed to create event'),
    });

    return (
        <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: 'auto', opacity: 1 }} className="card mb-6">
            <h3 className="font-semibold mb-4">Create New Event</h3>
            <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(form); }} className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <label className="block text-xs text-text-secondary mb-1">Sport</label>
                    <select
                        value={form.sport_id}
                        onChange={(e) => setForm({ ...form, sport_id: Number(e.target.value) })}
                        className="input-field"
                        required
                    >
                        {sports.map((sport: any) => (
                            <option key={sport.id} value={sport.id}>
                                {sport.icon} {sport.name}
                            </option>
                        ))}
                    </select>
                </div>
                <div>
                    <label className="block text-xs text-text-secondary mb-1">Start Time</label>
                    <input type="datetime-local" value={form.start_time} onChange={(e) => setForm({ ...form, start_time: e.target.value })}
                        className="input-field" required />
                </div>
                <div>
                    <label className="block text-xs text-text-secondary mb-1">Home Team</label>
                    <input type="text" value={form.home_team} onChange={(e) => setForm({ ...form, home_team: e.target.value })}
                        className="input-field" placeholder="e.g. Real Madrid" required />
                </div>
                <div>
                    <label className="block text-xs text-text-secondary mb-1">Away Team</label>
                    <input type="text" value={form.away_team} onChange={(e) => setForm({ ...form, away_team: e.target.value })}
                        className="input-field" placeholder="e.g. Barcelona" required />
                </div>
                <div className="md:col-span-2 flex gap-2">
                    <button type="submit" disabled={mutation.isPending} className="btn-primary !py-2 text-sm">
                        {mutation.isPending ? 'Creating...' : 'Create Event'}
                    </button>
                    <button type="button" onClick={onClose} className="btn-secondary !py-2 text-sm">Cancel</button>
                </div>
            </form>
        </motion.div>
    );
}
