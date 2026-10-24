<script setup>
import { ref, computed, onMounted, watch } from 'vue';
import Multiselect from '@vueform/multiselect';
import '@vueform/multiselect/themes/default.css';
import api from '@/api';
import Swal from 'sweetalert2';
import Modal from '@/components/Modal.vue';

const coas = ref([]);
const coaTypes = ref([]);
const loading = ref(true);
const error = ref(null);
const search = ref('');
const showModal = ref(false);
const saving = ref(false);
const editingCoa = ref(null);
const kodeManual = ref(false); // true kalau user sudah mengetik kode sendiri
const formError = ref('');
const formErrors = ref({}); // error per field dari backend (422)

const emptyForm = () => ({
    coa_no: '',
    coa_name: '',
    coa_type_id: '',
    coa_parent: '',
    coa_head: false,
});
const form = ref(emptyForm());

const actionButtonClass =
    'inline-flex h-9 w-9 items-center justify-center rounded-md border text-gray-600 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500';

const Toast = Swal.mixin({
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
});

// ===== Data =====

// Backend tidak punya fitur search, jadi difilter di sini
const filteredCoas = computed(() => {
    const q = search.value.trim().toLowerCase();
    if (!q) return coas.value;
    return coas.value.filter(
        (c) =>
            c.coa_no.toLowerCase().includes(q) ||
            c.coa_name.toLowerCase().includes(q) ||
            (c.type?.coa_type_name || '').toLowerCase().includes(q),
    );
});

// Hanya akun header yang boleh jadi parent (akun yang sedang diedit tidak ikut)
const parentOptions = computed(() =>
    coas.value.filter((c) => c.coa_head && c.coa_id !== editingCoa.value?.coa_id),
);
const parentSelectOptions = computed(() =>
    parentOptions.value.map((coa) => ({
        value: coa.coa_id,
        label: `${coa.coa_no} - ${coa.coa_name}`,
    })),
);
const coaTypeSelectOptions = computed(() =>
    coaTypes.value.map((type) => ({
        value: type.coa_type_id,
        label: type.coa_type_name,
    })),
);

// Level akun (untuk indentasi nama)
const levelOf = (coa) => {
    let level = 0;
    let cur = coa;
    while (cur?.coa_parent && level < 10) {
        cur = coas.value.find((c) => c.coa_id === cur.coa_parent);
        level++;
    }
    return level;
};

const getCoa = async () => {
    loading.value = true;
    error.value = null;

    try {
        const res = await api.get('/coa'); // sudah urut coa_no dari backend
        coas.value = res.data.data;
    } catch (e) {
        error.value = 'Gagal mengambil data COA. Pastikan backend sudah jalan.';
        console.error(e);
    } finally {
        loading.value = false;
    }
};

const getCoaTypes = async () => {
    try {
        const res = await api.get('/coa-types');
        coaTypes.value = res.data.data;
    } catch (e) {
        console.error(e);
    }
};

// ===== Kode otomatis (saran, tetap bisa diedit) =====

const ambilKodeBaru = async () => {
    if (editingCoa.value || kodeManual.value) return;

    try {
        const params = form.value.coa_parent ? { parent_id: form.value.coa_parent } : {};
        const res = await api.get('/coa/next-code', { params });
        form.value.coa_no = res.data.coa_no;
    } catch (e) {
        // mis. kode parent tidak berformat 000.000.000 -> user isi manual
        form.value.coa_no = '';
        console.error(e);
    }
};

watch(
    () => form.value.coa_parent,
    (parentId) => {
        if (editingCoa.value) return;

        // tipe ikut parent kalau belum dipilih
        if (parentId && !form.value.coa_type_id) {
            const p = coas.value.find((c) => c.coa_id === parentId);
            if (p) form.value.coa_type_id = p.coa_type_id;
        }
        ambilKodeBaru();
    },
);

// ===== Modal =====

const bukaModal = (coa = null) => {
    editingCoa.value = coa;
    kodeManual.value = false;
    formError.value = '';
    formErrors.value = {};

    form.value = coa
        ? {
              coa_no: coa.coa_no,
              coa_name: coa.coa_name,
              coa_type_id: coa.coa_type_id,
              coa_parent: coa.coa_parent ?? '',
              coa_head: coa.coa_head,
          }
        : emptyForm();

    showModal.value = true;
    if (!coa) ambilKodeBaru();
};

