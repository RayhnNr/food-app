<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import api from '@/api';
import Swal from 'sweetalert2';
import Modal from '@/components/Modal.vue';
import Pagination from '@/components/Pagination.vue';
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
const currentPage = ref(1);
const limit = ref(10);
const totalData = ref(0);
const totalPages = ref(0);
const showForm = ref(false);
const editingId = ref(null);
const saving = ref(false);
const errors = ref({});
const formError = ref('');
const provinces = ref([]);
const cities = ref([]);
const subdistricts = ref([]);

const statusOptions = [
    { id: 'all', text: 'Semua status' },
    { id: 'true', text: 'Aktif' },
    { id: 'false', text: 'Tidak aktif' },
];
const regionSettings = { minimumResultsForSearch: 8 };
const isEdit = computed(() => editingId.value !== null);
const actionButtonClass =
    'inline-flex h-9 w-9 items-center justify-center rounded-md border text-gray-600 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500';

const today = () => {
    const date = new Date();
    date.setMinutes(date.getMinutes() - date.getTimezoneOffset());
    return date.toISOString().slice(0, 10);
};

const emptyForm = () => ({
    bandul_name: '',
    bandul_date_join: today(),
    bandul_date_birth: '',
    bandul_desc: '',
    bandul_phone: '',
    bandul_mobile_phone: '',
    province_id: '',
    city_id: '',
    subdistrict_id: '',
    bandul_address: '',
    bandul_active: true,
});
const form = reactive(emptyForm());
const toOptions = (items) => items.map((item) => ({ id: String(item.id), text: item.name }));

const load = async () => {
    loading.value = true;
    loadError.value = '';

    try {
        const params = { page: currentPage.value, limit: limit.value };
        if (search.value.trim()) params.q = search.value.trim();
        if (activeFilter.value !== 'all') params.active = activeFilter.value;

        const res = await api.get('/bandul', { params });
        rows.value = res.data.data;
        currentPage.value = res.data.currentPage;
        totalData.value = res.data.totalData;
        totalPages.value = res.data.totalPages;
    } catch (e) {
        loadError.value = e.response?.data?.message || 'Gagal mengambil data bandul.';
        console.error(e);
    } finally {
        loading.value = false;
    }
};

const loadProvinces = async () => {
    if (provinces.value.length) return;
    const res = await api.get('/provinces');
    provinces.value = toOptions(res.data.data);
};

const loadCities = async (provinceId) => {
    cities.value = [];
    if (!provinceId) return;
    const res = await api.get(`/provinces/${provinceId}/cities`);
    cities.value = toOptions(res.data.data);
};

const loadSubdistricts = async (cityId) => {
    subdistricts.value = [];
    if (!cityId) return;
    const res = await api.get(`/cities/${cityId}/subdistricts`);
    subdistricts.value = toOptions(res.data.data);
};

const resetErrors = () => {
    errors.value = {};
    formError.value = '';
};

const openCreate = async () => {
    editingId.value = null;
    Object.assign(form, emptyForm());
    cities.value = [];
    subdistricts.value = [];
    resetErrors();
    showForm.value = true;
    try {
        await loadProvinces();
    } catch (e) {
        formError.value = e.response?.data?.message || 'Gagal mengambil data provinsi.';
    }
};

const openEdit = async (row) => {
    resetErrors();
    showForm.value = true;
    saving.value = true;
    try {
        const [detail] = await Promise.all([api.get(`/bandul/${row.bandul_id}`), loadProvinces()]);
        const item = detail.data.data;
        editingId.value = item.bandul_id;
        Object.assign(form, {
            bandul_name: item.bandul_name,
            bandul_date_join: item.bandul_date_join,
            bandul_date_birth: item.bandul_date_birth || '',
            bandul_desc: item.bandul_desc || '',
            bandul_phone: item.bandul_phone || '',
            bandul_mobile_phone: item.bandul_mobile_phone || '',
            province_id: item.province_id ? String(item.province_id) : '',
            city_id: item.city_id ? String(item.city_id) : '',
            subdistrict_id: item.subdistrict_id ? String(item.subdistrict_id) : '',
            bandul_address: item.bandul_address || '',
            bandul_active: item.bandul_active,
        });
        await loadCities(form.province_id);
        await loadSubdistricts(form.city_id);
    } catch (e) {
        showForm.value = false;
        Swal.fire({
            icon: 'error',
            title: 'Gagal membuka data',
            text: e.response?.data?.message || 'Data bandul gagal dimuat.',
        });
    } finally {
        saving.value = false;
    }
};

const closeForm = () => {
    if (saving.value) return;
    showForm.value = false;
};

const changeProvince = async (value) => {
    form.province_id = value || '';
    form.city_id = '';
    form.subdistrict_id = '';
    subdistricts.value = [];
    try {
        await loadCities(form.province_id);
    } catch (e) {
        errors.value.province_id = e.response?.data?.message || 'Gagal mengambil data kota.';
    }
};

