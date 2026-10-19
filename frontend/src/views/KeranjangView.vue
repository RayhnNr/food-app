<script setup>
import { ref, onMounted, computed } from 'vue';
import api from '@/api';
import { useToast } from 'vue-toast-notification';
import { useCart } from '@/composables/useCart';

const { getCartCount } = useCart();
const keranjangs = ref([]);
const pesanans = ref([]);
const loading = ref(true);
const errorKeranjang = ref(null);
const errorPesanan = ref(null);
const tabAktif = ref('keranjang');
const toast = useToast();
const pesan = ref({
    nama: '',
    noMeja: '',
});

function formatRupiah(angka) {
    return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
    }).format(angka);
}

const totalHarga = computed(() => {
    return keranjangs.value.reduce((total, item) => {
        return total + item.product.harga * item.jumlah;
    }, 0);
});

const getKeranjang = async () => {
    try {
        const res = await api.get('/keranjangs');
        keranjangs.value = res.data;
    } catch (e) {
        console.error('Error:', e);
        console.error('Response:', e.response?.data);

        errorKeranjang.value = 'Gagal mengambil data keranjang.';
    }
};

const getPesanans = async () => {
    try {
        const res = await api.get('/pesanans');
        pesanans.value = res.data;
    } catch (e) {
        console.error('Error:', e);
        console.error('Response:', e.response?.data);

        errorPesanan.value = 'Gagal mengambil data pesanan.';
    }
};

const checkout = async () => {
    if (!pesan.value.nama.trim() || !pesan.value.noMeja || Number(pesan.value.noMeja) <= 0) {
        toast.error('Nama dan Nomor Meja harus diisi');
        return;
    }

    if (keranjangs.value.length === 0) {
        toast.error('Keranjang masih kosong, silakan pilih makanan terlebih dahulu');
        return;
    }

    try {
        await api.post('/pesanans', {
            nama: pesan.value.nama,
            no_meja: Number(pesan.value.noMeja),
        });

        keranjangs.value = [];
        pesan.value = {
            nama: '',
            noMeja: '',
        };

        await getCartCount();
        await getPesanans();

        toast.success('Pesanan berhasil dibuat');
    } catch (e) {
        toast.error(e.response?.data?.message || 'Gagal membuat pesanan');
        console.error(e);
    }
};

const warnaStatus = (status) =>
    status === 'selesai' ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700';

const hapusPesanan = async (id) => {
    if (!confirm('Apakah anda yakin ingin menghapus pesanan ini?')) {
        return;
    }

    try {
        await api.delete(`/keranjangs/${id}`);
        await getCartCount();
        keranjangs.value = keranjangs.value.filter((item) => item.id !== id);
        toast.success('Berhasil menghapus pesanan');
    } catch (e) {
        toast.error('Gagal menghapus pesanan');
    }
};

const ubahJumlah = async (keranjang, perubahan) => {
    const jumlahBaru = keranjang.jumlah + perubahan;

    if (jumlahBaru < 1) {
        toast.error('Jumlah pesanan minimal 1');
        return;
    }

    try {
        await api.patch(`/keranjangs/${keranjang.id}`, {
            jumlah: jumlahBaru,
            keterangan: keranjang.keterangan || '',
        });

        keranjang.jumlah = jumlahBaru;
    } catch (e) {
        toast.error('Gagal mengubah jumlah pesanan');
        console.error(e);
    }
};

onMounted(async () => {
    await Promise.all([getKeranjang(), getPesanans()]);
    loading.value = false;
});
</script>

