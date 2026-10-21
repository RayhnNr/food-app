import { ref, computed } from 'vue';
import api from './api';

const token = ref(localStorage.getItem('token'));
const user = ref(JSON.parse(localStorage.getItem('user') || 'null'));

function simpanSesi(data) {
    token.value = data.token;
    user.value = data.user;
    localStorage.setItem('token', data.token);
    localStorage.setItem('user', JSON.stringify(data.user));
}

export function useAuth() {
    const isLoggedIn = computed(() => !!token.value);

    async function login(email, password, captchaToken) {
        const res = await api.post('/auth/login', {
            email,
            password,
            captcha_token: captchaToken,
        });
        simpanSesi(res.data);
    }

    async function register(name, email, password, captchaToken) {
        const res = await api.post('/auth/register', { 
            name, 
            email, 
            password,
            captcha_token: captchaToken,
        });
        simpanSesi(res.data); // register sudah langsung auto-login
    }

    function logout() {
        token.value = null;
        user.value = null;
        localStorage.removeItem('token');
        localStorage.removeItem('user');
    }

    return { token, user, isLoggedIn, login, register, logout };
}
