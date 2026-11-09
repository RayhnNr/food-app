<script setup>
import { computed, ref } from 'vue';
import { RouterLink } from 'vue-router';
import api from '@/api';

const email = ref('');
const otp = ref('');
const newPassword = ref('');
const confirmPassword = ref('');
const step = ref('email');
const error = ref('');
const message = ref('');
const loading = ref(false);

const canRequestCode = computed(() => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim()));
const canResetPassword = computed(
    () =>
        otp.value.trim().length === 10 &&
        newPassword.value.length >= 6 &&
        newPassword.value.length <= 72 &&
        confirmPassword.value === newPassword.value,
);

async function requestCode() {
    if (loading.value || !canRequestCode.value) return;

    error.value = '';
    message.value = '';
    loading.value = true;
    try {
        const response = await api.post('/password-reset', { email: email.value.trim() });
        message.value =
            response.data?.message ||
            'Jika email terdaftar, kode reset akan dikirim. Periksa kotak masuk email Anda.';
        step.value = 'otp';
    } catch (err) {
        error.value =
            err.response?.data?.message ||
            err.response?.data?.error ||
            'Kode reset belum dapat dikirim. Silakan coba lagi.';
    } finally {
        loading.value = false;
    }
}

async function resetPassword() {
    if (loading.value || !canResetPassword.value) return;

    error.value = '';
    message.value = '';
    loading.value = true;
    try {
        const response = await api.post('/password-reset/confirm', {
            email: email.value.trim(),
            otp: otp.value.trim().toUpperCase(),
            new_password: newPassword.value,
        });
        message.value = response.data?.message || 'Password berhasil diperbarui.';
        step.value = 'done';
    } catch (err) {
        error.value =
            err.response?.data?.message ||
            err.response?.data?.error ||
            'Password belum dapat diperbarui. Periksa kode dan coba lagi.';
    } finally {
        loading.value = false;
    }
}

