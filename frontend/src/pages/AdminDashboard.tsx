import { useQuery } from '@tanstack/react-query';
import { adminAPI } from '../lib/api';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';

interface Stats {
    total_users: number;
    total_bets: number;
    pending_bets: number;
    total_staked: number;
    total_paid_out: number;
    active_events: number;
    total_deposited: number;
}

export default function AdminDashboard() {
    const { data, isLoading } = useQuery({
        queryKey: ['admin-stats'],
        queryFn: () => adminAPI.stats(),
        refetchInterval: 15000,
    });

    const stats: Stats | null = data?.data || null;
    const profit = stats ? stats.total_staked - stats.total_paid_out : 0;

    const cards = stats ? [
        { label: 'Total Users', value: stats.total_users, icon: '👥', color: 'text-blue-400' },
        { label: 'Active Events', value: stats.active_events, icon: '⚽', color: 'text-accent' },
        { label: 'Total Bets', value: stats.total_bets, icon: '🎲', color: 'text-purple-400' },
        { label: 'Pending Bets', value: stats.pending_bets, icon: '⏳', color: 'text-warning' },
        { label: 'Total Staked', value: `$${stats.total_staked.toFixed(2)}`, icon: '💰', color: 'text-accent' },
        { label: 'Total Paid Out', value: `$${stats.total_paid_out.toFixed(2)}`, icon: '💸', color: 'text-danger' },
        { label: 'Net Profit', value: `$${profit.toFixed(2)}`, icon: '📊', color: profit >= 0 ? 'text-accent' : 'text-danger' },
        { label: 'Total Deposited', value: `$${stats.total_deposited.toFixed(2)}`, icon: '🏦', color: 'text-blue-400' },
    ] : [];

    return (
        <div>
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-2xl font-bold">Admin Dashboard</h1>
                    <p className="text-text-muted text-sm">Platform overview & management</p>
                </div>
                <div className="flex gap-2 flex-wrap">
                    <Link to="/admin/events" className="btn-secondary !py-2 !px-4 text-sm">Manage Events</Link>
                    <Link to="/admin/settle" className="btn-secondary !py-2 !px-4 text-sm">Settle Bets</Link>
                    <Link to="/admin/deposit" className="btn-primary !py-2 !px-4 text-sm">💰 Deposit</Link>
                    <Link to="/admin/bet" className="btn-primary !py-2 !px-4 text-sm">🎲 Create Market</Link>
                </div>
            </div>

            {isLoading ? (
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    {[...Array(8)].map((_, i) => <div key={i} className="card animate-pulse h-24" />)}
                </div>
            ) : (
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    {cards.map((card, idx) => (
                        <motion.div
                            key={card.label}
                            initial={{ y: 20, opacity: 0 }}
                            animate={{ y: 0, opacity: 1 }}
                            transition={{ delay: idx * 0.05 }}
                            className="card"
                        >
                            <div className="flex items-center gap-2 mb-2">
                                <span className="text-lg">{card.icon}</span>
                                <span className="text-text-muted text-xs uppercase tracking-wide">{card.label}</span>
                            </div>
                            <p className={`text-2xl font-bold ${card.color}`}>{card.value}</p>
                        </motion.div>
                    ))}
                </div>
            )}

            {/* Quick actions */}
            <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-4">
                <Link to="/admin/events" className="card group hover:border-accent/30 transition-all">
                    <h3 className="font-semibold mb-1 group-hover:text-accent transition-colors">📋 Event Management</h3>
                    <p className="text-text-muted text-sm">Create, edit, and manage events and their statuses</p>
                </Link>
                <Link to="/admin/settle" className="card group hover:border-accent/30 transition-all">
                    <h3 className="font-semibold mb-1 group-hover:text-accent transition-colors">⚖️ Bet Settlement</h3>
                    <p className="text-text-muted text-sm">Settle markets and process bet payouts</p>
                </Link>
                <Link to="/admin/deposit" className="card group hover:border-accent/30 transition-all">
                    <h3 className="font-semibold mb-1 group-hover:text-accent transition-colors">💰 Deposit to User</h3>
                    <p className="text-text-muted text-sm">Add funds to a user account by email</p>
                </Link>
                <Link to="/admin/bet" className="card group hover:border-accent/30 transition-all">
                    <h3 className="font-semibold mb-1 group-hover:text-accent transition-colors">🎲 Create Custom Market</h3>
                    <p className="text-text-muted text-sm">Add custom betting markets to events (e.g., Red Card, Clean Sheet)</p>
                </Link>
            </div>
        </div>
    );
}
