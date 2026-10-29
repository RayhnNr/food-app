<script setup>
import { ref, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import api from '@/api';
import Modal from '@/components/Modal.vue';

const props = defineProps({
    show: { type: Boolean, default: false },
    // Satu baris dari GET /barcode (item_id, goods_partai_name, code_min, code_max, ...)
    batch: { type: Object, default: null },
});

const emit = defineEmits(['close']);

const items = ref([]);
const loading = ref(false);
const error = ref('');
const page = ref(1);
const perPage = 50;

const router = useRouter();

// key dipakai sebagai query "jenis" di halaman cetak (BarcodePrintView).
const aksiCetak = [
    { key: 'pita-5x2', label: 'Print Pita (5x2)', singkat: 'Pita 5x2' },
    { key: 'pita-5x2.5', label: 'Print Pita (5x2.5)', singkat: 'Pita 5x2.5' },
    { key: 'bal', label: 'Print Bal', singkat: 'Bal' },
];

const tombolCetak =
    'inline-flex cursor-pointer items-center justify-center gap-1 rounded-md border border-gray-300 bg-white text-gray-700 transition hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-50';

const rentang = computed(() => {
    if (!props.batch) return '';
    const { code_min: min, code_max: max } = props.batch;
    return min === max ? min : `${min} s/d ${max}`;
});

const judul = computed(() => (props.batch ? `Detail Barcode ${rentang.value}` : ''));

const totalHalaman = computed(() => Math.max(1, Math.ceil(items.value.length / perPage)));
const offset = computed(() => (page.value - 1) * perPage);
const itemsHalaman = computed(() => items.value.slice(offset.value, offset.value + perPage));

const formatTanggal = (value) =>
    new Date(value).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });

const load = async () => {
    if (!props.batch) return;

    const itemID = props.batch.item_id;
    loading.value = true;
    error.value = '';
    items.value = [];
    page.value = 1;

    try {
        const res = await api.get(`/barcode/${itemID}`);
        // Abaikan hasil lama kalau modal sudah dibuka untuk batch lain.
        if (props.batch?.item_id !== itemID) return;
        items.value = res.data.data;
    } catch (e) {
        if (props.batch?.item_id !== itemID) return;
        error.value = e.response?.data?.message || 'Gagal mengambil detail barcode';
        console.error(e);
    } finally {
        if (props.batch?.item_id === itemID) loading.value = false;
    }
};

// Buka halaman cetak di tab baru. Tanpa itemCode berarti cetak semua kode di batch.
const cetak = (jenis, itemCode = null) => {
    if (!props.batch) return;

    const query = { jenis };
    if (itemCode) query.item_code = itemCode;

    const { href } = router.resolve({ path: `/cetak/barcode/${props.batch.item_id}`, query });
    // noopener: tab cetak berjalan terpisah, jadi dialog print yang masih terbuka tidak membebani tab ini.
    window.open(href, '_blank', 'noopener,noreferrer');
};

watch(
    () => props.show,
    (visible) => {
        if (visible) load();
    },
);
</script>

<template>
    <Modal :show="show" :title="judul" max-width="max-w-5xl" @close="emit('close')">
        <div class="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <p class="text-sm text-gray-500">
                {{ items.length }} barcode<span v-if="batch?.goods_partai_name">
                    pada partai {{ batch.goods_partai_name }}</span
                >
            </p>

            <div class="flex flex-wrap items-center gap-2">
                <span class="text-sm text-gray-500">Print semua:</span>
                <button
                    v-for="aksi in aksiCetak"
                    :key="aksi.key"
                    type="button"
                    :disabled="loading || items.length === 0"
                    :class="[tombolCetak, 'px-3 py-2 text-sm']"
                    @click="cetak(aksi.key)"
                >
                    <i class="fa-solid fa-print"></i> {{ aksi.label }}
                </button>
            </div>
        </div>

        <div class="max-h-[55vh] overflow-auto rounded-lg border border-gray-200">
            <table class="w-full table-auto">
                <thead class="sticky top-0 bg-gray-200 text-sm leading-normal text-gray-600 uppercase">
                    <tr>
                        <th class="py-3 px-4 text-left">#</th>
                        <th class="py-3 px-4 text-left">Item Code</th>
                        <th class="py-3 px-4 text-left">Partai</th>
                        <th class="py-3 px-4 text-left">Tanggal Buat</th>
                        <th class="py-3 px-4 text-left">Status</th>
                        <th class="py-3 px-4 text-left">Aksi</th>
                    </tr>
                </thead>
                <tbody class="text-sm text-gray-600">
                    <tr v-if="loading">
                        <td colspan="6" class="py-10 text-center text-gray-500">Memuat data...</td>
                    </tr>
                    <tr v-else-if="error">
                        <td colspan="6" class="py-10 text-center text-red-600">{{ error }}</td>
                    </tr>
                    <tr v-else-if="items.length === 0">
                        <td colspan="6" class="py-10 text-center text-gray-500">
                            Tidak ada barcode pada batch ini.
                        </td>
                    </tr>
                    <template v-else>
                        <tr
                            v-for="(item, index) in itemsHalaman"
                            :key="item.item_code"
                            class="border-b border-gray-200 hover:bg-gray-100"
                        >
                            <td class="py-2 px-4 text-left">{{ offset + index + 1 }}</td>
                            <td class="py-2 px-4 text-left font-mono text-xs font-semibold">
                                {{ item.item_code }}
                            </td>
                            <td class="py-2 px-4 text-left">{{ batch?.goods_partai_name || '-' }}</td>
                            <td class="py-2 px-4 text-left whitespace-nowrap">
                                {{ formatTanggal(item.item_date_create) }}
                            </td>
                            <td class="py-2 px-4 text-left">
                                <span
                                    class="inline-flex whitespace-nowrap rounded px-2 py-1 text-xs font-semibold"
                                    :class="
                                        item.item_code_status === 1
                                            ? 'bg-emerald-100 text-emerald-800'
                                            : 'bg-slate-100 text-slate-600'
                                    "
                                >
                                    {{ item.item_code_status === 1 ? 'Tersedia' : 'Tidak Tersedia' }}
                                </span>
                            </td>
                            <td class="py-2 px-4 text-left">
                                <div class="flex items-center gap-2">
                                    <button
                                        v-for="aksi in aksiCetak"
                                        :key="aksi.key"
                                        type="button"
                                        :title="aksi.label"
                                        @click="cetak(aksi.key, item.item_code)"
                                        :aria-label="`${aksi.label} ${item.item_code}`"
                                        :class="[tombolCetak, 'px-2 py-1 text-xs whitespace-nowrap']"
                                    >
                                        <i class="fa-solid fa-print"></i> {{ aksi.singkat }}
                                    </button>
                                </div>
                            </td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>

        <div v-if="totalHalaman > 1" class="mt-4 flex items-center justify-between text-sm text-gray-600">
            <span>Halaman {{ page }} dari {{ totalHalaman }}</span>
            <div class="flex gap-2">
                <button
                    type="button"
                    :disabled="page === 1"
                    @click="page--"
                    :class="[tombolCetak, 'px-3 py-1']"
                >
                    Sebelumnya
                </button>
                <button
                    type="button"
                    :disabled="page === totalHalaman"
                    @click="page++"
                    :class="[tombolCetak, 'px-3 py-1']"
                >
                    Berikutnya
                </button>
            </div>
        </div>
    </Modal>
</template>