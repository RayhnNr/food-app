<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import Swal from 'sweetalert2';
import api from '@/api';

const route = useRoute();
const router = useRouter();

const file = ref(null);
const preview = ref('');
const dragging = ref(false);
const saving = ref(false);

const isEdit = computed(() => !!route.params.id);

const form = ref({
    nama: '',
    harga: 0,
    is_ready: true,
    gambar: '',
});

const errors = ref({});

// Alamat dasar server hanya ditulis di src/api.js
const urlGambar = (nama) => `${api.defaults.baseURL}/uploads/${nama}`;

// ===== Harga =====
const hargaTampil = computed(() =>
    form.value.harga ? Number(form.value.harga).toLocaleString('id-ID') : '',
);

const handleHarga = (event) => {
    const angka = event.target.value.replace(/\D/g, '');
    form.value.harga = angka === '' ? 0 : Number(angka);
    event.target.value = angka === '' ? '' : Number(angka).toLocaleString('id-ID');
};

// ===== Ambil data saat edit =====
onMounted(async () => {
    if (isEdit.value) {
        try {
            const res = await api.get(`/products/${route.params.id}`);
            form.value = res.data;
        } catch (e) {
            Swal.fire({ icon: 'error', title: 'Gagal', text: 'Gagal mengambil data menu' });
            console.error(e);
        }
    }
});

onBeforeUnmount(() => {
    if (preview.value) URL.revokeObjectURL(preview.value);
});

// ===== Validasi =====
const validasi = () => {
    const err = {};

    if (!form.value.nama) {
        err.nama = 'Nama wajib diisi';
    } else if (form.value.nama.length < 3) {
        err.nama = 'Nama minimal 3 karakter';
    }

    if (!form.value.harga || form.value.harga <= 0) {
        err.harga = 'Harga harus lebih dari 0';
    }

    if (errors.value.gambar) err.gambar = errors.value.gambar;

    errors.value = err;
    return Object.keys(err).length === 0;
};

// ===== File =====
const prosesFile = (f) => {
    if (!f) return;
    if (!['image/jpeg', 'image/png', 'image/webp'].includes(f.type)) {
        errors.value.gambar = 'Hanya JPG, PNG, atau WEBP';
        return;
    }
    if (f.size > 2 * 1024 * 1024) {
        errors.value.gambar = 'Ukuran maksimal 2 MB';
        return;
    }
    errors.value.gambar = '';
    if (preview.value) URL.revokeObjectURL(preview.value);
    file.value = f;
    preview.value = URL.createObjectURL(f);
};

const pilihFile = (e) => {
    prosesFile(e.target.files[0]);
    e.target.value = null;
};

const dropFile = (e) => {
    dragging.value = false;
    prosesFile(e.dataTransfer.files[0]);
};

const hapusGambar = () => {
    if (preview.value) URL.revokeObjectURL(preview.value);
    file.value = null;
    preview.value = '';
    form.value.gambar = '';
};

const ukuranFile = (bytes) =>
    bytes > 1024 * 1024
        ? (bytes / 1024 / 1024).toFixed(1) + ' MB'
        : Math.round(bytes / 1024) + ' KB';

// ===== Simpan =====
const simpan = async () => {
    if (saving.value || !validasi()) return;
    saving.value = true;

    try {
        if (file.value) {
            const data = new FormData();
            data.append('gambar', file.value);
            const res = await api.post('/upload', data);
            form.value.gambar = res.data.gambar;
        }

        if (isEdit.value) {
            await api.put(`/products/${route.params.id}`, form.value);
        } else {
            await api.post('/products', form.value);
        }

        await Swal.fire({
            icon: 'success',
            title: 'Berhasil',
            text: isEdit.value ? 'Berhasil diperbarui' : 'Berhasil ditambahkan',
            timer: 1500,
            showConfirmButton: false,
        });

        router.push('/admin/foods');
    } catch (e) {
        if (e.response?.status === 422 && e.response.data.errors) {
            errors.value = e.response.data.errors;
            return;
        }
        Swal.fire({
            icon: 'error',
            title: 'Gagal',
            text: e.response?.data?.message || 'Gagal menyimpan',
        });
    } finally {
        saving.value = false;
    }
};

// Hapus error nama/harga saat diubah (error gambar dipertahankan)
watch(
    () => [form.value.nama, form.value.harga],
    () => {
        const { gambar } = errors.value;
        errors.value = gambar ? { gambar } : {};
    },
);
</script>

