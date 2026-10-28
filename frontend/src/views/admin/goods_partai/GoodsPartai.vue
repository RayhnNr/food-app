<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue';
import api from '@/api';
import Swal from 'sweetalert2';

const goodspartais = ref([]);
const loading = ref(true);
const error = ref(null);
const search = ref('');
const actionButtonClass =
    'inline-flex h-9 w-9 items-center justify-center rounded-md border text-gray-600 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500';

const getGoodsPartai = async () => {
    loading.value = true;
    error.value = null;

    try {
        const res = await api.get('/goods-partai', {
            params: {
                q: search.value,
                _sort: 'goods_partai_id',
                _order: 'DESC',
            },
        });
        goodspartais.value = res.data.data;
    } catch (e) {
        error.value = 'Gagal mengambil data. Pastikan npm run api sudah jalan.';
        console.error(e);
    } finally {
        loading.value = false;
    }
};

const hapusPartai = async (goods) => {
    const result = await Swal.fire({
        title: `Yakin hapus ${goods.goods_partai_code}?`,
        text: 'Data tidak bisa dikembalikan',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Ya, hapus',
        cancelButtonText: 'Batal',
    });

    if (!result.isConfirmed) return;

    try {
        await api.delete(`/goods-partai/${goods.goods_partai_id}`);
        await getGoodsPartai();
        Swal.fire({ icon: 'success', title: 'Berhasil menghapus Partai' });
    } catch (e) {
        Swal.fire({
            icon: 'error',
            title: 'Gagal menghapus',
            text: e.response?.data?.message || 'Gagal menghapus Partai',
        });
        console.error(e);
    }
};

onMounted(getGoodsPartai);

let timer;
watch(search, () => {
    clearTimeout(timer);
    timer = setTimeout(getGoodsPartai, 400);
});

onBeforeUnmount(() => {
    clearTimeout(timer);
});
</script>

<template>
    <section class="p-6">
        <h1 class="text-3xl font-bold mb-8">Data Partai Barang</h1>

        <div class="flex flex-col md:flex-row justify-between items-center mb-6">
            <div class="w-full md:w-1/3 mb-4 md:mb-0">
                <input
                    v-model="search"
                    type="text"
                    placeholder="Cari kode atau nama partai..."
                    class="w-full px-4 py-2 rounded-md border border-gray-300 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
            </div>
            <RouterLink
                to="/admin/goods-partai/tambah"
                class="px-4 py-2 rounded-xl object-cover bg-emerald-600 text-white hover:bg-emerald-700"
            >
                <i class="fa-solid fa-plus mr-1"></i> Tambah
            </RouterLink>
        </div>

        <!-- User Table -->
        <div class="overflow-x-auto bg-white rounded-lg shadow">
            <table class="w-full table-auto">
                <thead>
                    <tr class="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                        <th class="py-3 px-6 text-left">#</th>
                        <th class="py-3 px-6 text-left">Kode Partai</th>
                        <th class="py-3 px-6 text-left">Nama Partai</th>
                        <th class="py-3 px-6 text-left">Tahun</th>
                        <th class="py-3 px-6 text-left">Kemasan</th>
                        <th class="py-3 px-6 text-left">Deskripsi</th>
                        <th class="py-3 px-6 text-left">Status</th>
                        <th class="py-3 px-6 text-left">Aksi</th>
                    </tr>
                </thead>
                <tbody class="text-gray-600 text-sm">
                    <tr v-if="goodspartais.length === 0">
                        <td colspan="8" class="text-center py-10 text-gray-500">
                            Belum ada data partai barang.
                        </td>
                    </tr>
                    <tr
                        v-else
                        v-for="(goods, index) in goodspartais"
                        :key="goods.goods_partai_id"
                        class="border-b border-gray-200 hover:bg-gray-100"
                    >
                        <td class="py-3 px-6 text-left">
                            {{ index + 1 }}
                        </td>
                        <!-- <td>{{ String(goods.goods_partai_code).padStart(3, '0') }}</td> -->
                        <td class="py-3 px-6 text-left">
                            {{ String(goods.goods_partai_code).padStart(3, '0') }}
                        </td>
                        <td class="py-3 px-6 text-left">{{ goods.goods_partai_name }}</td>
                        <td class="py-3 px-6 text-left">{{ goods.goods_partai_year }}</td>
                        <td class="py-3 px-6 text-left">{{ goods.goods_partai_bungkus }}</td>
                        <td class="py-3 px-6 text-left">{{ goods.goods_partai_desc }}</td>
                        <td class="py-3 px-6 text-left">
                            <span
                                class="inline-flex rounded px-2 py-1 text-xs font-semibold"
                                :class="
                                    goods.goods_partai_active
                                        ? 'bg-emerald-100 text-emerald-800'
                                        : 'bg-slate-100 text-slate-600'
                                "
                            >
                                {{ goods.goods_partai_active ? 'Aktif' : 'Nonaktif' }}
                            </span>
                        </td>
                        <td class="py-3 px-6 text-left">
                            <div class="flex items-center justify-start gap-2">
                                <RouterLink
                                    :to="`/admin/goods-partai/${goods.goods_partai_id}/edit`"
                                    :class="[actionButtonClass, 'border-blue-200 bg-blue-100']"
                                    title="Ubah partai"
                                    aria-label="Ubah partai"
                                >
                                    <i class="fa-solid fa-pencil"></i>
                                </RouterLink>
                                <RouterLink
                                    :to="`/admin/goods-partai/${goods.goods_partai_id}/tarra-decimal`"
                                    :class="[actionButtonClass, 'border-amber-200 bg-amber-100']"
                                    title="Tarra & Desimal"
                                    aria-label="Tarra dan desimal"
                                >
                                    <i class="fa-solid fa-scale-balanced"></i>
                                </RouterLink>
                                <button
                                    type="button"
                                    @click="hapusPartai(goods)"
                                    :class="[
                                        actionButtonClass,
                                        'border-red-200 bg-red-100 cursor-pointer',
                                    ]"
                                    title="Hapus Partai"
                                    aria-label="Hapus Partai"
                                >
                                    <i class="fa-solid fa-trash-can"></i>
                                </button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </section>
</template>
