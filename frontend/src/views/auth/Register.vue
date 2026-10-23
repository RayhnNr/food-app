<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuth } from '@/auth';

const router = useRouter();
const { register, user } = useAuth();

const name = ref('');
const username = ref(''); // Tambahan field username
const email = ref('');
const phone = ref('');    // Tambahan field phone
const password = ref('');
const error = ref('');
const loading = ref(false);

// Captcha
const captchaToken = ref('');
const captchaEl = ref(null);
let widgetId = null;
let timer = null;

onMounted(() => {
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

async function handleRegister() {
    if (loading.value) return;
    error.value = '';

    if (!captchaToken.value) {
        error.value = 'Centang captcha terlebih dahulu.';
        return;
    }

    loading.value = true;

    try {
        await register(
            name.value.trim(),
            username.value.trim().toLowerCase(),
            email.value.trim(),
            phone.value.trim(),
            password.value,
            captchaToken.value,
        );
        await router.push(user.value?.role === 'admin' ? '/admin/dashboard' : '/');
    } catch (err) {
        resetCaptcha();
        error.value =
            err.response?.data?.errors?.captcha ||
            err.response?.data?.message ||
            'Registrasi gagal. Coba lagi.';
    } finally {
        loading.value = false;
    }
}
</script>

<template>
    <main
        class="grid min-h-screen grid-cols-1 grid-rows-[auto_1fr] text-[#1b2420] md:grid-cols-[5fr_6fr] md:grid-rows-1"
    >
        <aside
            aria-hidden="true"
            class="flex flex-col justify-center gap-3 bg-[#12372a] px-5 py-6 text-white md:gap-8 md:px-12 md:py-12 lg:px-16"
        >
            <h2 class="max-w-[14ch] text-3xl font-extrabold leading-[1.05] md:text-5xl lg:text-6xl">
                Mulai pilih menu favoritmu.
            </h2>
        </aside>

        <section class="grid place-items-center bg-[#f7f8f5] px-5 py-10">
            <form
                class="flex w-full max-w-sm flex-col gap-5"
                novalidate
                @submit.prevent="handleRegister"
            >
                <div>
                    <h1 class="text-3xl font-extrabold leading-tight">Buat akun</h1>
                    <p class="mt-2 text-[#5d6b64]">
                        Sudah punya akun?
                        <RouterLink
                            to="/login"
                            class="font-semibold text-[#12372a] underline underline-offset-4 hover:text-[#0b2219]"
                        >
                            Masuk
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

                <div class="flex flex-col gap-1.5">
                    <label for="name" class="text-sm font-semibold">Nama</label>
                    <input
                        id="name"
                        v-model="name"
                        type="text"
                        autocomplete="name"
                        minlength="3"
                        maxlength="100"
                        required
                        class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 text-base focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                    />
                </div>

                <div class="flex flex-col gap-1.5">
                    <label for="username" class="text-sm font-semibold">Username</label>
                    <input
                        id="username"
                        v-model="username"
                        type="text"
                        autocomplete="username"
                        minlength="3"
                        maxlength="50"
                        required
                        class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 text-base focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                    />
                </div>

                <div class="flex flex-col gap-1.5">
                    <label for="email" class="text-sm font-semibold">Email</label>
                    <input
                        id="email"
                        v-model="email"
                        type="email"
                        autocomplete="email"
                        maxlength="100"
                        required
                        class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 text-base focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                    />
                </div>

                <div class="flex flex-col gap-1.5">
                    <label for="phone" class="text-sm font-semibold">Nomor Telepon</label>
                    <input
                        id="phone"
                        v-model="phone"
                        type="tel"
                        autocomplete="tel"
                        minlength="10"
                        maxlength="20"
                        required
                        placeholder="Contoh: 081234567890"
                        class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 text-base focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                    />
                </div>

                <div class="flex flex-col gap-1.5">
                    <label for="password" class="text-sm font-semibold">Password</label>
                    <input
                        id="password"
                        v-model="password"
                        type="password"
                        autocomplete="new-password"
                        minlength="6"
                        maxlength="100"
                        required
                        class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 text-base focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                    />
                </div>
                <div ref="captchaEl"></div>

                <button
                    type="submit"
                    :disabled="loading || !name.trim() || username.trim().length < 3 || !email.trim() || phone.trim().length < 10 || password.length < 6"
                    class="mt-1 inline-flex items-center justify-center rounded-md bg-[#f2b705] px-4 py-3.5 text-lg font-extrabold text-[#12372a] transition-colors hover:bg-[#ffc61f] disabled:cursor-not-allowed disabled:bg-[#e4e8e5] disabled:text-[#8a968f]"
                >
                    {{ loading ? 'Memproses…' : 'Daftar' }}
                </button>
            </form>
        </section>
    </main>
</template>