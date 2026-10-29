<script setup>
import { ref, computed, onMounted } from 'vue';
import api from '@/api';
import Swal from 'sweetalert2';
import BarcodeDetailModal from './BarcodeDetailModal.vue';

const barcodes = ref([]);
const loading = ref(true);
const error = ref(null);
const search = ref('');

const detailBarcode = ref(null);
const showDetail = ref(false);

const bukaDetail = (barcode) => {
    detailBarcode.value = barcode;
    showDetail.value = true;
};

const tutupDetail = () => {
    showDetail.value = false;
};

const actionButtonClass =
    'inline-flex h-9 w-9 items-center justify-center rounded-md border text-gray-600 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500';

const getBarcode = async () => {
    loading.value = true;
    error.value = null;

    try {
        const res = await api.get('/barcode');
        barcodes.value = res.data.data;
    } catch (e) {
        error.value = 'Gagal mengambil data barcode. Pastikan server API sudah jalan.';
        console.error(e);
    } finally {
        loading.value = false;
    }
};

// Pencarian dilakukan di browser, karena endpoint /barcode belum punya filter.
const filteredBarcodes = computed(() => {
    const q = search.value.trim().toLowerCase();
    if (!q) return barcodes.value;

    return barcodes.value.filter(
        (b) =>
            b.code_min.toLowerCase().includes(q) ||
            b.code_max.toLowerCase().includes(q) ||
            (b.goods_partai_name || '').toLowerCase().includes(q),
    );
});

const rentangKode = (b) => (b.code_min === b.code_max ? b.code_min : `${b.code_min} s/d ${b.code_max}`);

const formatTanggal = (value) =>
    new Date(value).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });

const hapusBarcode = async (barcode) => {
    const result = await Swal.fire({
        title: `Yakin hapus ${barcode.total} barcode?`,
        text: rentangKode(barcode),
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Ya, hapus',
        cancelButtonText: 'Batal',
    });

    if (!result.isConfirmed) return;

    try {
        await api.delete(`/barcode/${barcode.item_id}`);
        await getBarcode();
        Swal.fire({ icon: 'success', title: 'Berhasil menghapus barcode' });
    } catch (e) {
        Swal.fire({
            icon: 'error',
            title: 'Gagal menghapus',
            text: e.response?.data?.message || 'Gagal menghapus barcode',
        });
        console.error(e);
    }
};

onMounted(getBarcode);
</script>

<template>
    <section class="p-6">
        <h1 class="text-3xl font-bold mb-8">Data Barcode</h1>

        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between mb-6">
            <div class="w-full md:w-1/3">
                <input
                    v-model="search"
                    type="text"
                    placeholder="Cari kode atau nama partai..."
                    class="w-full px-4 py-2 rounded-md border border-gray-300 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
            </div>

            <div class="flex gap-2">
                <RouterLink
                    to="/admin/barcode/internal"
                    class="px-4 py-2 rounded-xl bg-emerald-600 text-white hover:bg-emerald-700"
                >
                    <i class="fa-solid fa-plus mr-1"></i> Barcode Internal
                </RouterLink>
                <RouterLink
                    to="/admin/barcode/external"
                    class="px-4 py-2 rounded-xl bg-blue-600 text-white hover:bg-blue-700"
                >
                    <i class="fa-solid fa-plus mr-1"></i> Barcode External
                </RouterLink>
            </div>
        </div>

        <p v-if="error" class="mb-4 rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            {{ error }}
        </p>

        <div class="overflow-x-auto bg-white rounded-lg shadow">
            <table class="w-full table-auto">
                <thead>
                    <tr class="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                        <th class="py-3 px-6 text-left">#</th>
                        <th class="py-3 px-6 text-left">Item Code</th>
                        <th class="py-3 px-6 text-left">Partai</th>
                        <th class="py-3 px-6 text-left">Tipe</th>
                        <th class="py-3 px-6 text-left">Tahun</th>
                        <th class="py-3 px-6 text-right">Total</th>
                        <th class="py-3 px-6 text-left">Tanggal Buat</th>
                        <th class="py-3 px-6 text-left">Aksi</th>
                    </tr>
                </thead>
                <tbody class="text-gray-600 text-sm">
                    <tr v-if="loading">
                        <td colspan="8" class="text-center py-10 text-gray-500">Memuat data...</td>
                    </tr>
                    <tr v-else-if="filteredBarcodes.length === 0">
                        <td colspan="8" class="text-center py-10 text-gray-500">
                            Belum ada data barcode.
                        </td>
                    </tr>
                    <template v-else>
                        <tr
                            v-for="(barcode, index) in filteredBarcodes"
                            :key="barcode.item_id"
                            class="border-b border-gray-200 hover:bg-gray-100"
                        >
                            <td class="py-3 px-6 text-left">{{ index + 1 }}</td>
                            <td class="py-3 px-6 text-left">
                                <a
                                    href="#"
                                    @click.prevent="bukaDetail(barcode)"
                                    class="block w-full rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-center font-mono text-xs font-semibold text-emerald-800 transition hover:bg-emerald-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
                                    title="Lihat detail"
                                >
                                    {{ rentangKode(barcode) }}
                                </a>
                            </td>
                            <td class="py-3 px-6 text-left">{{ barcode.goods_partai_name || '-' }}</td>
                            <td class="py-3 px-6 text-left">
                                <span
                                    class="inline-flex rounded px-2 py-1 text-xs font-semibold"
                                    :class="
                                        barcode.item_type === 'internal'
                                            ? 'bg-emerald-100 text-emerald-800'
                                            : 'bg-blue-100 text-blue-800'
                                    "
                                >
                                    {{ barcode.item_type === 'internal' ? 'Internal' : 'External' }}
                                </span>
                            </td>
                            <td class="py-3 px-6 text-left">{{ barcode.item_year }}</td>
                            <td class="py-3 px-6 text-right">{{ barcode.total }}</td>
                            <td class="py-3 px-6 text-left whitespace-nowrap">
                                {{ formatTanggal(barcode.item_date_create) }}
                            </td>
                            <td class="py-3 px-6 text-left">
                                <button
                                    type="button"
                                    @click="hapusBarcode(barcode)"
                                    :class="[
                                        actionButtonClass,
                                        'border-red-200 bg-red-100 cursor-pointer hover:bg-red-200',
                                    ]"
                                    title="Hapus barcode"
                                    aria-label="Hapus barcode"
                                >
                                    <i class="fa-solid fa-trash-can"></i>
                                </button>
                            </td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>

        <BarcodeDetailModal :show="showDetail" :batch="detailBarcode" @close="tutupDetail" />
    </section>
</template>