const tutupModal = () => {
    if (saving.value) return;
    showModal.value = false;
};

const simpanCoa = async () => {
    if (saving.value) return;

    saving.value = true;
    formError.value = '';
    formErrors.value = {};
    const isEdit = Boolean(editingCoa.value);

    const payload = {
        coa_no: form.value.coa_no.trim(),
        coa_name: form.value.coa_name.trim(),
        coa_type_id: Number(form.value.coa_type_id) || 0,
        coa_parent: form.value.coa_parent ? Number(form.value.coa_parent) : null,
        coa_head: form.value.coa_head,
    };

    try {
        if (isEdit) {
            await api.put(`/coa/${editingCoa.value.coa_id}`, payload);
        } else {
            await api.post('/coa', payload);
        }

        showModal.value = false;
        await getCoa();
        Toast.fire({
            icon: 'success',
            title: isEdit ? 'Berhasil memperbarui COA' : 'Berhasil menambahkan COA',
        });
    } catch (e) {
        const data = e.response?.data;
        formErrors.value = data?.errors || {};
        formError.value = data?.errors ? '' : data?.message || 'Gagal menyimpan COA.';
        console.error(e);
    } finally {
        saving.value = false;
    }
};

const hapusCoa = async (coa) => {
    const result = await Swal.fire({
        title: `Yakin hapus ${coa.coa_no}?`,
        text: 'Data tidak bisa dikembalikan',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Ya, hapus',
        cancelButtonText: 'Batal',
    });

    if (!result.isConfirmed) return;

    try {
        await api.delete(`/coa/${coa.coa_id}`);
        await getCoa();
        Toast.fire({ icon: 'success', title: 'Berhasil menghapus COA' });
    } catch (e) {
        // mis. 409: akun masih punya anak
        Swal.fire({
            icon: 'error',
            title: 'Gagal menghapus',
            text: e.response?.data?.message || 'Gagal menghapus COA',
        });
        console.error(e);
    }
};

onMounted(() => {
    getCoa();
    getCoaTypes();
});
</script>