<template>
    <section class="max-w-7xl mx-auto px-4 py-12 md:py-16">
        <div class="flex items-end justify-between gap-4">
            <div>
                <h2 class="text-2xl font-bold text-slate-800">Pesanan Saya</h2>
                <p class="mt-1 text-sm text-slate-600">
                    Kelola keranjang dan lihat riwayat pesanan.
                </p>
            </div>
            <RouterLink
                v-if="tabAktif === 'keranjang'"
                to="/foods"
                class="text-emerald-600 font-semibold hover:underline whitespace-nowrap"
            >
                Pesan Menu Lain <i class="fa-solid fa-arrow-right ml-1"></i>
            </RouterLink>
        </div>

        <div class="mt-8 border-b border-slate-200" role="tablist" aria-label="Pesanan Saya">
            <button
                id="tab-keranjang"
                type="button"
                role="tab"
                :aria-selected="tabAktif === 'keranjang'"
                aria-controls="panel-keranjang"
                @click="tabAktif = 'keranjang'"
                class="border-b-2 px-4 py-3 text-sm font-semibold transition"
                :class="
                    tabAktif === 'keranjang'
                        ? 'border-emerald-600 text-emerald-700'
                        : 'border-transparent text-slate-500 hover:text-slate-800'
                "
            >
                Keranjang <span class="ml-1 text-xs">({{ keranjangs.length }})</span>
            </button>
            <button
                id="tab-riwayat"
                type="button"
                role="tab"
                :aria-selected="tabAktif === 'riwayat'"
                aria-controls="panel-riwayat"
                @click="tabAktif = 'riwayat'"
                class="border-b-2 px-4 py-3 text-sm font-semibold transition"
                :class="
                    tabAktif === 'riwayat'
                        ? 'border-emerald-600 text-emerald-700'
                        : 'border-transparent text-slate-500 hover:text-slate-800'
                "
            >
                Riwayat Pesanan <span class="ml-1 text-xs">({{ pesanans.length }})</span>
            </button>
        </div>

        <div
            v-if="tabAktif === 'keranjang'"
            id="panel-keranjang"
            role="tabpanel"
            aria-labelledby="tab-keranjang"
        >
            <p v-if="loading" class="mt-6 text-slate-500">Memuat data...</p>
            <p v-else-if="errorKeranjang" class="mt-6 text-red-600">{{ errorKeranjang }}</p>

            <div v-else class="flex flex-col md:flex-row gap-6 my-4">
                <!-- Tabel keranjang -->
                <div class="w-full md:w-3/4">
                    <div class="bg-white rounded-md shadow-md p-3 sm:p-6">
                        <div class="w-full overflow-x-auto">
                            <table class="w-full min-w-225">
                                <thead>
                                    <tr>
                                        <th class="text-left font-semibold">#</th>
                                        <th class="text-left font-semibold">Makanan</th>
                                        <th class="text-left font-semibold">Keterangan</th>
                                        <th class="text-left font-semibold">Jumlah</th>
                                        <th class="text-left font-semibold">Harga</th>
                                        <th class="text-left font-semibold">Total Harga</th>
                                        <th class="text-left font-semibold">Hapus</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-if="keranjangs.length === 0">
                                        <td colspan="7" class="text-center py-10 text-gray-500">
                                            Keranjang masih kosong.
                                        </td>
                                    </tr>
                                    <template v-else>
                                        <tr
                                            v-for="(keranjang, index) in keranjangs"
                                            :key="keranjang.id"
                                        >
                                            <td class="p-4">
                                                {{ index + 1 }}
                                            </td>
                                            <td class="py-4">
                                                <div class="flex items-center">
                                                    <img
                                                        class="h-16 w-16 mr-4"
                                                        :src="`/gambar/${keranjang.product.gambar}`"
                                                        :alt="keranjang.product.nama"
                                                    />

                                                    <span class="font-semibold">
                                                        {{ keranjang.product.nama }}
                                                    </span>
                                                </div>
                                            </td>
                                            <td class="py-4">{{ keranjang.keterangan }}</td>
                                            <td class="py-4">
                                                <div class="flex items-center">
                                                    <button
                                                        @click="ubahJumlah(keranjang, -1)"
                                                        class="border rounded-md py-2 px-3 mr-2 hover:bg-emerald-500 hover:text-white"
                                                    >
                                                        <i class="fa-solid fa-minus"></i>
                                                    </button>
                                                    <span class="text-center w-8">
                                                        {{ keranjang.jumlah }}
                                                    </span>
                                                    <button
                                                        @click="ubahJumlah(keranjang, 1)"
                                                        class="border rounded-md py-2 px-3 ml-2 hover:bg-emerald-500 hover:text-white"
                                                    >
                                                        <i class="fa-solid fa-plus"></i>
                                                    </button>
                                                </div>
                                            </td>
                                            <td class="py-4">
                                                {{ formatRupiah(keranjang.product.harga) }}
                                            </td>
                                            <td class="py-4">
                                                {{
                                                    formatRupiah(
                                                        keranjang.product.harga * keranjang.jumlah,
                                                    )
                                                }}
                                            </td>
                                            <td class="py-4">
                                                <button
                                                    @click="hapusPesanan(keranjang.id)"
                                                    class="bg-red-500 hover:bg-red-600 text-white px-3 py-2 rounded-md transition"
                                                >
                                                    <i class="fa-solid fa-trash"></i>
                                                </button>
                                            </td>
                                        </tr>
                                    </template>
                                    <tr class="border-t-2 border-slate-200 bg-slate-50">
                                        <td
                                            colspan="5"
                                            class="py-4 px-4 text-right font-bold text-slate-700"
                                        >
                                            Total Belanja:
                                        </td>
                                        <td
                                            class="py-4 font-bold text-emerald-600 whitespace-nowrap"
                                        >
                                            {{ formatRupiah(totalHarga) }}
                                        </td>
                                        <td></td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>

                <!-- Summary -->
                <div class="w-full md:w-1/4">
                    <div class="bg-white rounded-lg shadow-md p-6">
                        <h2 class="text-lg font-semibold mb-4">Summary</h2>
                        <form v-on:submit.prevent class="mt-6 space-y-4">
                            <div>
                                <label
                                    for="nama"
                                    class="block text-sm font-medium text-slate-700 mb-1"
                                    >Nama</label
                                >
                                <input
                                    id="nama"
                                    v-model="pesan.nama"
                                    type="text"
                                    class="w-full px-4 py-2 rounded-sm border border-slate-300 focus:outline-none focus:ring-4 focus:ring-emerald-300"
                                />
                            </div>

                            <div>
                                <label
                                    for="nomeja"
                                    class="block text-sm font-medium text-slate-700 mb-1"
                                    >Nomor Meja</label
                                >
                                <input
                                    id="nomeja"
                                    v-model="pesan.noMeja"
                                    type="number"
                                    class="w-full px-4 py-2 rounded-sm border border-slate-300 focus:outline-none focus:ring-4 focus:ring-emerald-300"
                                />
                            </div>
                        </form>
                        <button
                            @click="checkout"
                            class="bg-emerald-600 hover:bg-emerald-700 text-white py-2 px-4 rounded-lg mt-4 w-full transition"
                        >
                            Checkout
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div v-else id="panel-riwayat" role="tabpanel" aria-labelledby="tab-riwayat" class="pt-6">
            <p v-if="loading" class="text-slate-500">Memuat data...</p>
            <p v-else-if="errorPesanan" class="text-red-600">{{ errorPesanan }}</p>
            <p v-else-if="pesanans.length === 0" class="text-slate-500">Belum ada pesanan.</p>

            <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div
                    v-for="pesanan in pesanans"
                    :key="pesanan.id"
                    class="bg-white rounded-2xl p-5 shadow-sm"
                >
                    <!-- Header -->
                    <div class="flex items-start justify-between gap-3">
                        <div>
                            <h3 class="font-semibold text-slate-800">{{ pesanan.nama }}</h3>
                            <p class="text-sm text-slate-500">
                                <i class="fa-solid fa-chair mr-1"></i>Meja {{ pesanan.no_meja }}
                            </p>
                        </div>
                        <span
                            v-if="pesanan.status"
                            class="px-3 py-1 rounded-full text-xs font-semibold capitalize"
                            :class="warnaStatus(pesanan.status)"
                        >
                            {{ pesanan.status }}
                        </span>
                    </div>

                    <!-- Daftar item -->
                    <ul class="mt-4 space-y-2 border-t border-slate-100 pt-3">
                        <li
                            v-for="item in pesanan.items"
                            :key="item.id"
                            class="flex items-start justify-between gap-3 text-sm"
                        >
                            <div>
                                <p class="font-medium text-slate-700">
                                    {{ item.nama }}
                                    <span class="text-slate-400">x{{ item.jumlah }}</span>
                                </p>

                                <p v-if="item.keterangan" class="text-xs text-slate-400">
                                    {{ item.keterangan }}
                                </p>
                            </div>

                            <span class="text-slate-600 whitespace-nowrap">
                                {{ formatRupiah(item.harga * item.jumlah) }}
                            </span>
                        </li>
                    </ul>

                    <!-- Total -->
                    <div
                        v-if="pesanan.total"
                        class="mt-4 flex items-center justify-between rounded-xl bg-emerald-50 px-4 py-2"
                    >
                        <span class="text-sm text-slate-600">Total</span>
                        <span class="font-bold text-emerald-600">{{
                            formatRupiah(pesanan.total)
                        }}</span>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>
