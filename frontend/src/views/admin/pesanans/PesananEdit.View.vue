<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useToast } from 'vue-toast-notification';
import api from '@/api';

const route = useRoute();
const router = useRouter();
const toast = useToast();

const loading = ref(true);
const saving = ref(false);
const loadError = ref('');
const errors = ref({});
const form = ref({
    nama: '',
    noMeja: '',
    keranjangs: [],
    status: 'diproses',
    tanggal: '',
    total: 0,
});

const total = computed(() =>
    form.value.keranjangs.reduce((sum, item) => {
        const harga = Number(item.harga) || 0;
        const jumlah = Number(item.jumlah) || 0;
        return sum + harga * jumlah;
    }, 0),
);

const formatRupiah = (angka) =>
    new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
    }).format(angka);

const muatPesanan = async () => {
    loading.value = true;
    loadError.value = '';

    try {
        const { data } = await api.get(
            `/admin/pesanans/${route.params.id}`,
        );

        form.value = {
            ...data,
            noMeja: data.no_meja ?? '',
            keranjangs: Array.isArray(data.items) ? data.items : [],
        };
    } catch (error) {
        loadError.value = 'Pesanan gagal dimuat.';
        toast.error(loadError.value);
        console.error(error);
    } finally {
        loading.value = false;
    }
};

const hapusItem = (index) => {
    form.value.keranjangs.splice(index, 1);
};

const validasi = () => {
    errors.value = {};

    if (!form.value.nama?.trim()) errors.value.nama = 'Nama pelanggan wajib diisi';

    const noMeja = Number(form.value.noMeja);
    if (!Number.isInteger(noMeja) || noMeja <= 0) {
        errors.value.noMeja = 'Nomor meja harus berupa angka lebih dari 0';
    }

    if (form.value.keranjangs.length === 0) {
        errors.value.keranjangs = 'Pesanan harus memiliki minimal satu item';
    } else if (
        form.value.keranjangs.some(
            (item) => !Number.isInteger(Number(item.jumlah)) || Number(item.jumlah) <= 0,
        )
    ) {
        errors.value.keranjangs = 'Jumlah setiap item harus berupa angka lebih dari 0';
    }

    return Object.keys(errors.value).length === 0;
};

const simpan = async () => {
    if (!validasi()) return;

    saving.value = true;

    try {
        await api.put(`/admin/pesanans/${route.params.id}`, {
            nama: form.value.nama.trim(),
            no_meja: Number(form.value.noMeja),
            status: form.value.status,
            items: form.value.keranjangs.map((item) => ({
                id: item.id,
                jumlah: Number(item.jumlah),
                keterangan: item.keterangan || '',
            })),
        });

        toast.success('Pesanan berhasil diperbarui');
        router.push('/admin/pesanans');
    } catch (error) {
        toast.error(
            error.response?.data?.message ||
                'Gagal memperbarui pesanan',
        );
        console.error(error);
    } finally {
        saving.value = false;
    }
};

watch(form, () => {
        errors.value = {};
    },
    { deep: true },
);

onMounted(muatPesanan);
</script>