<template>
    <section class="max-w-xl mx-auto px-4 py-12">
        <h1 class="text-2xl font-bold text-slate-800 mb-6">
            {{ isEdit ? 'Edit Menu' : 'Tambah Menu' }}
        </h1>

        <form @submit.prevent="simpan" novalidate class="space-y-4">
            <!-- Kode (hanya saat edit, tidak bisa diubah) -->
            <div v-if="isEdit">
                <label class="block text-sm font-medium mb-1">Kode</label>
                <input
                    :value="form.kode"
                    readonly
                    class="w-full px-4 py-2 rounded-xl border border-slate-200 bg-slate-100 text-slate-500 cursor-not-allowed focus:outline-none"
                />
            </div>

            <!-- Nama -->
            <div>
                <label class="block text-sm font-medium mb-1">Nama</label>
                <input
                    v-model.trim="form.nama"
                    class="w-full px-4 py-2 rounded-xl border focus:outline-none focus:ring-4"
                    :class="
                        errors.nama
                            ? 'border-red-500 focus:ring-red-200'
                            : 'border-slate-300 focus:ring-emerald-300'
                    "
                />
                <p v-if="errors.nama" class="mt-1 text-sm text-red-600">{{ errors.nama }}</p>
            </div>

            <!-- Harga -->
            <div>
                <label class="block text-sm font-medium mb-1">Harga</label>
                <input
                    :value="hargaTampil"
                    @input="handleHarga"
                    type="text"
                    inputmode="numeric"
                    placeholder="Contoh: 1.500.000"
                    class="w-full px-4 py-2 rounded-xl border focus:outline-none focus:ring-4"
                    :class="
                        errors.harga
                            ? 'border-red-500 focus:ring-red-200'
                            : 'border-slate-300 focus:ring-emerald-300'
                    "
                />
                <p v-if="errors.harga" class="mt-1 text-sm text-red-600">{{ errors.harga }}</p>
            </div>

            <!-- Gambar -->
            <div>
                <label class="block text-sm font-medium text-slate-700 mb-1">Foto Menu</label>

                <label
                    v-if="!preview && !form.gambar"
                    @dragover.prevent="dragging = true"
                    @dragleave.prevent="dragging = false"
                    @drop.prevent="dropFile"
                    class="flex flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed px-6 py-10 text-center cursor-pointer transition"
                    :class="
                        dragging
                            ? 'border-emerald-500 bg-emerald-50'
                            : errors.gambar
                              ? 'border-red-400 bg-red-50'
                              : 'border-slate-300 bg-slate-50 hover:border-emerald-400 hover:bg-emerald-50/50'
                    "
                >
                    <span
                        class="flex h-12 w-12 items-center justify-center rounded-full bg-emerald-100 text-emerald-600"
                    >
                        <i class="fa-solid fa-cloud-arrow-up text-xl"></i>
                    </span>
                    <p class="text-sm text-slate-600">
                        <span class="font-semibold text-emerald-600">Klik untuk memilih</span>
                        atau seret foto ke sini
                    </p>
                    <p class="text-xs text-slate-400">JPG, PNG, atau WEBP, maksimal 2 MB</p>
                    <input
                        type="file"
                        accept="image/png,image/jpeg,image/webp"
                        class="hidden"
                        @change="pilihFile"
                    />
                </label>

                <div
                    v-else
                    class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm"
                >
                    <img
                        :src="preview || urlGambar(form.gambar)"
                        alt="Preview foto menu"
                        class="h-48 w-full object-cover"
                    />
                    <div class="flex items-center justify-between gap-3 px-4 py-3">
                        <div class="min-w-0">
                            <p class="truncate text-sm font-medium text-slate-700">
                                {{ file ? file.name : form.gambar }}
                            </p>
                            <p v-if="file" class="text-xs text-slate-400">
                                {{ ukuranFile(file.size) }}
                            </p>
                            <p v-else class="text-xs text-slate-400">Foto saat ini</p>
                        </div>
                        <div class="flex shrink-0 gap-2">
                            <label
                                class="cursor-pointer rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-medium text-slate-600 transition hover:bg-slate-50"
                            >
                                Ganti
                                <input
                                    type="file"
                                    accept="image/png,image/jpeg,image/webp"
                                    class="hidden"
                                    @change="pilihFile"
                                />
                            </label>
                            <button
                                type="button"
                                @click="hapusGambar"
                                class="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-medium text-red-600 transition hover:bg-red-100"
                            >
                                <i class="fa-solid fa-trash mr-1"></i>Hapus
                            </button>
                        </div>
                    </div>
                </div>

                <p v-if="errors.gambar" class="mt-1 text-sm text-red-600">
                    <i class="fa-solid fa-circle-exclamation mr-1"></i>{{ errors.gambar }}
                </p>
            </div>

            <!-- Tersedia (toggle) -->
            <label class="relative inline-flex items-center cursor-pointer gap-3 text-slate-800">
                <input type="checkbox" v-model="form.is_ready" class="sr-only peer" />
                <div
                    class="w-14 h-8 bg-slate-300 rounded-full transition-colors duration-200 peer-checked:bg-emerald-600 peer-focus-visible:ring-4 peer-focus-visible:ring-emerald-300"
                ></div>
                <span
                    class="absolute left-1 top-1 w-6 h-6 bg-white rounded-full shadow transition-transform duration-200 peer-checked:translate-x-6"
                ></span>
                <span>Tersedia</span>
            </label>

            <!-- Tombol -->
            <div class="flex gap-3">
                <button
                    type="submit"
                    :disabled="saving"
                    class="px-5 py-2.5 rounded-xl bg-emerald-600 text-white hover:bg-emerald-700 disabled:bg-slate-300 disabled:cursor-not-allowed"
                >
                    {{ saving ? 'Menyimpan...' : 'Simpan' }}
                </button>
                <RouterLink
                    to="/admin/foods"
                    class="px-5 py-2.5 rounded-xl border border-slate-300"
                >
                    Batal
                </RouterLink>
            </div>
        </form>
    </section>
</template>