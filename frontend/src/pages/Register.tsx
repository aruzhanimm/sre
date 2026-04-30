import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store';
import { authAPI } from '../lib/api';
import { toast } from 'sonner';
import { motion } from 'framer-motion';

export default function RegisterPage() {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [loading, setLoading] = useState(false);
    const setAuth = useAuthStore((s) => s.setAuth);
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (password !== confirm) {
            toast.error('Passwords do not match');
            return;
        }
        setLoading(true);
        try {
            const { data } = await authAPI.register(email, password);
            setAuth(data.user, data.tokens);
            toast.success('Account created! 🎉');
            navigate('/');
        } catch (err: any) {
            toast.error(err.response?.data?.error || 'Registration failed');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center px-4 bg-bg-primary"
            style={{ backgroundImage: 'radial-gradient(ellipse 80% 50% at 50% -20%, rgba(16, 185, 129, 0.12), transparent)' }}>
            <motion.div initial={{ y: 20, opacity: 0 }} animate={{ y: 0, opacity: 1 }} className="w-full max-w-md">
                <div className="text-center mb-8">
                    <Link to="/" className="inline-flex items-center gap-2 mb-6">
                        <div className="w-10 h-10 rounded-xl bg-accent flex items-center justify-center font-bold text-bg-primary text-xl">B</div>
                        <span className="text-2xl font-bold bg-gradient-to-r from-accent to-emerald-300 bg-clip-text text-transparent">BetKZ</span>
                    </Link>
                    <h1 className="text-2xl font-bold">Create account</h1>
                    <p className="text-text-muted mt-1">Join BetKZ and start betting</p>
                </div>

                <form onSubmit={handleSubmit} className="card space-y-4">
                    <div>
                        <label className="block text-sm text-text-secondary mb-1.5">Email</label>
                        <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} className="input-field" placeholder="you@example.com" required />
                    </div>
                    <div>
                        <label className="block text-sm text-text-secondary mb-1.5">Password</label>
                        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} className="input-field" placeholder="Min 8 chars, upper+lower+digit" required />
                    </div>
                    <div>
                        <label className="block text-sm text-text-secondary mb-1.5">Confirm Password</label>
                        <input type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} className="input-field" placeholder="••••••••" required />
                    </div>
                    <button type="submit" disabled={loading} className="btn-primary w-full disabled:opacity-50">
                        {loading ? 'Creating...' : 'Create Account'}
                    </button>
                </form>

                <p className="text-center mt-4 text-text-muted text-sm">
                    Already have an account? <Link to="/login" className="text-accent hover:text-accent-hover transition-colors">Sign In</Link>
                </p>
            </motion.div>
        </div>
    );
}
