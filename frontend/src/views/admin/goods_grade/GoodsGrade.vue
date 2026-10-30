<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue';
import api from '@/api';
import Swal from 'sweetalert2';
import Modal from '@/components/Modal.vue';
import Select2Field from '@/components/Select2Field.vue';

const Toast = Swal.mixin({
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
});

const rows = ref([]);
const loading = ref(false);
const loadError = ref('');
const search = ref('');
const activeFilter = ref('all');
const statusOptions = [
    { id: 'all', text: 'Semua status' },
    { id: 'true', text: 'Aktif' },
    { id: 'false', text: 'Tidak aktif' },
];

const showForm = ref(false);
const editingId = ref(null);
const saving = ref(false);
const priceFocused = ref(false);
const errors = ref({});
const formError = ref('');
const actionButtonClass =
    'inline-flex h-9 w-9 items-center justify-center rounded-md border text-gray-600 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500';

const emptyForm = () => ({
    goods_grade_name: '',
    goods_grade_desc: '',
    goods_grade_price: '',
    goods_grade_active: true,
});
const form = reactive(emptyForm());

const isEdit = computed(() => editingId.value !== null);

const formatRupiah = (value) => `Rp ${Number(value).toLocaleString('id-ID')}`;
const formatPriceInput = (value) => {
    const text = String(value ?? '');
    if (!text) return 'Rp ';

    const [integerPart, fractionPart] = text.split('.');
    const integer = Number(integerPart || 0).toLocaleString('id-ID', {
        maximumFractionDigits: 0,
    });
    return `Rp ${integer}${text.includes('.') ? `,${fractionPart ?? ''}` : ''}`;
};

const parseFormattedPrice = (value) => {
    const text = String(value).trim();
    if (!/^Rp\s*/i.test(text)) return text.replace(',', '.');
    return text
        .replace(/^Rp\s*/i, '')
        .replace(/\./g, '')
        .replace(',', '.')
        .trim();
};

const focusPrice = () => {
    if (priceFocused.value) return;
    const value = String(form.goods_grade_price ?? '').trim();
    form.goods_grade_price = value ? parseFormattedPrice(value) : '';
    priceFocused.value = true;
};

const inputPrice = (event) => {
    const input = event.target.value;
    const cursorPosition = event.target.selectionStart ?? input.length;
    const textBeforeCursor = input.slice(0, cursorPosition);
    const digitsBeforeCursor = textBeforeCursor.replace(/\D/g, '').length;
    const hasDecimalBeforeCursor = textBeforeCursor.includes(',');
    const fractionDigitsBeforeCursor = textBeforeCursor.split(',').at(-1).replace(/\D/g, '').length;
    const value = input.replace(/^Rp\s*/i, '').replace(/[^\d,.]/g, '');
    const hasDecimal = value.includes(',');
    const [integerPart, ...fractionParts] = value.split(',');
    const integerDigits = integerPart.replace(/\D/g, '');
    const fractionDigits = fractionParts.join('').replace(/\D/g, '');
    const numericValue = `${integerDigits}${hasDecimal ? `.${fractionDigits}` : ''}`;
    const formattedValue = formatPriceInput(numericValue);

    form.goods_grade_price = numericValue;
    event.target.value = formattedValue;

    let nextCursorPosition = formattedValue.indexOf('Rp ') + 3;
    if (hasDecimalBeforeCursor) {
        nextCursorPosition = formattedValue.indexOf(',') + 1 + fractionDigitsBeforeCursor;
    } else if (digitsBeforeCursor > 0) {
        let digitsSeen = 0;
        for (let index = 0; index < formattedValue.length; index += 1) {
            if (/\d/.test(formattedValue[index])) digitsSeen += 1;
            if (digitsSeen === digitsBeforeCursor) {
                nextCursorPosition = index + 1;
                break;
            }
        }
    }
    event.target.setSelectionRange(nextCursorPosition, nextCursorPosition);
};

const blurPrice = () => {
    const value = String(form.goods_grade_price ?? '')
        .trim()
        .replace(',', '.');
    const amount = Number(value);
    if (value && Number.isFinite(amount) && amount >= 0) {
        form.goods_grade_price = formatRupiah(amount);
    }
    priceFocused.value = false;
};

const load = async () => {
    loading.value = true;
    loadError.value = '';

    try {
        const params = {};
        if (search.value.trim()) params.q = search.value.trim();
        if (activeFilter.value !== 'all') params.active = activeFilter.value;

        const res = await api.get('/goods-grade', { params });
        rows.value = res.data.data;
    } catch (e) {
        loadError.value = e.response?.data?.message || 'Gagal mengambil data grade.';
        console.error(e);
    } finally {
        loading.value = false;
    }
};

let searchTimer;
watch(search, () => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(load, 300);
});
watch(activeFilter, load);