const changeCity = async (value) => {
    form.city_id = value || '';
    form.subdistrict_id = '';
    try {
        await loadSubdistricts(form.city_id);
    } catch (e) {
        errors.value.city_id = e.response?.data?.message || 'Gagal mengambil data kecamatan.';
    }
};

const save = async () => {
    if (saving.value) return;
    resetErrors();
    saving.value = true;

    const payload = {
        ...form,
        province_id: form.province_id ? Number(form.province_id) : null,
        city_id: form.city_id ? Number(form.city_id) : null,
        subdistrict_id: form.subdistrict_id ? Number(form.subdistrict_id) : null,
    };

    try {
        if (isEdit.value) {
            await api.put(`/bandul/${editingId.value}`, payload);
        } else {
            await api.post('/bandul', payload);
        }
        showForm.value = false;
        await load();
        Toast.fire({
            icon: 'success',
            title: isEdit.value ? 'Berhasil mengubah bandul' : 'Berhasil menambahkan bandul',
        });
    } catch (e) {
        const data = e.response?.data;
        errors.value = data?.errors || {};
        formError.value = data?.errors ? '' : data?.message || 'Gagal menyimpan bandul.';
    } finally {
        saving.value = false;
    }
};

const remove = async (row) => {
    const result = await Swal.fire({
        title: 'Yakin hapus bandul ini?',
        text: row.bandul_name,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Ya, hapus',
        cancelButtonText: 'Batal',
    });
    if (!result.isConfirmed) return;

    try {
        await api.delete(`/bandul/${row.bandul_id}`);
        if (currentPage.value > 1 && rows.value.length === 1) currentPage.value -= 1;
        await load();
        Toast.fire({ icon: 'success', title: 'Berhasil menghapus bandul' });
    } catch (e) {
        Swal.fire({
            icon: 'error',
            title: 'Gagal menghapus',
            text: e.response?.data?.message || 'Gagal menghapus bandul.',
        });
    }
};

const changePage = (page) => {
    if (page < 1 || page > totalPages.value) return;
    currentPage.value = page;
    load();
};

let searchTimer;
watch(search, () => {
    clearTimeout(searchTimer);
    currentPage.value = 1;
    searchTimer = setTimeout(load, 300);
});
watch(activeFilter, () => {
    currentPage.value = 1;
    load();
});
onMounted(load);
onBeforeUnmount(() => clearTimeout(searchTimer));
</script>

