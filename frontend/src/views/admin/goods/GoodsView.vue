<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import api from '@/api';
import Swal from 'sweetalert2';
import Select2Field from '@/components/Select2Field.vue';

const rows = ref([]);
const categories = ref([{ id: '', text: 'Semua kategori' }]);
const loading = ref(false);
const loadError = ref('');
const categoryError = ref('');
const search = ref('');
const activeFilter = ref('all');
const categoryFilter = ref('');
const actionButtonClass =
    'inline-flex h-9 w-9 items-center justify-center rounded-md border text-gray-600 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500';

const statusOptions = [
    { id: 'all', text: 'Semua status' },
    { id: 'true', text: 'Aktif' },
    { id: 'false', text: 'Nonaktif' },
];

const loadCategories = async () => {
    try {
        const response = await api.get('/goods-category-options');
        categories.value = [
            { id: '', text: 'Semua kategori' },
            ...response.data.data.map((category) => ({
                id: String(category.goods_category_id),
                text: category.goods_category_name,
            })),
        ];
    } catch (error) {
        categoryError.value = error.response?.data?.message || 'Gagal mengambil kategori barang.';
        console.error('Gagal mengambil kategori barang:', error);
    }
};

const loadGoods = async () => {
    loading.value = true;
    loadError.value = '';

    try {
        const params = {};
        if (search.value.trim()) params.q = search.value.trim();
        if (activeFilter.value !== 'all') params.active = activeFilter.value;
        if (categoryFilter.value) params.category_id = categoryFilter.value;

        const response = await api.get('/goods', { params });
        rows.value = response.data.data;
    } catch (error) {
        loadError.value = error.response?.data?.message || 'Gagal mengambil data barang.';
        console.error('Gagal mengambil data barang:', error);
    } finally {
        loading.value = false;
    }
};

const deleteGoods = async (goods) => {
    const result = await Swal.fire({
        title: `Yakin hapus ${goods.goods_name}?`,
        text: 'Data yang sudah dihapus tidak dapat dikembalikan.',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Ya, hapus',
        cancelButtonText: 'Batal',
        confirmButtonColor: '#dc2626',
    });

    if (!result.isConfirmed) return;

    try {
        await api.delete(`/goods/${goods.goods_id}`);
        await loadGoods();
        await Swal.fire({ icon: 'success', title: 'Barang berhasil dihapus' });
    } catch (error) {
        await Swal.fire({
            icon: 'error',
            title: 'Gagal menghapus barang',
            text: error.response?.data?.message || 'Terjadi kesalahan saat menghapus barang.',
        });
        console.error('Gagal menghapus barang:', error);
    }
};

let searchTimer;
watch(search, () => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(loadGoods, 350);
});
watch([activeFilter, categoryFilter], loadGoods);

onMounted(() => {
    loadCategories();
    loadGoods();
});

onBeforeUnmount(() => clearTimeout(searchTimer));
</script>

<template>
    <section class="min-h-[calc(100vh-4rem)] bg-slate-100 p-4 md:p-6">
        <div class="mx-auto max-w-7xl">
            <header class="mb-6">
                <p class="text-xs font-semibold uppercase tracking-wide text-emerald-700">
                    Inventaris
                </p>
                <h1 class="mt-1 text-2xl font-bold text-slate-900">Data Barang</h1>
            </header>

            <p
                v-if="loadError"
                role="alert"
                class="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
            >
                {{ loadError }}
            </p>
            <p
                v-if="categoryError"
                role="alert"
                class="mb-4 rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800"
            >
                {{ categoryError }}
            </p>

            <div class="mb-5 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                <div class="grid w-full gap-3 sm:grid-cols-2 lg:max-w-3xl lg:grid-cols-3">
                    <input
                        v-model="search"
                        type="search"
                        placeholder="Cari nama barang..."
                        aria-label="Cari nama barang"
                        class="h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/15"
                    />
                    <Select2Field
                        v-model="activeFilter"
                        :options="statusOptions"
                        aria-label="Filter status"
                    />
                    <Select2Field
                        v-model="categoryFilter"
                        :options="categories"
                        aria-label="Filter kategori"
                    />
                </div>
                <RouterLink
                    to="/admin/goods/tambah"
                    class="inline-flex h-10 shrink-0 items-center justify-center gap-2 rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800"
                >
                    <i class="fa-solid fa-plus" aria-hidden="true"></i>
                    Tambah Barang
                </RouterLink>
            </div>

            <div class="overflow-x-auto rounded-md border border-slate-200 bg-white shadow-sm">
                <table class="w-full min-w-[760px] table-auto">
                    <thead>
                        <tr class="bg-slate-100 text-xs font-semibold uppercase tracking-wide text-slate-600">
                            <th class="px-5 py-3 text-left">#</th>
                            <th class="px-5 py-3 text-left">Nama Barang</th>
                            <th class="px-5 py-3 text-left">Kategori</th>
                            <th class="px-5 py-3 text-left">Deskripsi</th>
                            <th class="px-5 py-3 text-left">Status</th>
                            <th class="px-5 py-3 text-left">Aksi</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100 text-sm text-slate-700">
                        <tr v-if="loading">
                            <td colspan="6" class="px-5 py-10 text-center text-slate-500">
                                Memuat data barang...
                            </td>
                        </tr>
                        <tr v-else-if="rows.length === 0">
                            <td colspan="6" class="px-5 py-10 text-center text-slate-500">
                                Belum ada data barang.
                            </td>
                        </tr>
                        <tr
                            v-for="(goods, index) in rows"
                            v-else
                            :key="goods.goods_id"
                            class="transition hover:bg-slate-50"
                        >
                            <td class="px-5 py-3">{{ index + 1 }}</td>
                            <td class="px-5 py-3 font-medium text-slate-900">
                                {{ goods.goods_name }}
                            </td>
                            <td class="px-5 py-3">{{ goods.goods_category_name || '-' }}</td>
                            <td class="max-w-sm truncate px-5 py-3" :title="goods.goods_desc">
                                {{ goods.goods_desc || '-' }}
                            </td>
                            <td class="px-5 py-3">
                                <span
                                    class="inline-flex rounded px-2 py-1 text-xs font-semibold"
                                    :class="
                                        goods.goods_active
                                            ? 'bg-emerald-100 text-emerald-800'
                                            : 'bg-slate-100 text-slate-600'
                                    "
                                >
                                    {{ goods.goods_active ? 'Aktif' : 'Nonaktif' }}
                                </span>
                            </td>
                            <td class="px-5 py-3">
                                <div class="flex items-center gap-2">
                                    <RouterLink
                                        :to="`/admin/goods/${goods.goods_id}/edit`"
                                        :class="[actionButtonClass, 'border-blue-200 bg-blue-100']"
                                        :title="`Ubah ${goods.goods_name}`"
                                        :aria-label="`Ubah ${goods.goods_name}`"
                                    >
                                        <i class="fa-solid fa-pencil" aria-hidden="true"></i>
                                    </RouterLink>
                                    <button
                                        type="button"
                                        :class="[
                                            actionButtonClass,
                                            'cursor-pointer border-red-200 bg-red-100',
                                        ]"
                                        :title="`Hapus ${goods.goods_name}`"
                                        :aria-label="`Hapus ${goods.goods_name}`"
                                        @click="deleteGoods(goods)"
                                    >
                                        <i class="fa-solid fa-trash-can" aria-hidden="true"></i>
                                    </button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </section>
</template>