function editEmail() {
    step.value = 'email';
    otp.value = '';
    newPassword.value = '';
    confirmPassword.value = '';
    error.value = '';
    message.value = '';
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
            <h2 class="max-w-[14ch] text-3xl font-extrabold leading-[1.05] tracking-tight md:text-5xl lg:text-6xl">
                Akunmu aman, aksesnya bisa dipulihkan.
            </h2>
        </aside>

        <section class="grid place-items-center bg-[#f7f8f5] px-5 py-10">
            <form
                v-if="step !== 'done'"
                class="flex w-full max-w-sm flex-col gap-5"
                novalidate
                @submit.prevent="step === 'email' ? requestCode() : resetPassword()"
            >
                <div>
                    <p class="text-sm font-bold uppercase tracking-[0.14em] text-[#527064]">
                        {{ step === 'email' ? 'Pemulihan akun' : 'Verifikasi email' }}
                    </p>
                    <h1 class="mt-2 text-3xl font-extrabold leading-tight tracking-tight">
                        {{ step === 'email' ? 'Lupa password?' : 'Buat password baru' }}
                    </h1>
                    <p class="mt-2 text-[#5d6b64]">
                        {{
                            step === 'email'
                                ? 'Masukkan email akunmu. Kami akan mengirimkan kode untuk mengatur ulang password.'
                                : `Masukkan kode 10 karakter yang dikirim ke ${email}.`
                        }}
                    </p>
                </div>

                <p
                    v-if="error"
                    role="alert"
                    class="border-l-4 border-[#b3261e] bg-[#fbeceb] px-4 py-3 text-sm leading-snug text-[#7a1b16]"
                >
                    {{ error }}
                </p>
                <p
                    v-if="message"
                    role="status"
                    class="border-l-4 border-[#28734e] bg-[#eaf4ed] px-4 py-3 text-sm leading-snug text-[#24543a]"
                >
                    {{ message }}
                </p>

                <template v-if="step === 'email'">
                    <div class="flex flex-col gap-1.5">
                        <label for="reset-email" class="text-sm font-semibold">Email akun</label>
                        <input
                            id="reset-email"
                            v-model="email"
                            type="email"
                            autocomplete="email"
                            maxlength="100"
                            placeholder="nama@email.com"
                            required
                            class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 text-base placeholder:text-[#94a09a] hover:border-[#a9b5ae] focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                        />
                    </div>
                </template>

                <template v-else>
                    <div class="flex flex-col gap-1.5">
                        <label for="reset-otp" class="text-sm font-semibold">Kode reset</label>
                        <input
                            id="reset-otp"
                            v-model="otp"
                            type="text"
                            autocomplete="one-time-code"
                            minlength="10"
                            maxlength="10"
                            inputmode="text"
                            placeholder="Contoh: AB12CD34EF"
                            required
                            class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 font-mono text-base uppercase tracking-[0.2em] placeholder:font-sans placeholder:tracking-normal placeholder:text-[#94a09a] hover:border-[#a9b5ae] focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                        />
                    </div>

                    <div class="flex flex-col gap-1.5">
                        <label for="new-password" class="text-sm font-semibold">Password baru</label>
                        <input
                            id="new-password"
                            v-model="newPassword"
                            type="password"
                            autocomplete="new-password"
                            minlength="6"
                            maxlength="72"
                            required
                            class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 text-base focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                        />
                        <p class="text-xs text-[#5d6b64]">Gunakan 6–72 karakter.</p>
                    </div>

                    <div class="flex flex-col gap-1.5">
                        <label for="confirm-password" class="text-sm font-semibold">
                            Ulangi password baru
                        </label>
                        <input
                            id="confirm-password"
                            v-model="confirmPassword"
                            type="password"
                            autocomplete="new-password"
                            minlength="6"
                            maxlength="72"
                            required
                            class="w-full rounded-md border-[1.5px] border-[#cfd6d1] bg-white px-3.5 py-3 text-base focus-visible:border-[#12372a] focus-visible:outline-3 focus-visible:outline-offset-1 focus-visible:outline-[#f2b705]"
                        />
                        <p
                            v-if="confirmPassword && confirmPassword !== newPassword"
                            class="text-sm text-[#7a1b16]"
                        >
                            Password belum sama.
                        </p>
                    </div>
                </template>

                <button
                    type="submit"
                    :disabled="loading || (step === 'email' ? !canRequestCode : !canResetPassword)"
                    class="mt-1 inline-flex items-center justify-center gap-2.5 rounded-md bg-[#f2b705] px-4 py-3.5 text-lg font-extrabold text-[#12372a] transition-colors hover:bg-[#ffc61f] focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-[#12372a] disabled:cursor-not-allowed disabled:bg-[#e4e8e5] disabled:text-[#8a968f]"
                >
                    <span
                        v-if="loading"
                        aria-hidden="true"
                        class="size-4 animate-spin rounded-full border-[2.5px] border-current border-r-transparent motion-reduce:[animation-duration:2s]"
                    ></span>
                    {{
                        loading
                            ? 'Memproses…'
                            : step === 'email'
                              ? 'Kirim kode reset'
                              : 'Simpan password baru'
                    }}
                </button>

                <button
                    v-if="step === 'otp'"
                    type="button"
                    class="self-center rounded px-2 py-1 text-sm font-semibold text-[#12372a] underline underline-offset-4 hover:text-[#0b2219] focus-visible:outline-3 focus-visible:outline-[#f2b705]"
                    @click="editEmail"
                >
                    Gunakan email lain
                </button>

                <RouterLink
                    to="/login"
                    class="self-center rounded px-2 py-1 text-sm font-semibold text-[#12372a] underline underline-offset-4 hover:text-[#0b2219] focus-visible:outline-3 focus-visible:outline-[#f2b705]"
                >
                    Kembali ke halaman masuk
                </RouterLink>
            </form>

            <div v-else class="flex w-full max-w-sm flex-col gap-5">
                <div
                    aria-hidden="true"
                    class="grid size-14 place-items-center rounded-full bg-[#eaf4ed] text-2xl font-bold text-[#28734e]"
                >
                    ✓
                </div>
                <div>
                    <h1 class="text-3xl font-extrabold leading-tight tracking-tight">
                        Password diperbarui
                    </h1>
                    <p class="mt-2 text-[#5d6b64]">{{ message }}</p>
                </div>
                <RouterLink
                    to="/login"
                    class="inline-flex items-center justify-center rounded-md bg-[#f2b705] px-4 py-3.5 text-lg font-extrabold text-[#12372a] transition-colors hover:bg-[#ffc61f] focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-[#12372a]"
                >
                    Masuk dengan password baru
                </RouterLink>
            </div>
        </section>
    </main>
</template>
