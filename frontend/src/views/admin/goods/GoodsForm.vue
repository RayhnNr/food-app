<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import Swal from 'sweetalert2';
import api from '@/api';
import Select2Field from '@/components/Select2Field.vue';

const route = useRoute();
const router = useRouter();
const goodsId = computed(() => route.params.id);
const isEditing = computed(() => Boolean(goodsId.value));
const form = ref(emptyForm());
const categories = ref([]);
const coaGoods = ref([]);
const coaOptions = ref([]);
const loading = ref(false);
const submitting = ref(false);
const fieldErrors = ref({});
const pageError = ref('');

function emptyForm() {
    return {
        goods_name: '',
        goods_desc: '',
        goods_active: true,
        goods_category_id: '',
        goods_brand_id: '',
        goods_unit_id: '',
        coas: [],
    };
}

function fillMappings(rows = []) {
    form.value.coas = rows.map((row) => ({
        coa_goods_id: Number(row.coa_goods_id),
        coa_goods_name: row.coa_goods_name,
        coa_id: row.coa_id == null ? '' : String(row.coa_id),
    }));
}

function clearFieldError(field) {
    delete fieldErrors.value[field];
    pageError.value = '';
}

function validateForm() {
    const errors = {};
    const name = form.value.goods_name.trim();

    if (!name) {
        errors.goods_name = 'Nama barang wajib diisi.';
    } else if (name.length > 256) {
        errors.goods_name = 'Nama barang maksimal 256 karakter.';
    }

    for (const field of ['goods_brand_id', 'goods_unit_id']) {
        const value = form.value[field];
        if (value !== '' && (!Number.isInteger(Number(value)) || Number(value) <= 0)) {
            errors[field] = 'Masukkan ID berupa bilangan bulat positif.';
        }
    }

    fieldErrors.value = errors;
    return Object.keys(errors).length === 0;
}

async function fetchCoaOptions() {
    const response = await api.get('/coa');
    return [
        { id: '', text: '' },
        ...response.data.data
            .filter((coa) => String(coa.coa_no || '').trim().endsWith(',00'))
            .map((coa) => ({
                id: String(coa.coa_id),
                text: `${coa.coa_no} - ${coa.coa_name}`,
            })),
    ];
}

async function addCategory() {
    const result = await Swal.fire({
        title: 'Tambah kategori barang',
        input: 'text',
        inputLabel: 'Nama kategori',
        inputPlaceholder: 'Masukkan nama kategori',
        showCancelButton: true,
        confirmButtonText: 'Simpan',
        cancelButtonText: 'Batal',
        confirmButtonColor: '#059669',
        inputValidator: (value) => {
            const name = value.trim();
            if (!name) return 'Nama kategori wajib diisi.';
            if (name.length > 256) return 'Nama kategori maksimal 256 karakter.';
            return undefined;
        },
    });

    if (!result.isConfirmed) return;

    try {
        const response = await api.post('/goods-categories', {
            goods_category_name: result.value.trim(),
        });
        const category = response.data.data;
        categories.value = [
            ...categories.value.filter((option) => option.id !== ''),
            { id: String(category.goods_category_id), text: category.goods_category_name },
        ];
        categories.value.sort((a, b) => a.text.localeCompare(b.text, 'id'));
        categories.value.unshift({ id: '', text: 'Tanpa kategori' });
        form.value.goods_category_id = String(category.goods_category_id);
        await Swal.fire({
            icon: 'success',
            title: 'Kategori berhasil ditambahkan',
            confirmButtonColor: '#059669',
        });
    } catch (error) {
        console.error('Gagal menambahkan kategori barang:', error);
        await Swal.fire({
            icon: 'error',
            title: 'Gagal menambahkan kategori',
            text:
                error.response?.data?.errors?.goods_category_name ||
                error.response?.data?.message ||
                'Terjadi kesalahan saat menyimpan kategori.',
        });
    }
}

