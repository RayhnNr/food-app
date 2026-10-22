import axios from 'axios';

const api = axios.create({
    baseURL: 'http://localhost:8080',
    timeout: 5000,
});

export const productImageUrl = (gambar) => {
    if (!gambar) return '';
    if (/^(https?:|data:|blob:)/i.test(gambar) || gambar.startsWith('/gambar/')) return gambar;

    const path = gambar.startsWith('/uploads/') ? gambar : `/uploads/${gambar}`;
    return `${api.defaults.baseURL}${path}`;
};

export const fallbackProductImage = (event, gambar) => {
    if (!gambar || event.currentTarget.dataset.fallback === 'true') return;

    event.currentTarget.dataset.fallback = 'true';
    event.currentTarget.src = `/gambar/${encodeURIComponent(gambar.split('/').pop())}`;
};

// Pasang token ke setiap request (BARU)
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

api.interceptors.response.use(
    (res) => res,
    (err) => {
        console.error('API error:', err.message);

        // Token kedaluwarsa/invalid -> paksa login ulang (BARU)
        // Dikecualikan untuk /auth/*, karena password salah saat login juga 401
        const isAuthRoute = err.config?.url?.startsWith('/auth/');
        if (err.response?.status === 401 && !isAuthRoute) {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/login';
        }

        return Promise.reject(err);
    },
);

export default api;
