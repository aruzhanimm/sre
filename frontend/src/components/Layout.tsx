import { Outlet, Link, useNavigate } from 'react-router-dom';
import { useAuthStore, useBetslipStore } from '../store';
import { motion, AnimatePresence } from 'framer-motion';
import { betsAPI } from '../lib/api';
import { toast } from 'sonner';
import { useState } from 'react';

export default function Layout() {
    const { user, isAuthenticated, logout } = useAuthStore();
    const navigate = useNavigate();
    const selectionsCount = useBetslipStore((s) => s.selections.length);

    return (
        <div className="min-h-screen flex flex-col">
            <Header user={user} isAuthenticated={isAuthenticated} logout={logout} navigate={navigate} />
            <div className="flex flex-1 max-w-[1440px] mx-auto w-full px-4 gap-4 py-4">
                <main className="flex-1 min-w-0">
                    <Outlet />
                </main>
                <DesktopBetslip />
            </div>
            {/* Mobile betslip */}
            {selectionsCount > 0 && <MobileBetslip />}
        </div>
    );
}

function Header({ user, isAuthenticated, logout, navigate }: any) {
    const [mobileOpen, setMobileOpen] = useState(false);

    return (
        <header className="glass sticky top-0 z-50 border-b border-border/50">
            <div className="max-w-[1440px] mx-auto px-4 h-16 flex items-center justify-between">
                <Link to="/" className="flex items-center gap-2">
                    <div className="w-8 h-8 rounded-lg bg-accent flex items-center justify-center font-bold text-bg-primary text-lg">B</div>
                    <span className="text-xl font-bold bg-gradient-to-r from-accent to-emerald-300 bg-clip-text text-transparent">BetKZ</span>
                </Link>

                {/* Desktop nav */}
                <nav className="hidden md:flex items-center gap-6">
                    <Link to="/" className="text-text-secondary hover:text-text-primary transition-colors text-sm font-medium">Events</Link>
                    {isAuthenticated && (
                        <>
                            <Link to="/bets" className="text-text-secondary hover:text-text-primary transition-colors text-sm font-medium">My Bets</Link>
                            <Link to="/profile" className="text-text-secondary hover:text-text-primary transition-colors text-sm font-medium">Profile</Link>
                            {user?.role === 'admin' && (
                                <Link to="/admin" className="text-warning hover:text-yellow-300 transition-colors text-sm font-medium">⚙ Admin</Link>
                            )}
                        </>
                    )}
                </nav>

                <div className="flex items-center gap-3">
                    {isAuthenticated ? (
                        <>
                            <div className="card !p-2 !px-4 flex items-center gap-2">
                                <span className="text-accent font-semibold">${user?.balance.toFixed(2)}</span>
                            </div>
                            <button onClick={() => { logout(); navigate('/'); }} className="hidden md:block text-text-muted hover:text-danger text-sm transition-colors">Logout</button>
                        </>
                    ) : (
                        <>
                            <Link to="/login" className="btn-secondary !py-2 !px-4 text-sm">Login</Link>
                            <Link to="/register" className="btn-primary !py-2 !px-4 text-sm hidden sm:inline-flex">Register</Link>
                        </>
                    )}
                    {/* Mobile hamburger */}
                    <button onClick={() => setMobileOpen(!mobileOpen)} className="md:hidden text-text-secondary p-1">
                        <svg width="24" height="24" fill="none" stroke="currentColor" strokeWidth="2"><path d={mobileOpen ? "M6 6l12 12M6 18L18 6" : "M4 6h16M4 12h16M4 18h16"} /></svg>
                    </button>
                </div>
            </div>

            {/* Mobile menu */}
            <AnimatePresence>
                {mobileOpen && (
                    <motion.div
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: 'auto', opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        className="md:hidden border-t border-border overflow-hidden"
                    >
                        <nav className="flex flex-col p-4 gap-3">
                            <Link to="/" onClick={() => setMobileOpen(false)} className="text-text-secondary hover:text-text-primary text-sm font-medium py-1">Events</Link>
                            {isAuthenticated && (
                                <>
                                    <Link to="/bets" onClick={() => setMobileOpen(false)} className="text-text-secondary hover:text-text-primary text-sm font-medium py-1">My Bets</Link>
                                    <Link to="/profile" onClick={() => setMobileOpen(false)} className="text-text-secondary hover:text-text-primary text-sm font-medium py-1">Profile</Link>
                                    {user?.role === 'admin' && (
                                        <Link to="/admin" onClick={() => setMobileOpen(false)} className="text-warning text-sm font-medium py-1">⚙ Admin Panel</Link>
                                    )}
                                    <button onClick={() => { logout(); navigate('/'); setMobileOpen(false); }} className="text-left text-danger text-sm font-medium py-1">Logout</button>
                                </>
                            )}
                        </nav>
                    </motion.div>
                )}
            </AnimatePresence>
        </header>
    );
}

