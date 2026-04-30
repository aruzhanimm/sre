import { useMutation } from '@tanstack/react-query';
import { adminAPI } from '../lib/api';
import { toast } from 'sonner';
import { motion } from 'framer-motion';
import { useState } from 'react';

export default function AdminDeposit() {
    const [email, setEmail] = useState('');
    const [amount, setAmount] = useState('');

    const depositMutation = useMutation({
        mutationFn: () => adminAPI.depositByEmail(email.trim(), parseFloat(amount)),
        onSuccess: (res) => {
            toast.success(`Deposited $${parseFloat(amount).toFixed(2)} to ${res.data.user_email}. New balance: $${res.data.new_balance.toFixed(2)}`);
            setAmount('');
        },
        onError: (err: any) => toast.error(err.response?.data?.error || 'Deposit failed'),
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!email.trim()) { toast.error('Email is required'); return; }
        const amt = parseFloat(amount);
        if (!amt || amt <= 0) { toast.error('Amount must be positive'); return; }
        depositMutation.mutate();
    };

    return (
        <div>
            <motion.div initial={{ y: 10, opacity: 0 }} animate={{ y: 0, opacity: 1 }}>
                <h1 className="text-2xl font-bold mb-2">💰 Deposit to User</h1>
                <p className="text-text-muted text-sm mb-6">Add funds to a user account by their email address</p>
            </motion.div>

            <motion.form
                initial={{ y: 20, opacity: 0 }}
                animate={{ y: 0, opacity: 1 }}
                transition={{ delay: 0.05 }}
                onSubmit={handleSubmit}
                className="card max-w-lg"
            >
                <div className="space-y-4">
                    <div>
                        <label className="block text-sm text-text-secondary mb-2">User Email</label>
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder="user@example.com"
                            className="input-field"
                            required
                        />
                    </div>

                    <div>
                        <label className="block text-sm text-text-secondary mb-2">Amount ($)</label>
                        <div className="relative">
                            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted">$</span>
                            <input
                                type="number"
                                value={amount}
                                onChange={(e) => setAmount(e.target.value)}
                                placeholder="100.00"
                                className="input-field !pl-7"
                                min="0.01"
                                step="0.01"
                                required
                            />
                        </div>
                    </div>

                    {/* Quick amount buttons */}
                    <div className="flex gap-2 flex-wrap">
                        {[50, 100, 250, 500, 1000, 5000].map((v) => (
                            <button
                                key={v}
                                type="button"
                                onClick={() => setAmount(v.toString())}
                                className={`text-xs px-3 py-1.5 rounded-lg border transition-all ${amount === v.toString()
                                        ? 'bg-accent/20 border-accent text-accent'
                                        : 'bg-bg-secondary border-border-light text-text-secondary hover:border-accent/30'
                                    }`}
                            >
                                ${v}
                            </button>
                        ))}
                    </div>
                </div>

                <div className="mt-6 pt-4 border-t border-border-light flex items-center justify-between">
                    <div className="text-sm text-text-muted">
                        {email && amount && parseFloat(amount) > 0 && (
                            <span>
                                Deposit <span className="text-accent font-semibold">${parseFloat(amount).toFixed(2)}</span> to{' '}
                                <span className="text-text-primary font-medium">{email}</span>
                            </span>
                        )}
                    </div>
                    <button
                        type="submit"
                        disabled={depositMutation.isPending}
                        className="btn-primary !py-2.5 !px-8 disabled:opacity-50"
                    >
                        {depositMutation.isPending ? 'Processing...' : 'Confirm Deposit'}
                    </button>
                </div>
            </motion.form>
        </div>
    );
}