async function loadPageData() {
    loading.value = true;
    pageError.value = '';

    try {
        const [categoryResponse, coaGoodsResponse, options, detailResponse] = await Promise.all([
            api.get('/goods-category-options'),
            api.get('/coa-goods'),
            fetchCoaOptions(),
            isEditing.value ? api.get(`/goods/${goodsId.value}`) : Promise.resolve(null),
        ]);

        categories.value = [
            { id: '', text: 'Tanpa kategori' },
            ...categoryResponse.data.data.map((category) => ({
                id: String(category.goods_category_id),
                text: category.goods_category_name,
            })),
        ];
        coaGoods.value = coaGoodsResponse.data.data;
        coaOptions.value = options;

        if (detailResponse) {
            const detail = detailResponse.data.data;
            form.value = {
                goods_name: detail.goods_name || '',
                goods_desc: detail.goods_desc || '',
                goods_active: Boolean(detail.goods_active),
                goods_category_id:
                    detail.goods_category_id == null ? '' : String(detail.goods_category_id),
                goods_brand_id: detail.goods_brand_id == null ? '' : String(detail.goods_brand_id),
                goods_unit_id: detail.goods_unit_id == null ? '' : String(detail.goods_unit_id),
                coas: [],
            };
            fillMappings(detail.coas);
        } else {
            fillMappings(coaGoods.value);
        }
    } catch (error) {
        console.error('Gagal memuat data barang:', error);
        pageError.value = error.response?.data?.message || 'Gagal memuat data barang.';
    } finally {
        loading.value = false;
    }
}

async function submitForm() {
    if (submitting.value || loading.value || coaGoods.value.length === 0) return;
    if (!validateForm()) return;

    submitting.value = true;
    pageError.value = '';

    const optionalId = (value) => (value === '' ? null : Number(value));
    const payload = {
        goods_name: form.value.goods_name.trim(),
        goods_desc: form.value.goods_desc.trim(),
        goods_active: form.value.goods_active,
        goods_category_id: optionalId(form.value.goods_category_id),
        goods_brand_id: optionalId(form.value.goods_brand_id),
        goods_unit_id: optionalId(form.value.goods_unit_id),
        coas: form.value.coas.map((item) => ({
            coa_goods_id: item.coa_goods_id,
            coa_id: item.coa_id === '' ? null : Number(item.coa_id),
        })),
    };

    try {
        if (isEditing.value) {
            await api.put(`/goods/${goodsId.value}`, payload);
        } else {
            await api.post('/goods', payload);
        }

        await Swal.fire({
            icon: 'success',
            title: isEditing.value ? 'Perubahan tersimpan' : 'Barang berhasil ditambahkan',
            text: 'Data barang berhasil disimpan.',
            confirmButtonColor: '#059669',
        });
        router.push('/admin/goods');
    } catch (error) {
        console.error('Gagal menyimpan barang:', error);
        const data = error.response?.data;
        if (error.response?.status === 422 && data?.errors) {
            const nextFieldErrors = {};
            const generalErrors = [];
            const supportedFields = [
                'goods_name',
                'goods_category_id',
                'goods_brand_id',
                'goods_unit_id',
            ];

            Object.entries(data.errors).forEach(([field, message]) => {
                if (
                    supportedFields.includes(field) ||
                    form.value.coas.some((item) => `coas.${item.coa_goods_id}` === field)
                ) {
                    nextFieldErrors[field] = message;
                } else {
                    generalErrors.push(message);
                }
            });

            fieldErrors.value = nextFieldErrors;
            pageError.value = generalErrors.join(' ');
        } else {
            pageError.value = data?.message || 'Data barang gagal disimpan.';
        }
    } finally {
        submitting.value = false;
    }
}

onMounted(loadPageData);
</script>

