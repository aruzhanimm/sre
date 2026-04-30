import { create } from 'zustand';

interface User {
    id: string;
    email: string;
    balance: number;
    role: string;
}

interface Tokens {
    access_token: string;
    refresh_token: string;
    expires_in: number;
}

interface AuthState {
    user: User | null;
    tokens: Tokens | null;
    isAuthenticated: boolean;
    setAuth: (user: User, tokens: Tokens) => void;
    updateBalance: (balance: number) => void;
    logout: () => void;
    loadFromStorage: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
    user: null,
    tokens: null,
    isAuthenticated: false,

    setAuth: (user, tokens) => {
        localStorage.setItem('betkz_user', JSON.stringify(user));
        localStorage.setItem('betkz_tokens', JSON.stringify(tokens));
        set({ user, tokens, isAuthenticated: true });
    },

    updateBalance: (balance) => {
        set((state) => {
            if (state.user) {
                const updated = { ...state.user, balance };
                localStorage.setItem('betkz_user', JSON.stringify(updated));
                return { user: updated };
            }
            return {};
        });
    },

    logout: () => {
        localStorage.removeItem('betkz_user');
        localStorage.removeItem('betkz_tokens');
        set({ user: null, tokens: null, isAuthenticated: false });
    },

    loadFromStorage: () => {
        const user = localStorage.getItem('betkz_user');
        const tokens = localStorage.getItem('betkz_tokens');
        if (user && tokens) {
            set({
                user: JSON.parse(user),
                tokens: JSON.parse(tokens),
                isAuthenticated: true,
            });
        }
    },
}));

// Betslip store
interface BetSelection {
    marketId: string;
    oddId: string;
    outcome: string;
    odds: number;
    eventName: string;
    marketType: string;
}

interface BetslipState {
    selections: BetSelection[];
    stake: number;
    addSelection: (sel: BetSelection) => void;
    removeSelection: (oddId: string) => void;
    clearSlip: () => void;
    setStake: (stake: number) => void;
    isSelected: (oddId: string) => boolean;
    totalOdds: () => number;
    potentialReturn: () => number;
}

export const useBetslipStore = create<BetslipState>((set, get) => ({
    selections: [],
    stake: 10,

    addSelection: (sel) => {
        set((state) => {
            // Replace if same market
            const filtered = state.selections.filter((s) => s.marketId !== sel.marketId);
            return { selections: [...filtered, sel] };
        });
    },

    removeSelection: (oddId) => {
        set((state) => ({
            selections: state.selections.filter((s) => s.oddId !== oddId),
        }));
    },

    clearSlip: () => set({ selections: [], stake: 10 }),

    setStake: (stake) => set({ stake: Math.max(0.5, Math.min(10000, stake)) }),

    isSelected: (oddId) => get().selections.some((s) => s.oddId === oddId),

    totalOdds: () => get().selections.reduce((acc, s) => acc * s.odds, 1),

    potentialReturn: () => get().stake * get().totalOdds(),
}));