<template>
    <section class="max-w-7xl mx-auto px-4 py-10">
        <h1 class="text-2xl font-bold text-slate-800 mb-6">Edit Pesanan</h1>

        <p v-if="loading" class="text-slate-500">Memuat pesanan...</p>
        <div v-else-if="loadError" class="rounded-xl bg-red-50 p-4 text-red-700">
            {{ loadError }}
        </div>

        <form
            v-else
            @submit.prevent="simpan"
            novalidate
            class="space-y-6 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm ring-1 ring-slate-100 sm:p-6"
        >
            <div class="grid gap-4 sm:grid-cols-2">
                <div>
                    <label for="nama" class="block text-sm font-medium mb-1">Nama pelanggan</label>
                    <input
                        id="nama"
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

                <div>
                    <label for="noMeja" class="block text-sm font-medium mb-1">Nomor meja</label>
                    <input
                        id="noMeja"
                        v-model.number="form.noMeja"
                        type="number"
                        min="1"
                        class="w-full px-4 py-2 rounded-xl border focus:outline-none focus:ring-4"
                        :class="
                            errors.noMeja
                                ? 'border-red-500 focus:ring-red-200'
                                : 'border-slate-300 focus:ring-emerald-300'
                        "
                    />
                    <p v-if="errors.noMeja" class="mt-1 text-sm text-red-600">
                        {{ errors.noMeja }}
                    </p>
                </div>
            </div>

            <div>
                <label for="status" class="block text-sm font-medium mb-1">Status</label>
                <select
                    id="status"
                    v-model="form.status"
                    class="w-full px-4 py-2 rounded-xl border border-slate-300 bg-white focus:outline-none focus:ring-4 focus:ring-emerald-300"
                >
                    <option value="diproses">Diproses</option>
                    <option value="selesai">Selesai</option>
                </select>
            </div>

            <div class="rounded-2xl border border-slate-200 bg-slate-50 p-4 sm:p-5">
                <h2 class="font-semibold text-slate-800 mb-3">Item pesanan</h2>
                <p v-if="errors.keranjangs" class="mb-3 text-sm text-red-600">
                    {{ errors.keranjangs }}
                </p>

                <div v-if="form.keranjangs.length" class="space-y-3">
                    <article
                        v-for="(item, index) in form.keranjangs"
                        :key="item.id"
                        class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
                    >
                        <div class="flex flex-wrap items-start justify-between gap-3">
                            <div>
                                <p class="font-medium text-slate-800">
                                    {{ item.nama || 'Produk tidak tersedia' }}
                                </p>
                                <p class="text-sm text-slate-500">
                                    {{ formatRupiah(Number(item.harga) || 0) }} / item
                                </p>
                            </div>
                            <button
                                type="button"
                                class="text-sm text-red-600 hover:text-red-800"
                                @click="hapusItem(index)"
                            >
                                Hapus
                            </button>
                        </div>

                        <div class="mt-3 grid gap-3 sm:grid-cols-2">
                            <div>
                                <label
                                    :for="`jumlah-${item.id}`"
                                    class="block text-sm font-medium mb-1"
                                >
                                    Jumlah
                                </label>
                                <input
                                    :id="`jumlah-${item.id}`"
                                    v-model.number="item.jumlah"
                                    type="number"
                                    min="1"
                                    class="w-full px-3 py-2 rounded-lg border border-slate-300"
                                />
                            </div>
                            <div>
                                <label
                                    :for="`keterangan-${item.id}`"
                                    class="block text-sm font-medium mb-1"
                                >
                                    Keterangan
                                </label>
                                <input
                                    :id="`keterangan-${item.id}`"
                                    v-model.trim="item.keterangan"
                                    class="w-full px-3 py-2 rounded-lg border border-slate-300"
                                />
                            </div>
                        </div>
                        <p class="mt-2 text-right text-sm font-medium text-slate-600">
                            Subtotal:
                            {{
                                formatRupiah(
                                    (Number(item.harga) || 0) *
                                    (Number(item.jumlah) || 0)
                                )
                            }}
                        </p>
                    </article>
                </div>
                <p v-else class="rounded-xl bg-white p-4 text-sm text-slate-500 shadow-sm">
                    Belum ada item dalam pesanan ini.
                </p>
            </div>

            <div class="border-t border-slate-200 pt-4">
                <p class="text-right text-lg font-bold text-emerald-700">
                    Total: {{ formatRupiah(total) }}
                </p>
            </div>

            <div class="flex gap-3">
                <button
                    type="submit"
                    :disabled="saving"
                    class="px-5 py-2.5 rounded-xl bg-emerald-600 text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {{ saving ? 'Menyimpan...' : 'Simpan perubahan' }}
                </button>
                <RouterLink
                    to="/admin/pesanans"
                    class="px-5 py-2.5 rounded-xl border border-slate-300"
                >
                    Batal
                </RouterLink>
            </div>
        </form>
    </section>
</template>