const resetErrors = () => {
    errors.value = {};
    formError.value = '';
};

const openCreate = () => {
    editingId.value = null;
    Object.assign(form, emptyForm());
    priceFocused.value = false;
    resetErrors();
    showForm.value = true;
};

const openEdit = (row) => {
    editingId.value = row.goods_grade_id;
    Object.assign(form, {
        goods_grade_name: row.goods_grade_name,
        goods_grade_desc: row.goods_grade_desc,
        goods_grade_price: row.goods_grade_price == null ? '' : formatRupiah(row.goods_grade_price),
        goods_grade_active: row.goods_grade_active,
    });
    priceFocused.value = false;
    resetErrors();
    showForm.value = true;
};

const closeForm = () => {
    if (saving.value) return;
    showForm.value = false;
};

const save = async () => {
    if (saving.value) return;

    resetErrors();

    const priceValue = priceFocused.value
        ? String(form.goods_grade_price).trim().replace(',', '.')
        : parseFormattedPrice(form.goods_grade_price);
    const validPrice = /^(?:\d+(?:\.\d*)?|\.\d+)$/.test(priceValue);
    if (priceValue && (!validPrice || !Number.isFinite(Number(priceValue)))) {
        errors.value.goods_grade_price = 'Harga hanya boleh berisi angka.';
        return;
    }

    saving.value = true;

    const payload = {
        ...form,
        goods_grade_price: priceValue === '' ? null : Number(priceValue),
    };

    try {
        if (isEdit.value) {
            await api.put(`/goods-grade/${editingId.value}`, payload);
        } else {
            await api.post('/goods-grade', payload);
        }

        showForm.value = false;
        await load();
        Toast.fire({
            icon: 'success',
            title: isEdit.value ? 'Berhasil mengubah grade' : 'Berhasil menambahkan grade',
        });
    } catch (e) {
        const data = e.response?.data;
        errors.value = data?.errors || {};
        formError.value = data?.errors ? '' : data?.message || 'Gagal menyimpan grade.';
        console.error(e);
    } finally {
        saving.value = false;
    }
};

const remove = async (row) => {
    const result = await Swal.fire({
        title: 'Yakin hapus grade ini?',
        text: row.goods_grade_name,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Ya, hapus',
        cancelButtonText: 'Batal',
    });

    if (!result.isConfirmed) return;

    try {
        await api.delete(`/goods-grade/${row.goods_grade_id}`);
        await load();
        Toast.fire({ icon: 'success', title: 'Berhasil menghapus grade' });
    } catch (e) {
        Swal.fire({
            icon: 'error',
            title: 'Gagal menghapus',
            text: e.response?.data?.message || 'Gagal menghapus grade',
        });
        console.error(e);
    }
};

onMounted(load);
</script>

