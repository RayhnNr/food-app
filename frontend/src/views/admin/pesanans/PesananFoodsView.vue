<script setup>
import { ref, onMounted, watch } from 'vue';
import api from '@/api';
import { useToast } from 'vue-toast-notification';
import { useCart } from '@/composables/useCart';
import Modal from '@/components/Modal.vue';

const pesanans = ref([]);
const loading = ref(true);
const error = ref(null);
const search = ref('');
const toast = useToast();
const showModal = ref(false);
const pesananDipilih = ref(null);
const actionButtonClass =
    'inline-flex h-9 w-9 items-center justify-center rounded-md border text-gray-600 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500';

const bukaDetail = (pesanan) => {
    pesananDipilih.value = pesanan; // kirim data ke modal
    showModal.value = true;
};

const tutup = () => {
    showModal.value = false;
    pesananDipilih.value = null;
};

const getPesanan = async () => {
    loading.value = true;
    error.value = null;

    try {
        const res = await api.get('/admin/pesanans', {
            params: {
                q: search.value,
            },
        });
        pesanans.value = Array.isArray(res.data) ? res.data : [];
    } catch (e) {
        error.value = 'Gagal mengambil data pesanan. Pastikan backend aktif.';
    } finally {
        loading.value = false;
    }
};

const hapusPesanan = async (id) => {
    if (!confirm('Apakah Anda yakin ingin menghapus pesanan ini?')) {
        return;
    }

    try {
        await api.delete(`/admin/pesanans/${id}`);
        pesanans.value = pesanans.value.filter((item) => item.id !== id);
        toast.success('Pesanan berhasil dihapus');
    } catch (e) {
        toast.error('Gagal menghapus pesanan');
    }
};

const ubahStatus = async (pesanan) => {
    try {
        await api.put(`/admin/pesanans/${pesanan.id}`, {
            nama: pesanan.nama,
            no_meja: pesanan.no_meja,
            status: 'selesai',
            items: (pesanan.items || []).map((item) => ({
                id: item.id,
                jumlah: item.jumlah,
                keterangan: item.keterangan || '',
            })),
        });
        toast.success('Status pesanan berhasil diperbarui');
        await getPesanan();
    } catch (error) {
        toast.error('Gagal memperbarui status');
        console.error(error);
    }
};

function formatRupiah(angka) {
    return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
    }).format(angka);
}

const formatTanggal = (iso) => {
    if (!iso) return '-';
    const tanggal = new Date(iso);
    if (Number.isNaN(tanggal.getTime())) return '-';
    return tanggal.toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
        timeZone: 'Asia/Jakarta',
    });
};
const warnaStatus = (status) =>
    status === 'selesai' ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700';

onMounted(getPesanan);

let timer;
watch(search, () => {
    clearTimeout(timer);
    timer = setTimeout(getPesanan, 400);
});
</script>