<template>
    <div class="min-h-[calc(100vh-4rem)] bg-slate-100 p-4 md:p-6">
        <div class="mx-auto max-w-7xl space-y-5">
            <header class="flex flex-wrap items-end justify-between gap-3">
                <div>
                    <p class="text-xs font-semibold uppercase tracking-wide text-emerald-700">
                        Inventaris
                    </p>
                    <h1 class="mt-1 text-2xl font-bold text-slate-900">
                        {{ isEditing ? 'Ubah Barang' : 'Tambah Barang' }}
                    </h1>
                </div>
                <RouterLink
                    to="/admin/goods"
                    class="inline-flex h-10 items-center gap-2 rounded-md border border-slate-300 bg-white px-4 text-sm font-semibold text-slate-700 transition hover:bg-slate-50"
                >
                    <i class="fa-solid fa-arrow-left" aria-hidden="true"></i>
                    Kembali
                </RouterLink>
            </header>

            <p
                v-if="pageError"
                role="alert"
                class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
            >
                {{ pageError }}
            </p>

            <form @submit.prevent="submitForm" novalidate class="space-y-5">
                <div
                    v-if="loading"
                    class="rounded-md border border-slate-200 bg-white p-10 text-center text-sm text-slate-500"
                >
                    Memuat data barang...
                </div>

                <div
                    v-else
                    class="grid items-start gap-5 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]"
                >
                    <section class="overflow-hidden rounded-md border border-slate-200 bg-white">
                        <div class="border-b border-slate-200 px-5 py-4">
                            <h2 class="font-semibold text-slate-900">Informasi Barang</h2>
                            <p class="mt-1 text-sm text-slate-500">
                                Detail dasar barang, kategori, dan status pencatatan.
                            </p>
                        </div>

                        <div class="grid gap-4 p-5 sm:grid-cols-2">
                            <label class="space-y-1.5 sm:col-span-2">
                                <span class="text-sm font-medium text-slate-700">Nama barang</span>
                                <input
                                    v-model="form.goods_name"
                                    type="text"
                                    maxlength="256"
                                    required
                                    :aria-invalid="Boolean(fieldErrors.goods_name)"
                                    class="h-10 w-full rounded-md border border-slate-300 px-3 text-sm text-slate-800 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/15"
                                    :class="
                                        fieldErrors.goods_name
                                            ? 'border-red-500 focus:border-red-600 focus:ring-red-600/15'
                                            : ''
                                    "
                                    placeholder="Masukkan nama barang"
                                    @input="clearFieldError('goods_name')"
                                />
                                <small
                                    v-if="fieldErrors.goods_name"
                                    class="block text-xs text-red-600"
                                >
                                    {{ fieldErrors.goods_name }}
                                </small>
                            </label>

                            <div class="space-y-1.5 sm:col-span-2">
                                <div class="flex items-center justify-between gap-3">
                                    <label
                                        for="goods-category"
                                        class="text-sm font-medium text-slate-700"
                                    >
                                        Kategori
                                    </label>
                                    <button
                                        type="button"
                                        class="text-xs font-semibold text-emerald-700 transition hover:text-emerald-900 disabled:opacity-50"
                                        :disabled="submitting"
                                        @click="addCategory"
                                    >
                                        <i class="fa-solid fa-plus mr-1" aria-hidden="true"></i>
                                        Tambah kategori
                                    </button>
                                </div>
                                <Select2Field
                                    id="goods-category"
                                    v-model="form.goods_category_id"
                                    :options="categories"
                                    placeholder="Cari atau pilih kategori..."
                                    :settings="{ allowClear: true }"
                                    :disabled="submitting"
                                    :class="
                                        fieldErrors.goods_category_id
                                            ? 'goods-select-error'
                                            : ''
                                    "
                                    @update:model-value="clearFieldError('goods_category_id')"
                                />
                                <small
                                    v-if="fieldErrors.goods_category_id"
                                    class="block text-xs text-red-600"
                                >
                                    {{ fieldErrors.goods_category_id }}
                                </small>
                            </div>

                            <label class="space-y-1.5">
                                <span class="text-sm font-medium text-slate-700">ID brand</span>
                                <input
                                    v-model="form.goods_brand_id"
                                    type="number"
                                    min="1"
                                    step="1"
                                    :aria-invalid="Boolean(fieldErrors.goods_brand_id)"
                                    class="h-10 w-full rounded-md border border-slate-300 px-3 text-sm text-slate-800 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/15"
                                    :class="
                                        fieldErrors.goods_brand_id
                                            ? 'border-red-500 focus:border-red-600 focus:ring-red-600/15'
                                            : ''
                                    "
                                    placeholder="Opsional"
                                    @input="clearFieldError('goods_brand_id')"
                                />
                                <small
                                    v-if="fieldErrors.goods_brand_id"
                                    class="block text-xs text-red-600"
                                >
                                    {{ fieldErrors.goods_brand_id }}
                                </small>
                            </label>

                            <label class="space-y-1.5">
                                <span class="text-sm font-medium text-slate-700">ID unit</span>
                                <input
                                    v-model="form.goods_unit_id"
                                    type="number"
                                    min="1"
                                    step="1"
                                    :aria-invalid="Boolean(fieldErrors.goods_unit_id)"
                                    class="h-10 w-full rounded-md border border-slate-300 px-3 text-sm text-slate-800 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/15"
                                    :class="
                                        fieldErrors.goods_unit_id
                                            ? 'border-red-500 focus:border-red-600 focus:ring-red-600/15'
                                            : ''
                                    "
                                    placeholder="Opsional"
                                    @input="clearFieldError('goods_unit_id')"
                                />
                                <small
                                    v-if="fieldErrors.goods_unit_id"
                                    class="block text-xs text-red-600"
                                >
                                    {{ fieldErrors.goods_unit_id }}
                                </small>
                            </label>

                            <label class="space-y-1.5 sm:col-span-2">
                                <span class="text-sm font-medium text-slate-700">Deskripsi</span>
                                <textarea
                                    v-model="form.goods_desc"
                                    rows="4"
                                    class="w-full resize-y rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-800 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/15"
                                    placeholder="Catatan atau keterangan barang"
                                ></textarea>
                            </label>

                            <div
                                class="flex items-center justify-between rounded-md border border-slate-200 bg-slate-50 px-3 py-3 sm:col-span-2"
                            >
                                <div>
                                    <p class="text-sm font-medium text-slate-800">Status barang</p>
                                    <p class="mt-0.5 text-xs text-slate-500">
                                        {{ form.goods_active ? 'Barang aktif' : 'Barang nonaktif' }}
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    role="switch"
                                    :aria-checked="form.goods_active"
                                    :aria-label="
                                        form.goods_active ? 'Nonaktifkan barang' : 'Aktifkan barang'
                                    "
                                    :disabled="submitting"
                                    class="relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition disabled:cursor-not-allowed disabled:opacity-60"
                                    :class="
                                        form.goods_active ? 'bg-emerald-600' : 'bg-slate-300'
                                    "
                                    @click="form.goods_active = !form.goods_active"
                                >
                                    <span
                                        class="inline-block size-5 rounded-full bg-white shadow transition-transform"
                                        :class="
                                            form.goods_active ? 'translate-x-5' : 'translate-x-0.5'
                                        "
                                    ></span>
                                </button>
                            </div>
                        </div>
                    </section>

                    <section class="overflow-hidden rounded-md border border-slate-200 bg-white">
                        <div
                            class="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200 px-5 py-4"
                        >
                            <div>
                                <h2 class="font-semibold text-slate-900">Pemetaan Akun COA</h2>
                                <p class="mt-1 text-sm text-slate-500">
                                    Pilih akun detail untuk setiap jenis barang.
                                </p>
                            </div>
                            <span
                                class="rounded bg-slate-100 px-2 py-1 text-xs font-medium text-slate-600"
                            >
                                {{ form.coas.length }} jenis akun
                            </span>
                        </div>

                        <div
                            v-if="form.coas.length === 0"
                            class="p-8 text-center text-sm text-slate-500"
                        >
                            Data jenis akun barang belum tersedia.
                        </div>
                        <div v-else class="divide-y divide-slate-100">
                            <div
                                v-for="item in form.coas"
                                :key="item.coa_goods_id"
                                class="grid gap-2 px-5 py-4 md:grid-cols-[minmax(125px,0.7fr)_minmax(0,1.3fr)] md:items-center md:gap-5"
                            >
                                <label
                                    :for="`goods-coa-${item.coa_goods_id}`"
                                    class="text-sm font-medium text-slate-700"
                                >
                                    {{ item.coa_goods_name }}
                                </label>
                                <Select2Field
                                    :id="`goods-coa-${item.coa_goods_id}`"
                                    v-model="item.coa_id"
                                    :options="coaOptions"
                                    placeholder="Cari atau pilih akun COA..."
                                    :settings="{ allowClear: true }"
                                    :disabled="submitting"
                                    :class="
                                        fieldErrors['coas.' + item.coa_goods_id]
                                            ? 'goods-select-error'
                                            : ''
                                    "
                                    @update:model-value="
                                        clearFieldError('coas.' + item.coa_goods_id)
                                    "
                                />
                                <small
                                    v-if="fieldErrors['coas.' + item.coa_goods_id]"
                                    class="block text-xs text-red-600 md:col-start-2"
                                >
                                    {{ fieldErrors['coas.' + item.coa_goods_id] }}
                                </small>
                            </div>
                        </div>
                    </section>
                </div>

                <footer
                    v-if="!loading"
                    class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end"
                >
                    <RouterLink
                        to="/admin/goods"
                        class="inline-flex h-10 items-center justify-center rounded-md border border-slate-300 bg-white px-5 text-sm font-semibold text-slate-700 transition hover:bg-slate-50"
                    >
                        Batal
                    </RouterLink>
                    <button
                        type="submit"
                        :disabled="submitting || coaGoods.length === 0"
                        class="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-emerald-700 px-5 text-sm font-semibold text-white transition hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-600 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                        <i
                            v-if="!submitting"
                            class="fa-solid fa-floppy-disk"
                            aria-hidden="true"
                        ></i>
                        {{
                            submitting
                                ? 'Menyimpan...'
                                : isEditing
                                  ? 'Simpan perubahan'
                                  : 'Simpan barang'
                        }}
                    </button>
                </footer>
            </form>
        </div>
    </div>
</template>

<style scoped>
.goods-select-error :deep(.select2-selection--single) {
    border-color: #dc2626;
}
</style>
