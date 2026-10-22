<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue';
import api from '@/api';
import Swal from 'sweetalert2';
import { useCart } from '@/composables/useCart';
import Modal from '@/components/Modal.vue';

const categories = ref([]);
const loading = ref(true);
const error = ref(null);
const search = ref('');
const showModal = ref(false);
const saving = ref(false);
const editingCategory = ref(null);
const form = ref({ nama: '', deskripsi: '' });
const formError = ref('');
const { getCartCount } = useCart();
const actionButtonClass =
    'inline-flex h-9 w-9 items-center justify-center rounded-md border text-gray-600 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500';

// Toast kecil ala SweetAlert
const Toast = Swal.mixin({
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
});

const bukaModal = (category = null) => {
    editingCategory.value = category;
    form.value = category
        ? { nama: category.nama, deskripsi: category.deskripsi || '' }
        : { nama: '', deskripsi: '' };
    formError.value = '';
    showModal.value = true;
};

const tutupModal = () => {
    if (saving.value) return;
    showModal.value = false;
};

const simpanCategory = async () => {
    if (saving.value) return;

    const nama = form.value.nama.trim();
    if (nama.length < 3) {
        formError.value = 'Nama kategori minimal 3 karakter.';
        return;
    }

    saving.value = true;
    formError.value = '';
    const isEdit = Boolean(editingCategory.value);

    try {
        const data = { nama, deskripsi: form.value.deskripsi.trim() };
        if (isEdit) {
            await api.put(`/category/${editingCategory.value.id}`, data);
        } else {
            await api.post('/category', data);
        }

        showModal.value = false;
        await getCategory();
        Toast.fire({
            icon: 'success',
            title: isEdit ? 'Berhasil memperbarui kategori' : 'Berhasil menambahkan kategori',
        });
    } catch (e) {
        formError.value = e.response?.data?.errors?.nama || e.response?.data?.message || 'Gagal menyimpan kategori.';
        console.error(e);
    } finally {
        saving.value = false;
    }
};

const getCategory = async () => {
    loading.value = true;
    error.value = null;

    try {
        const res = await api.get('/category', {
            params: {
                q: search.value,
                _sort: 'id',
                _order: 'DESC',
            },
        });
        categories.value = res.data.data;
    } catch (e) {
        error.value = 'Gagal mengambil data. Pastikan npm run api sudah jalan.';
        console.error(e);
    } finally {
        loading.value = false;
    }
};

const hapuscategory = async (id) => {
    const result = await Swal.fire({
        title: 'Yakin hapus category ini?',
        text: 'Data tidak bisa dikembalikan',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Ya, hapus',
        cancelButtonText: 'Batal',
    });

    if (!result.isConfirmed) {
        return;
    }

    try {
        await api.delete(`/category/${id}`);
        await getCartCount();
        categories.value = categories.value.filter((item) => item.id !== id);
        Toast.fire({ icon: 'success', title: 'Berhasil menghapus category' });
    } catch (e) {
        Toast.fire({ icon: 'error', title: 'Gagal menghapus category' });
        console.error(e);
    }
};

onMounted(getCategory);

let timer;
watch(search, () => {
    clearTimeout(timer);
    timer = setTimeout(getCategory, 400);
});

onBeforeUnmount(() => {
    clearTimeout(timer);
});
</script>

<template>
    <section class="p-6">
        <h1 class="text-3xl font-bold mb-8">Data Category</h1>

        <div class="flex flex-col md:flex-row justify-between items-center mb-6">
            <div class="w-full md:w-1/3 mb-4 md:mb-0">
                <input
                    v-model="search"
                    type="text"
                    placeholder="Cari kategori..."
                    class="w-full px-4 py-2 rounded-md border border-gray-300 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
            </div>
            <button
                type="button"
                @click="bukaModal()"
                class="px-4 py-2 rounded-xl bg-emerald-600 text-white hover:bg-emerald-700"
            >
                <i class="fa-solid fa-plus mr-1"></i> Tambah
            </button>
        </div>

        <!-- User Table -->
        <div class="overflow-x-auto bg-white rounded-lg shadow">
            <table class="w-full table-auto">
                <thead>
                    <tr class="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                        <th class="py-3 px-6 text-left">ID</th>
                        <th class="py-3 px-2 text-left">Category</th>
                        <th class="py-3 px-6 text-left">Deskripsi</th>
                        <th class="py-3 w-25 text-center">Actions</th>
                    </tr>
                </thead>
                <tbody class="text-gray-600 text-sm">
                    <tr v-if="categories.length === 0">
                        <td colspan="7" class="text-center py-10 text-gray-500">
                            Category masih kosong.
                        </td>
                    </tr>
                    <tr
                        v-else
                        v-for="(category, index) in categories"
                        :key="category.id"
                        class="border-b border-gray-200 hover:bg-gray-100"
                    >
                        <td class="py-3 px-6 text-left">
                            {{ index + 1 }}
                        </td>
                        <td class="py-3 px-2 text-left">{{ category.nama }}</td>
                        <td class="py-3 px-2 text-left">{{ category.deskripsi }}</td>
                        <td class="py-3 px-6 text-left">
                            <div class="flex items-center justify-start gap-2">
                                <button
                                    type="button"
                                    @click="bukaModal(category)"
                                    :class="[actionButtonClass, 'border-blue-200 bg-blue-100']"
                                    title="Edit kategori"
                                    aria-label="Edit kategori"
                                >
                                    <i class="fa-solid fa-pencil"></i>
                                </button>
                                <button
                                    @click="hapuscategory(category.id)"
                                    :class="[
                                        actionButtonClass,
                                        'border-red-200 bg-red-100 cursor-pointer',
                                    ]"
                                    title="Hapus menu"
                                    aria-label="Hapus menu"
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

    <Modal
        :show="showModal"
        :title="editingCategory ? 'Edit Kategori' : 'Tambah Kategori'"
        @close="tutupModal"
    >
        <form class="space-y-4" @submit.prevent="simpanCategory">
            <div>
                <label for="category-name" class="mb-1 block text-sm font-medium text-slate-700">
                    Nama
                </label>
                <input
                    id="category-name"
                    v-model="form.nama"
                    type="text"
                    required
                    minlength="3"
                    maxlength="50"
                    class="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    placeholder="Nama kategori"
                />
            </div>
            <div>
                <label for="category-description" class="mb-1 block text-sm font-medium text-slate-700">
                    Deskripsi
                </label>
                <textarea
                    id="category-description"
                    v-model="form.deskripsi"
                    rows="3"
                    class="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    placeholder="Deskripsi kategori"
                ></textarea>
            </div>
            <p v-if="formError" role="alert" class="text-sm text-red-600">{{ formError }}</p>
            <div class="flex justify-end gap-2 pt-2">
                <button
                    type="button"
                    @click="tutupModal"
                    :disabled="saving"
                    class="rounded-md border border-gray-300 px-4 py-2 text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                >
                    Batal
                </button>
                <button
                    type="submit"
                    :disabled="saving"
                    class="rounded-md bg-emerald-600 px-4 py-2 text-white hover:bg-emerald-700 disabled:opacity-50"
                >
                    {{ saving ? 'Menyimpan...' : 'Simpan' }}
                </button>
            </div>
        </form>
    </Modal>
</template>