<template>
    <section class="p-6">
        <h1 class="text-3xl font-bold mb-8">Data CoA</h1>

        <div class="flex flex-col md:flex-row justify-between items-center mb-6">
            <div class="w-full md:w-1/3 mb-4 md:mb-0">
                <input
                    v-model="search"
                    type="text"
                    placeholder="Cari kode, nama, atau tipe..."
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

        <p v-if="error" role="alert" class="mb-4 text-sm text-red-600">{{ error }}</p>

        <div class="overflow-x-auto bg-white rounded-lg shadow">
            <table class="w-full table-sm">
                <thead>
                    <tr class="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                        <th class="py-3 px-6 text-left">No CoA</th>
                        <th class="py-3 px-2 text-left">Nama CoA</th>
                        <th class="py-3 px-6 text-left">Tipe CoA</th>
                        <th class="py-3 px-6 text-left">Jenis</th>
                        <th class="py-3 w-25 text-center">Actions</th>
                    </tr>
                </thead>
                <tbody class="text-gray-600 text-sm">
                    <tr v-if="loading">
                        <td colspan="5" class="text-center py-10 text-gray-500">Memuat data...</td>
                    </tr>
                    <tr v-else-if="filteredCoas.length === 0">
                        <td colspan="5" class="text-center py-10 text-gray-500">
                            CoA masih kosong.
                        </td>
                    </tr>
                    <template v-else>
                        <tr
                            v-for="coa in filteredCoas"
                            :key="coa.coa_id"
                            class="border-b border-gray-200 hover:bg-gray-100"
                        >
                            <td class="py-1 px-6 text-left font-mono">
                                <span
                                    :style="{ paddingLeft: levelOf(coa) * 20 + 'px' }"
                                    :class="coa.coa_head ? 'font-semibold' : ''"
                                >
                                    {{ coa.coa_no }}
                                </span>
                            </td>
                            <td class="py-1 px-2 text-left">
                                <span
                                    :style="{ paddingLeft: levelOf(coa) * 20 + 'px' }"
                                    :class="coa.coa_head ? 'font-semibold' : ''"
                                >
                                    {{ coa.coa_name }}
                                </span>
                            </td>
                            <td class="py-1 px-6 text-left">{{ coa.type?.coa_type_name }}</td>
                            <td class="py-1 px-6 text-left">
                                <span
                                    :class="[
                                        'rounded-full px-2 py-0.5 text-xs',
                                        coa.coa_head
                                            ? 'bg-amber-100 text-amber-700'
                                            : 'bg-emerald-100 text-emerald-700',
                                    ]"
                                >
                                    {{ coa.coa_head ? 'Header' : 'Detail' }}
                                </span>
                            </td>
                            <td class="py-1 px-6 text-left">
                                <div class="flex items-center justify-start gap-2">
                                    <button
                                        type="button"
                                        @click="bukaModal(coa)"
                                        :class="[actionButtonClass, 'border-blue-200 bg-blue-100']"
                                        title="Edit CoA"
                                        aria-label="Edit CoA"
                                    >
                                        <i class="fa-solid fa-pencil"></i>
                                    </button>
                                    <button
                                        type="button"
                                        @click="hapusCoa(coa)"
                                        :class="[
                                            actionButtonClass,
                                            'border-red-200 bg-red-100 cursor-pointer',
                                        ]"
                                        title="Hapus CoA"
                                        aria-label="Hapus CoA"
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
    </section>

    <Modal :show="showModal" :title="editingCoa ? 'Edit CoA' : 'Tambah CoA'" @close="tutupModal">
        <form class="space-y-4" @submit.prevent="simpanCoa">
            <div>
                <label for="coa-parent" class="mb-1 block text-sm font-medium text-slate-700">
                    Parent (opsional)
                </label>
                <Multiselect
                    id="coa-parent"
                    v-model="form.coa_parent"
                    :options="parentSelectOptions"
                    value-prop="value"
                    label="label"
                    placeholder="Tanpa parent (paling atas)"
                    :searchable="true"
                />
                <p v-if="formErrors.coa_parent" class="mt-1 text-sm text-red-600">
                    {{ formErrors.coa_parent }}
                </p>
            </div>

            <div>
                <label for="coa-type" class="mb-1 block text-sm font-medium text-slate-700">
                    Tipe
                </label>
                <Multiselect
                    id="coa-type"
                    v-model="form.coa_type_id"
                    :options="coaTypeSelectOptions"
                    value-prop="value"
                    label="label"
                    placeholder="Pilih tipe"
                    :searchable="true"
                    :can-clear="false"
                />
                <p v-if="formErrors.coa_type_id" class="mt-1 text-sm text-red-600">
                    {{ formErrors.coa_type_id }}
                </p>
            </div>

            <div>
                <label for="coa-no" class="mb-1 block text-sm font-medium text-slate-700">
                    Kode
                </label>
                <input
                    id="coa-no"
                    v-model="form.coa_no"
                    @input="kodeManual = true"
                    type="text"
                    maxlength="50"
                    class="w-full rounded-md border border-gray-300 px-3 py-2 font-mono focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    placeholder="100.000.000"
                />
                <p v-if="formErrors.coa_no" class="mt-1 text-sm text-red-600">
                    {{ formErrors.coa_no }}
                </p>
            </div>

            <div>
                <label for="coa-name" class="mb-1 block text-sm font-medium text-slate-700">
                    Nama
                </label>
                <input
                    id="coa-name"
                    v-model="form.coa_name"
                    type="text"
                    maxlength="150"
                    class="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    placeholder="Nama akun"
                />
                <p v-if="formErrors.coa_name" class="mt-1 text-sm text-red-600">
                    {{ formErrors.coa_name }}
                </p>
            </div>

            <div>
                <label class="flex items-center gap-2 text-sm text-slate-700">
                    <input v-model="form.coa_head" type="checkbox" class="h-4 w-4" />
                    Akun header (pengelompok, tidak dipakai transaksi)
                </label>
                <p v-if="formErrors.coa_head" class="mt-1 text-sm text-red-600">
                    {{ formErrors.coa_head }}
                </p>
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
