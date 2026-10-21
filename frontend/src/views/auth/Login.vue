<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuth } from '@/auth';

const route = useRoute();
const router = useRouter();
const { login, user } = useAuth();

const email = ref('');
const password = ref('');
const showPassword = ref(false);
const error = ref('');
const loading = ref(false);

// Captcha
const captchaToken = ref('');
const captchaEl = ref(null);
let widgetId = null;
let timer = null;

onMounted(() => {
    // tunggu script Google selesai dimuat
    timer = setInterval(() => {
        if (window.grecaptcha && window.grecaptcha.render) {
            clearInterval(timer);
            widgetId = window.grecaptcha.render(captchaEl.value, {
                sitekey: import.meta.env.VITE_RECAPTCHA_SITEKEY,
                callback: (token) => (captchaToken.value = token),
                'expired-callback': () => (captchaToken.value = ''),
            });
        }
    }, 100);
});

onUnmounted(() => clearInterval(timer));

function resetCaptcha() {
    if (widgetId !== null) window.grecaptcha.reset(widgetId);
    captchaToken.value = '';
}

async function handleLogin() {
    if (loading.value) return;
    error.value = '';

    if (!captchaToken.value) {
        error.value = 'Centang captcha terlebih dahulu.';
        return;
    }

    loading.value = true;
    try {
        await login(email.value.trim(), password.value, captchaToken.value);

        const tujuan =
            user.value?.role === 'admin' ? '/admin/dashboard' : route.query.redirect || '/';
        await router.push(tujuan);
    } catch (err) {
        resetCaptcha(); // token hanya sekali pakai
        if (!err.response) {
            error.value = 'Tidak bisa terhubung ke server. Pastikan backend sudah berjalan.';
        } else {
            error.value =
                err.response.data?.errors?.captcha ||
                err.response.data?.message ||
                'Login gagal. Coba lagi.';
        }
    } finally {
        loading.value = false;
    }
}
</script>

<template>
    <main class="grid min-h-screen grid-cols-1 grid-rows-[auto_1fr] text-[#1b2420] md:grid-cols-[5fr_6fr] md:grid-rows-1">
        <!-- Panel kiri -->
        <aside aria-hidden="true" class="flex flex-col justify-center gap-3 bg-[#12372a] px-5 py-6 text-white md:gap-8 md:px-12 md:py-12 lg:px-16">

            <h2 class="max-w-[14ch] text-3xl font-extrabold leading-[1.05] tracking-tight md:text-5xl lg:text-6xl">
                Lapar? Pilih, pesan, tinggal makan.
            </h2>
        </aside>

        <!-- Panel kanan -->
        <section class="grid place-items-center bg-[#f7f8f5] px-5 py-10">
            <form
                class="flex w-full max-w-sm flex-col gap-5"
                novalidate
                @submit.prevent="handleLogin"
            >
                <div>
                    <h1 class="text-3xl font-extrabold leading-tight tracking-tight">
                        Masuk ke akunmu
                    </h1>
                    <p class="mt-2 text-[#5d6b64]">
                        Belum punya akun?
                        <RouterLink
                            to="/register"
                            class="font-semibold text-[#12372a] underline underline-offset-4 hover:text-[#0b2219] focus-visible:outline-3 focus-visible:outline-[#f2b705]"
                        >
                            Daftar sekarang
                        </RouterLink>
                    </p>
                </div>

                <p
                    v-if="error"
                    role="alert"
                    class="border-l-4 border-[#b3261e] bg-[#fbeceb] px-4 py-3 text-sm leading-snug text-[#7a1b16]"
                >
                    {{ error }}
                </p>

                <!-- Email -->
                <div class="flex flex-col gap-1.5">
                    <label for="email" class="text-sm font-semibold">Email</label>
                    <input
                        id="email"
                        v-model="email"
                        type="email"
                        inputmode="email"
                        autocomplete="email"
                        placeholder="nama@email.com"
                        required
                        class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 text-base placeholder:text-[#94a09a] hover:border-[#a9b5ae] focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                    />
                </div>

                <!-- Password -->
                <div class="flex flex-col gap-1.5">
                    <label for="password" class="text-sm font-semibold">Password</label>
                    <div class="relative">
                        <input
                            id="password"
                            v-model="password"
                            :type="showPassword ? 'text' : 'password'"
                            autocomplete="current-password"
                            placeholder="Masukkan password"
                            required
                            class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white py-3 pl-3.5 pr-28 text-base placeholder:text-[#94a09a] hover:border-[#a9b5ae] focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                        />
                        <button
                            type="button"
                            :aria-pressed="showPassword"
                            class="absolute right-2 top-1/2 -translate-y-1/2 rounded px-2.5 py-1 text-sm font-semibold text-[#12372a] hover:bg-[#e7ece8] focus-visible:outline-3 focus-visible:outline-[#f2b705]"
                            @click="showPassword = !showPassword"
                        >
                            {{ showPassword ? 'Sembunyikan' : 'Tampilkan' }}
                        </button>
                    </div>
                </div>
                <div ref="captchaEl"></div>

                <!-- Tombol -->
                <button
                    type="submit"
                    :disabled="loading || !email || !password"
                    class="mt-1 inline-flex items-center justify-center gap-2.5 rounded-md bg-[#f2b705] px-4 py-3.5 text-lg font-extrabold text-[#12372a] transition-colors hover:bg-[#ffc61f] focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-[#12372a] disabled:cursor-not-allowed disabled:bg-[#e4e8e5] disabled:text-[#8a968f]"
                >
                    <span
                        v-if="loading"
                        aria-hidden="true"
                        class="size-4 animate-spin rounded-full border-[2.5px] border-current border-r-transparent motion-reduce:[animation-duration:2s]"
                    ></span>
                    {{ loading ? 'Memproses…' : 'Masuk' }}
                </button>
            </form>
        </section>
    </main>
</template>