<template>
    <section class="p-6">
        <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
            <h1 class="text-3xl font-bold">Goods Grade</h1>
            <button
                type="button"
                @click="openCreate"
                class="rounded-md bg-emerald-600 px-4 py-2 text-white transition hover:bg-emerald-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2"
            >
                <i class="fa-solid fa-plus mr-1"></i> Tambah Grade
            </button>
        </div>

        <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <input
                v-model="search"
                type="search"
                placeholder="Cari nama grade..."
                aria-label="Cari nama grade"
                class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 focus:outline-none focus:ring-2 focus:ring-emerald-500 sm:max-w-sm"
            />
            <div class="w-full sm:w-52 sm:flex-none">
                <Select2Field
                    v-model="activeFilter"
                    :options="statusOptions"
                    aria-label="Filter status"
                    placeholder="Semua status"
                />
            </div>
        </div>

        <div class="overflow-hidden rounded-lg bg-white shadow">
            <div class="overflow-x-auto p-4">
                <p v-if="loadError" role="alert" class="mb-4 text-sm text-red-600">
                    {{ loadError }}
                </p>

                <table class="w-full min-w-180 table-auto text-sm">
                    <thead>
                        <tr class="bg-gray-200 text-sm uppercase leading-normal text-gray-600">
                            <th class="px-4 py-3 text-center">#</th>
                            <th class="px-4 py-3 text-left">Nama Grade</th>
                            <th class="px-4 py-3 text-left">Deskripsi</th>
                            <th class="px-4 py-3 text-right">Harga</th>
                            <th class="px-4 py-3 text-center">Status</th>
                            <th class="px-4 py-3 text-center">Aksi</th>
                        </tr>
                    </thead>
                    <tbody class="text-sm text-gray-600">
                        <tr v-if="loading">
                            <td colspan="6" class="py-6 text-center text-gray-500">
                                Memuat data...
                            </td>
                        </tr>
                        <tr v-else-if="rows.length === 0">
                            <td colspan="6" class="py-6 text-center text-gray-500">
                                Belum ada data.
                            </td>
                        </tr>
                        <template v-else>
                            <tr
                                v-for="(row, index) in rows"
                                :key="row.goods_grade_id"
                                class="border-b border-gray-200 transition hover:bg-gray-50"
                            >
                                <td class="px-4 py-3 text-center">
                                    {{ index + 1 }}
                                </td>
                                <td class="px-4 py-3 font-medium text-gray-800">
                                    {{ row.goods_grade_name }}
                                </td>
                                <td class="max-w-sm px-4 py-3">
                                    {{ row.goods_grade_desc || '-' }}
                                </td>
                                <td class="whitespace-nowrap px-4 py-3 text-right tabular-nums">
                                    {{ formatRupiah(row.goods_grade_price) }}
                                </td>
                                <td class="px-4 py-3 text-center">
                                    <span
                                        class="inline-flex rounded px-2 py-1 text-xs font-semibold"
                                        :class="
                                            row.goods_grade_active
                                                ? 'bg-emerald-100 text-emerald-800'
                                                : 'bg-slate-100 text-slate-600'
                                        "
                                    >
                                        {{ row.goods_grade_active ? 'Aktif' : 'Nonaktif' }}
                                    </span>
                                </td>
                                <td class="px-4 py-3 text-center">
                                    <div class="flex items-center justify-center gap-2">
                                        <button
                                            type="button"
                                            @click="openEdit(row)"
                                            :class="[
                                                actionButtonClass,
                                                'border-blue-200 bg-blue-100 cursor-pointer',
                                            ]"
                                            title="Ubah"
                                            aria-label="Ubah"
                                        >
                                            <i class="fa-solid fa-pencil"></i>
                                        </button>
                                        <button
                                            type="button"
                                            @click="remove(row)"
                                            :class="[
                                                actionButtonClass,
                                                'border-red-200 bg-red-100 cursor-pointer',
                                            ]"
                                            title="Hapus"
                                            aria-label="Hapus"
                                        >
                                            <i class="fa-solid fa-trash-can"></i>
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        </template>
                    </tbody>
                </table>
            </div>
        </div>

        <Modal :show="showForm" :title="isEdit ? 'Ubah Grade' : 'Tambah Grade'" @close="closeForm">
            <form class="space-y-4" @submit.prevent="save">
                <div>
                    <label for="grade-name" class="mb-1 block text-sm font-medium text-slate-700">
                        Nama Grade
                    </label>
                    <input
                        id="grade-name"
                        v-model="form.goods_grade_name"
                        type="text"
                        maxlength="256"
                        required
                        class="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                    <p v-if="errors.goods_grade_name" class="mt-1 text-sm text-red-600">
                        {{ errors.goods_grade_name }}
                    </p>
                </div>

                <div>
                    <label for="grade-price" class="mb-1 block text-sm font-medium text-slate-700">
                        Harga
                    </label>
                    <input
                        id="grade-price"
                        :value="
                            priceFocused
                                ? formatPriceInput(form.goods_grade_price)
                                : form.goods_grade_price
                        "
                        type="text"
                        inputmode="decimal"
                        autocomplete="off"
                        placeholder="Rp 0"
                        @focus="focusPrice"
                        @input="inputPrice"
                        @blur="blurPrice"
                        class="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                    <p v-if="errors.goods_grade_price" class="mt-1 text-sm text-red-600">
                        {{ errors.goods_grade_price }}
                    </p>
                </div>

                <div>
                    <label for="grade-desc" class="mb-1 block text-sm font-medium text-slate-700">
                        Deskripsi
                    </label>
                    <textarea
                        id="grade-desc"
                        v-model="form.goods_grade_desc"
                        rows="3"
                        class="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    ></textarea>
                </div>

                <div>
                    <span class="mb-2 block text-sm font-medium text-slate-700">Status</span>
                    <label class="inline-flex cursor-pointer items-center gap-3">
                        <input
                            v-model="form.goods_grade_active"
                            type="checkbox"
                            role="switch"
                            class="peer sr-only"
                        />
                        <span
                            class="relative h-6 w-11 rounded-full bg-gray-300 transition peer-checked:bg-emerald-600 peer-focus-visible:ring-2 peer-focus-visible:ring-emerald-500 after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:transition peer-checked:after:translate-x-5"
                        ></span>
                        <span class="text-sm text-gray-600">
                            {{ form.goods_grade_active ? 'Aktif' : 'Nonaktif' }}
                        </span>
                    </label>
                </div>

                <p v-if="formError" role="alert" class="text-sm text-red-600">
                    {{ formError }}
                </p>
                <div class="flex justify-end gap-2 pt-2">
                    <button
                        type="button"
                        @click="closeForm"
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
                        <i class="fa-solid fa-floppy-disk mr-1"></i>
                        {{ saving ? 'Menyimpan...' : 'Simpan' }}
                    </button>
                </div>
            </form>
        </Modal>
    </section>
</template>
