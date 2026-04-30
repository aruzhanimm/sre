import axios from 'axios';

const api = axios.create({
    baseURL: '/api',
    headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use((config) => {
    const tokens = localStorage.getItem('betkz_tokens');
    if (tokens) {
        const { access_token } = JSON.parse(tokens);
        config.headers.Authorization = `Bearer ${access_token}`;
    }
    return config;
});

api.interceptors.response.use(
    (res) => res,
    async (error) => {
        const original = error.config;
        if (error.response?.status === 401 && !original._retry) {
            original._retry = true;
            try {
                const tokens = localStorage.getItem('betkz_tokens');
                if (tokens) {
                    const { refresh_token } = JSON.parse(tokens);
                    const { data } = await axios.post('/api/auth/refresh', { refresh_token });
                    localStorage.setItem('betkz_tokens', JSON.stringify(data));
                    original.headers.Authorization = `Bearer ${data.access_token}`;
                    return api(original);
                }
            } catch {
                localStorage.removeItem('betkz_tokens');
                localStorage.removeItem('betkz_user');
                window.location.href = '/login';
            }
        }
        return Promise.reject(error);
    }
);

export default api;

// Auth
export const authAPI = {
    register: (email: string, password: string) => api.post('/auth/register', { email, password }),
    login: (email: string, password: string) => api.post('/auth/login', { email, password }),
    refresh: (refresh_token: string) => api.post('/auth/refresh', { refresh_token }),
    profile: () => api.get('/auth/profile'),
    logout: () => api.post('/auth/logout'),
};

// Sports & Events
export const sportsAPI = {
    list: () => api.get('/sports'),
};

export const eventsAPI = {
    list: (params?: Record<string, string>) => api.get('/events', { params }),
    get: (id: string) => api.get(`/events/${id}`),
    markets: (id: string) => api.get(`/events/${id}/markets`),
};

// Bets
export const betsAPI = {
    place: (data: { stake: number; selections: Array<{ market_id: string; odd_id: string; outcome: string }> }) =>
        api.post('/bets', data),
    list: (params?: Record<string, string>) => api.get('/bets', { params }),
    get: (id: string) => api.get(`/bets/${id}`),
};

// Transactions
export const transactionsAPI = {
    list: (params?: Record<string, string>) => api.get('/transactions', { params }),
};

// Admin
export const adminAPI = {
    stats: () => api.get('/admin/stats'),
    listBets: (params?: Record<string, string>) => api.get('/admin/bets', { params }),
    settle: (data: { market_id: string; winning_outcome: string }) => api.post('/admin/settle', data),
    createEvent: (data: any) => api.post('/admin/events', data),
    updateEvent: (id: string, data: any) => api.put(`/admin/events/${id}`, data),
    updateEventStatus: (id: string, status: string) => api.patch(`/admin/events/${id}/status`, { status }),
    deleteEvent: (id: string) => api.delete(`/admin/events/${id}`),
    createMarket: (data: any) => api.post('/admin/markets', data),
    overrideOdds: (id: string, newOdds: number) => api.put(`/admin/odds/${id}`, { new_odds: newOdds }),
    deposit: (userId: string, amount: number) => api.post(`/admin/users/${userId}/deposit`, { amount }),
    depositByEmail: (email: string, amount: number) => api.post('/admin/deposit-by-email', { email, amount }),
};