<template>
    <section class="p-6">
        <h1 class="text-3xl font-bold mb-8">Daftar Pesanan</h1>

        <!-- Search -->
        <div class="flex flex-col md:flex-row justify-between items-center mb-6">
            <div class="w-full md:w-1/3 mb-4 md:mb-0">
                <input
                    v-model="search"
                    type="search"
                    placeholder="Cari nama pelanggan..."
                    class="w-full px-4 py-2 rounded-md border border-gray-300 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
            </div>
        </div>

        <!-- User Table -->
        <div class="overflow-x-auto bg-white rounded-lg shadow">
            <table class="w-full table-auto">
                <thead>
                    <tr class="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                        <th class="py-3 px-6 text-left">ID</th>
                        <th class="py-3 px-2 text-left">Customer</th>
                        <th class="py-3 px-6 text-left">No. Meja</th>
                        <th class="py-3 px-6 text-left">Tanggal</th>
                        <th class="py-3 px-6 text-left">Total</th>
                        <th class="py-3 px-6 text-left">Status</th>
                        <th class="py-3 px-6 text-center">Actions</th>
                    </tr>
                </thead>
                <tbody class="text-gray-600 text-sm">
                    <tr v-if="loading">
                        <td colspan="7" class="text-center py-10 text-gray-500">
                            Memuat pesanan...
                        </td>
                    </tr>
                    <tr v-else-if="error">
                        <td colspan="7" class="text-center py-10 text-red-600">
                            {{ error }}
                            <button @click="getPesanan" class="ml-2 underline">Coba lagi</button>
                        </td>
                    </tr>
                    <tr v-else-if="pesanans.length === 0">
                        <td colspan="7" class="text-center py-10 text-gray-500">
                            {{ search ? 'Pesanan tidak ditemukan.' : 'Pesanan masih kosong.' }}
                        </td>
                    </tr>
                    <tr
                        v-else
                        v-for="pesanan in pesanans"
                        :key="pesanan.id"
                        class="border-b border-gray-200 hover:bg-gray-100"
                    >
                        <td class="py-3 px-6 text-left">
                            {{ pesanan.id }}
                        </td>
                        <td class="py-3 px-2 text-left">{{ pesanan.nama }}</td>
                        <td class="py-3 px-2 text-left">{{ pesanan.no_meja }}</td>
                        <td class="py-3 px-2 text-left">{{ formatTanggal(pesanan.tanggal) }}</td>
                        <td class="py-3 px-6 text-left">{{ formatRupiah(pesanan.total) }}</td>
                        <td class="py-3 px-6 text-left">
                            <span
                                class="px-3 py-1 rounded-full text-xs font-semibold capitalize"
                                :class="warnaStatus(pesanan.status)"
                            >
                                {{ pesanan.status }}
                            </span>
                        </td>
                        <td class="py-3 px-6 text-left">
                            <div class="flex items-center justify-start gap-2">
                                <RouterLink
                                    v-if="pesanan.status !== 'selesai'"
                                    :to="`/admin/pesanans/${pesanan.id}/edit`"
                                    :class="[actionButtonClass, 'border-blue-200 bg-blue-100']"
                                    title="Edit pesanan"
                                    aria-label="Edit pesanan"
                                >
                                    <svg
                                        class="h-3.5 w-3.5"
                                        xmlns="http://www.w3.org/2000/svg"
                                        fill="none"
                                        viewBox="0 0 24 24"
                                        stroke="currentColor"
                                    >
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
                                        />
                                    </svg>
                                </RouterLink>
                                <button
                                    v-if="pesanan.status !== 'selesai'"
                                    @click="ubahStatus(pesanan)"
                                    title="Tandai selesai"
                                    aria-label="Tandai selesai"
                                    :class="[
                                        actionButtonClass,
                                        'border-emerald-200 bg-emerald-100',
                                    ]"
                                >
                                    <i class="fa-solid fa-check text-sm"></i>
                                </button>
                                <button
                                    @click="bukaDetail(pesanan)"
                                    title="Detail"
                                    aria-label="Detail pesanan"
                                    :class="[actionButtonClass, 'border-sky-200 bg-sky-100']"
                                >
                                    <i class="fa-solid fa-eye text-sm"></i>
                                </button>
                                <button
                                    v-if="pesanan.status !== 'selesai'"
                                    @click="hapusPesanan(pesanan.id)"
                                    title="Hapus pesanan"
                                    aria-label="Hapus pesanan"
                                    :class="[actionButtonClass, 'border-red-200 bg-red-100']"
                                >
                                    <i class="fa-solid fa-trash text-sm"></i>
                                </button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </section>
    <Modal :show="showModal" title="Detail Pesanan" @close="tutup">
        <div v-if="pesananDipilih">
            <p class="font-semibold">
                {{ pesananDipilih.nama }} - Meja {{ pesananDipilih.no_meja }}
            </p>

            <ul class="mt-3 space-y-1 text-sm">
                <li v-for="item in pesananDipilih.items" :key="item.id">
                    {{ item.nama }} x{{ item.jumlah }} —
                    {{ formatRupiah(item.harga * item.jumlah) }}

                    <p v-if="item.keterangan" class="text-gray-500">
                        {{ item.keterangan }}
                    </p>
                </li>
            </ul>
            <p v-if="!pesananDipilih.items?.length" class="mt-3 text-sm text-gray-500">
                Tidak ada item pada pesanan ini.
            </p>

            <p class="mt-4 font-bold text-emerald-600">
                Total: {{ formatRupiah(pesananDipilih.total) }}
            </p>
        </div>
    </Modal>
</template>
