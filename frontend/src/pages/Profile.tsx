import { useQuery } from '@tanstack/react-query';
import { transactionsAPI, authAPI } from '../lib/api';
import { useAuthStore } from '../store';
import { useEffect } from 'react';
import { motion } from 'framer-motion';

interface Transaction {
    id: string; type: string; amount: number; balance_before: number;
    balance_after: number; status: string; created_at: string;
}

const txTypeLabels: Record<string, string> = {
    deposit: '💰 Deposit',
    withdraw: '💸 Withdrawal',
    bet_placed: '🎲 Bet Placed',
    bet_won: '🏆 Bet Won',
    bet_lost: '❌ Bet Lost',
    bet_cancelled: '↩️ Bet Cancelled',
};

export default function ProfilePage() {
    const { user, setAuth } = useAuthStore();

    // Refresh profile on mount
    useEffect(() => {
        authAPI.profile().then(({ data }) => {
            const tokens = localStorage.getItem('betkz_tokens');
            if (tokens) setAuth(data, JSON.parse(tokens));
        }).catch(() => { });
    }, [setAuth]);

    const { data: txData, isLoading } = useQuery({
        queryKey: ['transactions'],
        queryFn: () => transactionsAPI.list({ limit: '30' }),
    });

    const transactions: Transaction[] = txData?.data?.transactions || [];

    return (
        <div>
            <h1 className="text-2xl font-bold mb-6">Profile</h1>

            {/* Balance card */}
            <div className="card mb-6 bg-gradient-to-br from-bg-card to-bg-secondary">
                <p className="text-text-muted text-sm">Total Balance</p>
                <p className="text-4xl font-bold text-accent mt-1">${user?.balance.toFixed(2)}</p>
                <p className="text-text-muted text-sm mt-2">{user?.email}</p>
            </div>

            {/* Transactions */}
            <h2 className="text-lg font-bold mb-4">Transaction History</h2>

            {isLoading ? (
                <div className="space-y-3">{[...Array(5)].map((_, i) => <div key={i} className="card animate-pulse h-16" />)}</div>
            ) : transactions.length === 0 ? (
                <div className="card text-center py-8">
                    <p className="text-text-muted">No transactions yet</p>
                </div>
            ) : (
                <div className="space-y-2">
                    {transactions.map((tx, idx) => (
                        <motion.div
                            key={tx.id}
                            initial={{ x: -10, opacity: 0 }}
                            animate={{ x: 0, opacity: 1 }}
                            transition={{ delay: idx * 0.03 }}
                            className="card !p-3 flex items-center justify-between"
                        >
                            <div>
                                <p className="text-sm font-medium">{txTypeLabels[tx.type] || tx.type}</p>
                                <p className="text-xs text-text-muted">
                                    {new Date(tx.created_at).toLocaleString('en', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                                </p>
                            </div>
                            <div className="text-right">
                                <p className={`font-bold text-sm ${tx.amount > 0 ? 'text-accent' : 'text-danger'}`}>
                                    {tx.amount > 0 ? '+' : ''}{tx.amount.toFixed(2)}
                                </p>
                                <p className="text-xs text-text-muted">${tx.balance_after.toFixed(2)}</p>
                            </div>
                        </motion.div>
                    ))}
                </div>
            )}
        </div>
    );
}