<template>
    <section class="p-6">
        <div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div>
                <p class="mb-1 text-sm font-medium uppercase text-emerald-700">Data master</p>
                <h1 class="text-3xl font-bold text-slate-800">Data Bandul</h1>
            </div>
            <button
                type="button"
                @click="openCreate"
                class="inline-flex items-center justify-center gap-2 rounded-md bg-emerald-600 px-4 py-2.5 font-medium text-white transition hover:bg-emerald-700"
            >
                <i class="fa-solid fa-plus"></i>
                Tambah Bandul
            </button>
        </div>

        <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center">
            <input
                v-model="search"
                type="search"
                placeholder="Cari nama atau nomor telepon..."
                aria-label="Cari bandul"
                class="w-full rounded-md border border-slate-300 bg-white px-4 py-2.5 text-sm focus:border-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-100 sm:max-w-md"
            />
            <Select2Field
                v-model="activeFilter"
                :options="statusOptions"
                :settings="regionSettings"
                aria-label="Filter status bandul"
                class="sm:max-w-52"
            />
        </div>

        <p
            v-if="loadError"
            role="alert"
            class="mb-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
        >
            {{ loadError }}
        </p>

        <div class="overflow-x-auto rounded-lg bg-white shadow-sm ring-1 ring-slate-200">
            <table class="w-full min-w-[900px] table-auto">
                <thead>
                    <tr class="bg-slate-100 text-xs uppercase text-slate-600">
                        <th class="px-5 py-3 text-left">#</th>
                        <th class="px-5 py-3 text-left">Nama Bandul</th>
                        <th class="px-5 py-3 text-left">Kontak</th>
                        <th class="px-5 py-3 text-left">Wilayah</th>
                        <th class="px-5 py-3 text-left">Bergabung</th>
                        <th class="px-5 py-3 text-left">Status</th>
                        <th class="px-5 py-3 text-left">Aksi</th>
                    </tr>
                </thead>
                <tbody class="text-sm text-slate-700">
                    <tr v-if="loading && rows.length === 0">
                        <td colspan="7" class="px-5 py-10 text-center text-slate-500">
                            Memuat data...
                        </td>
                    </tr>
                    <tr v-else-if="rows.length === 0">
                        <td colspan="7" class="px-5 py-10 text-center text-slate-500">
                            Belum ada data bandul.
                        </td>
                    </tr>
                    <tr
                        v-for="(row, index) in rows"
                        :key="row.bandul_id"
                        class="border-t border-slate-100 hover:bg-slate-50"
                    >
                        <td class="px-5 py-3">{{ (currentPage - 1) * limit + index + 1 }}</td>
                        <td class="px-5 py-3">
                            <p class="font-semibold text-slate-800">{{ row.bandul_name }}</p>
                            <p
                                v-if="row.bandul_desc"
                                class="mt-0.5 max-w-xs truncate text-xs text-slate-500"
                            >
                                {{ row.bandul_desc }}
                            </p>
                        </td>
                        <td class="px-5 py-3">
                            <p>{{ row.bandul_mobile_phone || '-' }}</p>
                            <p v-if="row.bandul_phone" class="text-xs text-slate-500">
                                Telp. {{ row.bandul_phone }}
                            </p>
                        </td>
                        <td class="px-5 py-3">
                            <p>
                                {{
                                    row.subdistrict_name ||
                                    row.city_name ||
                                    row.province_name ||
                                    '-'
                                }}
                            </p>
                            <p
                                v-if="row.province_name && row.city_name"
                                class="text-xs text-slate-500"
                            >
                                {{ row.city_name }}, {{ row.province_name }}
                            </p>
                        </td>
                        <td class="px-5 py-3">{{ row.bandul_date_join }}</td>
                        <td class="px-5 py-3">
                            <span
                                class="inline-flex rounded px-2 py-1 text-xs font-semibold"
                                :class="
                                    row.bandul_active
                                        ? 'bg-emerald-100 text-emerald-800'
                                        : 'bg-slate-100 text-slate-600'
                                "
                            >
                                {{ row.bandul_active ? 'Aktif' : 'Nonaktif' }}
                            </span>
                        </td>
                        <td class="px-5 py-3">
                            <div class="flex items-center gap-2">
                                <button
                                    type="button"
                                    @click="openEdit(row)"
                                    :class="[actionButtonClass, 'border-blue-200 bg-blue-50']"
                                    title="Ubah bandul"
                                    aria-label="Ubah bandul"
                                >
                                    <i class="fa-solid fa-pencil"></i>
                                </button>
                                <button
                                    type="button"
                                    @click="remove(row)"
                                    :class="[actionButtonClass, 'border-red-200 bg-red-50']"
                                    title="Hapus bandul"
                                    aria-label="Hapus bandul"
                                >
                                    <i class="fa-solid fa-trash-can"></i>
                                </button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
            <Pagination
                :current-page="currentPage"
                :total-pages="totalPages"
                :total-items="totalData"
                :page-size="limit"
                :loading="loading"
                @change="changePage"
            />
        </div>
    </section>

    <Modal
        :show="showForm"
        :title="isEdit ? 'Ubah Bandul' : 'Tambah Bandul'"
        max-width="max-w-4xl"
        @close="closeForm"
    >
        <form class="max-h-[75vh] space-y-5 overflow-y-auto pr-1" @submit.prevent="save">
            <fieldset :disabled="saving" class="space-y-5">
                <div class="grid gap-4 sm:grid-cols-2">
                    <div>
                        <label
                            for="bandul-name"
                            class="mb-1 block text-sm font-medium text-slate-700"
                            >Nama bandul *</label
                        >
                        <input
                            id="bandul-name"
                            v-model="form.bandul_name"
                            required
                            maxlength="256"
                            type="text"
                            class="w-full rounded-md border border-slate-300 px-3 py-2.5 focus:border-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-100"
                        />
                        <p v-if="errors.bandul_name" class="mt-1 text-sm text-red-600">
                            {{ errors.bandul_name }}
                        </p>
                    </div>
                    <div>
                        <label
                            for="bandul-date-join"
                            class="mb-1 block text-sm font-medium text-slate-700"
                            >Tanggal bergabung *</label
                        >
                        <input
                            id="bandul-date-join"
                            v-model="form.bandul_date_join"
                            required
                            type="date"
                            class="w-full rounded-md border border-slate-300 px-3 py-2.5 focus:border-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-100"
                        />
                        <p v-if="errors.bandul_date_join" class="mt-1 text-sm text-red-600">
                            {{ errors.bandul_date_join }}
                        </p>
                    </div>
                    <div>
                        <label
                            for="bandul-date-birth"
                            class="mb-1 block text-sm font-medium text-slate-700"
                            >Tanggal lahir</label
                        >
                        <input
                            id="bandul-date-birth"
                            v-model="form.bandul_date_birth"
                            type="date"
                            :max="today()"
                            class="w-full rounded-md border border-slate-300 px-3 py-2.5 focus:border-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-100"
                        />
                        <p v-if="errors.bandul_date_birth" class="mt-1 text-sm text-red-600">
                            {{ errors.bandul_date_birth }}
                        </p>
                    </div>
                    <div>
                        <label
                            for="bandul-mobile"
                            class="mb-1 block text-sm font-medium text-slate-700"
                            >Nomor HP</label
                        >
                        <input
                            id="bandul-mobile"
                            v-model="form.bandul_mobile_phone"
                            type="tel"
                            maxlength="50"
                            class="w-full rounded-md border border-slate-300 px-3 py-2.5 focus:border-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-100"
                        />
                        <p v-if="errors.bandul_mobile_phone" class="mt-1 text-sm text-red-600">
                            {{ errors.bandul_mobile_phone }}
                        </p>
                    </div>
                    <div>
                        <label
                            for="bandul-phone"
                            class="mb-1 block text-sm font-medium text-slate-700"
                            >Telepon</label
                        >
                        <input
                            id="bandul-phone"
                            v-model="form.bandul_phone"
                            type="tel"
                            maxlength="50"
                            class="w-full rounded-md border border-slate-300 px-3 py-2.5 focus:border-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-100"
                        />
                        <p v-if="errors.bandul_phone" class="mt-1 text-sm text-red-600">
                            {{ errors.bandul_phone }}
                        </p>
                    </div>
                    <div>
                        <label class="mb-1 block text-sm font-medium text-slate-700"
                            >Provinsi</label
                        >
                        <Select2Field
                            :model-value="form.province_id"
                            :options="provinces"
                            :settings="regionSettings"
                            placeholder="Pilih provinsi"
                            @update:model-value="changeProvince"
                        />
                        <p v-if="errors.province_id" class="mt-1 text-sm text-red-600">
                            {{ errors.province_id }}
                        </p>
                    </div>
                    <div>
                        <label class="mb-1 block text-sm font-medium text-slate-700"
                            >Kota / Kabupaten</label
                        >
                        <Select2Field
                            :model-value="form.city_id"
                            :options="cities"
                            :settings="regionSettings"
                            :disabled="!form.province_id"
                            placeholder="Pilih kota"
                            @update:model-value="changeCity"
                        />
                        <p v-if="errors.city_id" class="mt-1 text-sm text-red-600">
                            {{ errors.city_id }}
                        </p>
                    </div>
                    <div>
                        <label class="mb-1 block text-sm font-medium text-slate-700"
                            >Kecamatan</label
                        >
                        <Select2Field
                            v-model="form.subdistrict_id"
                            :options="subdistricts"
                            :settings="regionSettings"
                            :disabled="!form.city_id"
                            placeholder="Pilih kecamatan"
                        />
                        <p v-if="errors.subdistrict_id" class="mt-1 text-sm text-red-600">
                            {{ errors.subdistrict_id }}
                        </p>
                    </div>
                    <div>
                        <label class="mb-1 block text-sm font-medium text-slate-700">Status</label>
                        <label
                            class="flex h-[42px] items-center gap-2 rounded-md border border-slate-300 px-3 text-sm text-slate-700"
                        >
                            <input
                                v-model="form.bandul_active"
                                type="checkbox"
                                class="h-4 w-4 accent-emerald-600"
                            />
                            Aktif
                        </label>
                    </div>
                </div>
                <div>
                    <label
                        for="bandul-address"
                        class="mb-1 block text-sm font-medium text-slate-700"
                        >Alamat</label
                    >
                    <textarea
                        id="bandul-address"
                        v-model="form.bandul_address"
                        rows="2"
                        class="w-full rounded-md border border-slate-300 px-3 py-2.5 focus:border-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-100"
                    ></textarea>
                </div>
                <div>
                    <label for="bandul-desc" class="mb-1 block text-sm font-medium text-slate-700"
                        >Deskripsi</label
                    >
                    <textarea
                        id="bandul-desc"
                        v-model="form.bandul_desc"
                        rows="2"
                        class="w-full rounded-md border border-slate-300 px-3 py-2.5 focus:border-emerald-600 focus:outline-none focus:ring-2 focus:ring-emerald-100"
                    ></textarea>
                </div>
            </fieldset>
            <p v-if="formError" role="alert" class="text-sm text-red-600">{{ formError }}</p>
            <div class="flex justify-end gap-2 border-t border-slate-100 pt-4">
                <button
                    type="button"
                    @click="closeForm"
                    :disabled="saving"
                    class="rounded-md border border-slate-300 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50 disabled:opacity-50"
                >
                    Batal
                </button>
                <button
                    type="submit"
                    :disabled="saving"
                    class="rounded-md bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 disabled:opacity-50"
                >
                    {{ saving ? 'Menyimpan...' : 'Simpan' }}
                </button>
            </div>
        </form>
    </Modal>
</template>
