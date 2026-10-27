```vue
<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import Swal from 'sweetalert2';
import api from '@/api';
import Select2Field from '@/components/Select2Field.vue';

const route = useRoute();
const router = useRouter();
const partaiId = computed(() => route.params.id);
const isEditing = computed(() => Boolean(partaiId.value));
const bungkusOptions = ['Tikar', 'Kain', 'Glansing', 'Karung Goni', 'Karung Plastik', 'Yute Rs'];
const bungkusSelectOptions = [
    { id: '', text: '' },
    ...bungkusOptions.map((option) => ({ 
        id: option, 
        text: option 
    })),
];
const currentYear = new Date().getFullYear();
const form = ref(emptyForm());
const coaGoods = ref([]);
const coaOptions = ref([]);
const loading = ref(false);
const submitting = ref(false);
const fieldErrors = ref({});
const pageError = ref('');

function emptyForm() {
    return {
        goods_partai_code: '',
        goods_partai_name: '',
        goods_partai_bungkus: '',
        goods_partai_year: currentYear,
        goods_partai_desc: '',
        goods_partai_active: true,
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
    const code = Number(form.value.goods_partai_code);
    const year = Number(form.value.goods_partai_year);
    const name = form.value.goods_partai_name.trim();

    if (!Number.isInteger(code) || code <= 0) {
        errors.goods_partai_code = 'Masukkan kode partai yang valid.';
    }
    if (!Number.isInteger(year) || year < 1990 || year > 2100) {
        errors.goods_partai_year = 'Tahun harus berada antara 1990 dan 2100.';
    }
    if (!name) {
        errors.goods_partai_name = 'Nama partai wajib diisi.';
    } else if (name.length > 256) {
        errors.goods_partai_name = 'Nama partai maksimal 256 karakter.';
    }
    if (!bungkusOptions.includes(form.value.goods_partai_bungkus)) {
        errors.goods_partai_bungkus = 'Pilih jenis bungkus.';
    }

    fieldErrors.value = errors;
    return Object.keys(errors).length === 0;
}

async function fetchCoaOptions() {
    const response = await api.get('/coa');
    return [
        { id: '', text: '' },
        ...response.data.data
            .filter((coa) =>
                String(coa.coa_no || '')
                    .trim()
                    .endsWith(',00'),
            )
            .map((coa) => ({
                id: String(coa.coa_id),
                text: `${coa.coa_no} - ${coa.coa_name}`,
            })),
    ];
}

async function loadPageData() {
    loading.value = true;

    try {
        const [goodsResponse, options, detailResponse] = await Promise.all([
            api.get('/coa-goods'),
            fetchCoaOptions(),
            isEditing.value ? api.get(`/goods-partai/${partaiId.value}`) : Promise.resolve(null),
        ]);

        coaGoods.value = goodsResponse.data.data;
        coaOptions.value = options;

        if (detailResponse) {
            form.value = {
                ...detailResponse.data.data,
                goods_partai_code: detailResponse.data.data.goods_partai_code,
                goods_partai_year: detailResponse.data.data.goods_partai_year,
                goods_partai_active: Boolean(detailResponse.data.data.goods_partai_active),
                coas: [],
            };
            fillMappings(detailResponse.data.data.coas);
        } else {
            fillMappings(coaGoods.value);
        }
    } catch (error) {
        console.error('Gagal mengambil data partai barang:', error);
        pageError.value = error.response?.data?.message || 'Gagal memuat data partai barang.';
    } finally {
        loading.value = false;
    }
}

async function submitForm() {
    if (submitting.value || loading.value || coaGoods.value.length === 0) return;
    if (!validateForm()) return;

    submitting.value = true;
    pageError.value = '';

    try {
        const payload = {
            goods_partai_code: Number(form.value.goods_partai_code),
            goods_partai_name: form.value.goods_partai_name.trim(),
            goods_partai_bungkus: form.value.goods_partai_bungkus,
            goods_partai_year: Number(form.value.goods_partai_year),
            goods_partai_desc: form.value.goods_partai_desc.trim(),
            goods_partai_active: form.value.goods_partai_active,
            coas: form.value.coas.map((item) => ({
                coa_goods_id: item.coa_goods_id,
                coa_id: item.coa_id === '' ? null : Number(item.coa_id),
            })),
        };

        if (isEditing.value) {
            await api.put(`/goods-partai/${partaiId.value}`, payload);
        } else {
            await api.post('/goods-partai', payload);
        }

        await Swal.fire({
            icon: 'success',
            title: isEditing.value ? 'Perubahan tersimpan' : 'Partai berhasil ditambahkan',
            text: 'Data partai barang berhasil disimpan.',
            confirmButtonColor: '#059669',
        });
        router.push('/admin/goods-partai');
    } catch (error) {
        console.error('Gagal menyimpan partai barang:', error);
        const data = error.response?.data;
        if (error.response?.status === 422 && data?.errors) {
            const nextFieldErrors = {};
            const generalErrors = [];

            Object.entries(data.errors).forEach(([field, message]) => {
                const coaGoodsId = field.match(/^coas\.(\d+)$/)?.[1];
                const row = form.value.coas.find(
                    (item) => String(item.coa_goods_id) === coaGoodsId,
                );

                if (row) {
                    nextFieldErrors[field] = message;
                } else if (
                    [
                        'goods_partai_code',
                        'goods_partai_year',
                        'goods_partai_name',
                        'goods_partai_bungkus',
                    ].includes(field)
                ) {
                    nextFieldErrors[field] = message;
                } else {
                    generalErrors.push(message);
                }
            });

            fieldErrors.value = nextFieldErrors;
            pageError.value = generalErrors.join(' ');
        } else {
            pageError.value = data?.message || 'Data partai barang gagal disimpan.';
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
                        {{ isEditing ? 'Ubah Partai Barang' : 'Tambah Partai Barang' }}
                    </h1>
                </div>
                <RouterLink
                    to="/admin/goods-partai"
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
                    Memuat data partai...
                </div>

                <div
                    v-else
                    class="grid items-start gap-5 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]"
                >
                    <section class="overflow-hidden rounded-md border border-slate-200 bg-white">
                        <div class="border-b border-slate-200 px-5 py-4">
                            <h2 class="font-semibold text-slate-900">Informasi Partai</h2>
                            <p class="mt-1 text-sm text-slate-500">
                                Detail dasar barang dan status pencatatan.
                            </p>
                        </div>

                        <div class="grid gap-4 p-5 sm:grid-cols-2">
                            <label class="space-y-1.5">
                                <span class="text-sm font-medium text-slate-700">Kode partai</span>
                                <input
                                    v-model="form.goods_partai_code"
                                    type="number"
                                    min="1"
                                    required
                                    :aria-invalid="Boolean(fieldErrors.goods_partai_code)"
                                    class="h-10 w-full rounded-md border border-slate-300 px-3 text-sm text-slate-800 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/15"
                                    :class="
                                        fieldErrors.goods_partai_code
                                            ? 'border-red-500 focus:border-red-600 focus:ring-red-600/15'
                                            : ''
                                    "
                                    placeholder="Contoh: 101"
                                    @input="clearFieldError('goods_partai_code')"
                                />
                                <small
                                    v-if="fieldErrors.goods_partai_code"
                                    class="block text-xs text-red-600"
                                >
                                    {{ fieldErrors.goods_partai_code }}
                                </small>
                            </label>

                            <label class="space-y-1.5">
                                <span class="text-sm font-medium text-slate-700">Tahun</span>
                                <input
                                    v-model="form.goods_partai_year"
                                    type="number"
                                    min="1990"
                                    max="2100"
                                    required
                                    :aria-invalid="Boolean(fieldErrors.goods_partai_year)"
                                    class="h-10 w-full rounded-md border border-slate-300 px-3 text-sm text-slate-800 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/15"
                                    :class="
                                        fieldErrors.goods_partai_year
                                            ? 'border-red-500 focus:border-red-600 focus:ring-red-600/15'
                                            : ''
                                    "
                                    @input="clearFieldError('goods_partai_year')"
                                />
                                <small
                                    v-if="fieldErrors.goods_partai_year"
                                    class="block text-xs text-red-600"
                                >
                                    {{ fieldErrors.goods_partai_year }}
                                </small>
                            </label>

                            <label class="space-y-1.5 sm:col-span-2">
                                <span class="text-sm font-medium text-slate-700">Nama partai</span>
                                <input
                                    v-model="form.goods_partai_name"
                                    type="text"
                                    maxlength="256"
                                    required
                                    :aria-invalid="Boolean(fieldErrors.goods_partai_name)"
                                    class="h-10 w-full rounded-md border border-slate-300 px-3 text-sm text-slate-800 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/15"
                                    :class="
                                        fieldErrors.goods_partai_name
                                            ? 'border-red-500 focus:border-red-600 focus:ring-red-600/15'
                                            : ''
                                    "
                                    placeholder="Masukkan nama partai barang"
                                    @input="clearFieldError('goods_partai_name')"
                                />
                                <small
                                    v-if="fieldErrors.goods_partai_name"
                                    class="block text-xs text-red-600"
                                >
                                    {{ fieldErrors.goods_partai_name }}
                                </small>
                            </label>

                            <label class="space-y-1.5 sm:col-span-2">
                                <span class="text-sm font-medium text-slate-700">Bungkus</span>
                                <Select2Field
                                    v-model="form.goods_partai_bungkus"
                                    :options="bungkusSelectOptions"
                                    placeholder="Pilih jenis bungkus..."
                                    required
                                    :aria-invalid="Boolean(fieldErrors.goods_partai_bungkus)"
                                    :class="
                                        fieldErrors.goods_partai_bungkus
                                            ? 'goods-partai-select-error'
                                            : ''
                                    "
                                    @update:model-value="clearFieldError('goods_partai_bungkus')"
                                />
                                <small
                                    v-if="fieldErrors.goods_partai_bungkus"
                                    class="block text-xs text-red-600"
                                >
                                    {{ fieldErrors.goods_partai_bungkus }}
                                </small>
                            </label>

                            <label class="space-y-1.5 sm:col-span-2">
                                <span class="text-sm font-medium text-slate-700">Deskripsi</span>
                                <textarea
                                    v-model="form.goods_partai_desc"
                                    rows="4"
                                    class="w-full resize-y rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-800 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/15"
                                    placeholder="Catatan atau keterangan partai"
                                ></textarea>
                            </label>

                            <div
                                class="flex items-center justify-between rounded-md border border-slate-200 bg-slate-50 px-3 py-3 sm:col-span-2"
                            >
                                <div>
                                    <p class="text-sm font-medium text-slate-800">Status partai</p>
                                    <p class="mt-0.5 text-xs text-slate-500">
                                        {{
                                            form.goods_partai_active
                                                ? 'Partai aktif'
                                                : 'Partai nonaktif'
                                        }}
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    role="switch"
                                    :aria-checked="form.goods_partai_active"
                                    :aria-label="
                                        form.goods_partai_active
                                            ? 'Nonaktifkan partai'
                                            : 'Aktifkan partai'
                                    "
                                    :disabled="submitting"
                                    class="relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition disabled:cursor-not-allowed disabled:opacity-60"
                                    :class="
                                        form.goods_partai_active ? 'bg-emerald-600' : 'bg-slate-300'
                                    "
                                    @click="form.goods_partai_active = !form.goods_partai_active"
                                >
                                    <span
                                        class="inline-block size-5 rounded-full bg-white shadow transition-transform"
                                        :class="
                                            form.goods_partai_active
                                                ? 'translate-x-5'
                                                : 'translate-x-0.5'
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
                                    :for="`coa-goods-${item.coa_goods_id}`"
                                    class="text-sm font-medium text-slate-700"
                                >
                                    {{ item.coa_goods_name }}
                                </label>
                                <Select2Field
                                    :id="`coa-goods-${item.coa_goods_id}`"
                                    v-model="item.coa_id"
                                    :options="coaOptions"
                                    placeholder="Cari atau pilih akun COA..."
                                    :settings="{ allowClear: true }"
                                    :disabled="submitting"
                                    :class="
                                        fieldErrors['coas.' + item.coa_goods_id]
                                            ? 'goods-partai-select-error'
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
                        to="/admin/goods-partai"
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
                                  : 'Simpan partai'
                        }}
                    </button>
                </footer>
            </form>
        </div>
    </div>
</template>

<style scoped>
.goods-partai-select-error :deep(.select2-selection--single) {
    border-color: #dc2626;
}
</style>