function DesktopBetslip() {
    const { selections, stake, setStake, removeSelection, clearSlip, totalOdds, potentialReturn } = useBetslipStore();
    const { isAuthenticated, updateBalance, user } = useAuthStore();
    const [loading, setLoading] = useState(false);

    const handlePlaceBet = async () => {
        if (!isAuthenticated) { toast.error('Please login to place a bet'); return; }
        if (selections.length === 0) return;
        setLoading(true);
        try {
            await betsAPI.place({ stake, selections: selections.map((s) => ({ market_id: s.marketId, odd_id: s.oddId, outcome: s.outcome })) });
            toast.success('Bet placed successfully! 🎉');
            if (user) updateBalance(user.balance - stake);
            clearSlip();
        } catch (err: any) { toast.error(err.response?.data?.error || 'Failed to place bet'); }
        finally { setLoading(false); }
    };

    if (selections.length === 0) return null;

    return (
        <aside className="hidden lg:block w-80 flex-shrink-0">
            <div className="card sticky top-20">
                <div className="flex items-center justify-between mb-4">
                    <h3 className="font-bold text-lg">Betslip</h3>
                    <span className="text-xs bg-accent/20 text-accent px-2 py-0.5 rounded-full font-medium">{selections.length}</span>
                </div>
                <div className="space-y-3 max-h-64 overflow-y-auto">
                    {selections.map((sel) => (
                        <div key={sel.oddId} className="bg-bg-secondary rounded-lg p-3 relative group">
                            <button onClick={() => removeSelection(sel.oddId)} className="absolute top-2 right-2 text-text-muted hover:text-danger opacity-0 group-hover:opacity-100 transition-opacity">✕</button>
                            <p className="text-xs text-text-muted">{sel.eventName}</p>
                            <p className="text-sm font-medium mt-1">{sel.outcome}</p>
                            <p className="text-accent font-bold text-sm mt-1">{sel.odds.toFixed(2)}</p>
                        </div>
                    ))}
                </div>
                <div className="mt-4 pt-4 border-t border-border-light">
                    <div className="flex items-center gap-2 mb-3">
                        <span className="text-text-secondary text-sm">Stake:</span>
                        <div className="relative flex-1">
                            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-sm">$</span>
                            <input type="number" value={stake} onChange={(e) => setStake(Number(e.target.value))} className="input-field !pl-7 text-right text-sm" min={0.5} max={10000} step={5} />
                        </div>
                    </div>
                    <div className="flex justify-between text-sm mb-1">
                        <span className="text-text-secondary">Total Odds:</span>
                        <span className="font-semibold">{totalOdds().toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between text-sm mb-4">
                        <span className="text-text-secondary">Potential Return:</span>
                        <span className="font-bold text-accent">${potentialReturn().toFixed(2)}</span>
                    </div>
                    <button onClick={handlePlaceBet} disabled={loading} className="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed">
                        {loading ? 'Placing...' : `Place Bet — $${stake.toFixed(2)}`}
                    </button>
                    <button onClick={clearSlip} className="text-text-muted hover:text-danger text-xs mt-2 w-full text-center transition-colors">Clear Betslip</button>
                </div>
            </div>
        </aside>
    );
}

function MobileBetslip() {
    const { selections, stake, setStake, clearSlip, totalOdds, potentialReturn } = useBetslipStore();
    const { isAuthenticated, updateBalance, user } = useAuthStore();
    const [expanded, setExpanded] = useState(false);
    const [loading, setLoading] = useState(false);

    const handlePlaceBet = async () => {
        if (!isAuthenticated) { toast.error('Please login to place a bet'); return; }
        setLoading(true);
        try {
            await betsAPI.place({ stake, selections: selections.map((s) => ({ market_id: s.marketId, odd_id: s.oddId, outcome: s.outcome })) });
            toast.success('Bet placed! 🎉');
            if (user) updateBalance(user.balance - stake);
            clearSlip();
        } catch (err: any) { toast.error(err.response?.data?.error || 'Failed'); }
        finally { setLoading(false); }
    };

    return (
        <div className="lg:hidden fixed bottom-0 left-0 right-0 z-50 glass border-t border-border">
            <button onClick={() => setExpanded(!expanded)} className="w-full p-3 flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <span className="bg-accent text-bg-primary text-xs font-bold w-6 h-6 rounded-full flex items-center justify-center">{selections.length}</span>
                    <span className="font-semibold text-sm">Betslip</span>
                </div>
                <span className="text-accent font-bold text-sm">${potentialReturn().toFixed(2)}</span>
            </button>
            <AnimatePresence>
                {expanded && (
                    <motion.div initial={{ height: 0 }} animate={{ height: 'auto' }} exit={{ height: 0 }} className="overflow-hidden border-t border-border-light">
                        <div className="p-4 space-y-3">
                            <div className="flex items-center gap-2">
                                <span className="text-text-secondary text-sm">Stake: $</span>
                                <input type="number" value={stake} onChange={(e) => setStake(Number(e.target.value))} className="input-field !py-1.5 text-sm flex-1" min={0.5} max={10000} />
                            </div>
                            <div className="flex justify-between text-sm">
                                <span className="text-text-muted">Odds: {totalOdds().toFixed(2)}</span>
                                <span className="text-accent font-bold">Return: ${potentialReturn().toFixed(2)}</span>
                            </div>
                            <div className="flex gap-2">
                                <button onClick={handlePlaceBet} disabled={loading} className="btn-primary flex-1 !py-2 text-sm">{loading ? 'Placing...' : 'Place Bet'}</button>
                                <button onClick={clearSlip} className="btn-danger !py-2 text-sm">Clear</button>
                            </div>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}
