<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import api from '@/api';
import Swal from 'sweetalert2';
import Select2Field from '@/components/Select2Field.vue';

const router = useRouter();

const partais = ref([]);
const errors = ref({});
const saving = ref(false);

const form = ref({
    header: '',
    tahun: new Date().getFullYear(),
    item_code_digit: 5,
    goods_partai_id: '',
    item_code_from: '',
    item_code_to: '',
});

const inputClass = (field) => [
    'w-full px-4 py-2 rounded-md border bg-white focus:outline-none focus:ring-2',
    errors.value[field]
        ? 'border-red-400 focus:ring-red-400'
        : 'border-gray-300 focus:ring-blue-500',
];

const toInt = (v) => (v === '' || v === null || v === undefined ? null : Number(v));

const totalKode = computed(() => {
    const from = toInt(form.value.item_code_from);
    const to = toInt(form.value.item_code_to);
    if (from === null || to === null || to < from) return 0;
    return to - from + 1;
});

// Contoh kode, dibentuk sama seperti buildExternalCode di backend: header + angka dengan nol di depan.
const kode = (serial) => {
    const digit = toInt(form.value.item_code_digit);
    if (serial === null || !digit || digit < 1 || digit > 9) return '';
    return form.value.header.trim() + String(serial).padStart(digit, '0');
};

const contohKode = computed(() => {
    const from = toInt(form.value.item_code_from);
    const to = toInt(form.value.item_code_to);
    if (!form.value.header.trim() || from === null) return '';
    if (to === null || to <= from) return kode(from);
    return `${kode(from)} s/d ${kode(to)}`;
});

const getPartai = async () => {
    try {
        const res = await api.get('/goods-partai');
        partais.value = res.data.data;
    } catch (e) {
        Swal.fire({ icon: 'error', title: 'Gagal mengambil data partai' });
        console.error(e);
    }
};

const labelPartai = (p) =>
    `${String(p.goods_partai_code).padStart(3, '0')} - ${p.goods_partai_name}`;

const partaiOptions = computed(() =>
    partais.value.map((p) => ({ id: String(p.goods_partai_id), text: labelPartai(p) })),
);

const simpan = async () => {
    errors.value = {};
    saving.value = true;

    try {
        const res = await api.post('/barcode/external', {
            header: form.value.header,
            tahun: toInt(form.value.tahun),
            item_code_digit: toInt(form.value.item_code_digit),
            goods_partai_id: toInt(form.value.goods_partai_id),
            item_code_from: toInt(form.value.item_code_from),
            item_code_to: toInt(form.value.item_code_to),
        });

        const d = res.data.data;
        await Swal.fire({
            icon: 'success',
            title: 'Berhasil generate barcode',
            text: `${d.total} kode: ${d.code_first} s/d ${d.code_last}`,
        });
        router.push('/admin/barcode');
    } catch (e) {
        if (e.response?.status === 422 && e.response.data.errors) {
            errors.value = e.response.data.errors;
        } else {
            Swal.fire({
                icon: 'error',
                title: 'Gagal generate barcode',
                text: e.response?.data?.message || 'Terjadi kesalahan pada server',
            });
            console.error(e);
        }
    } finally {
        saving.value = false;
    }
};

onMounted(getPartai);
</script>

<template>
    <section class="p-6">
        <h1 class="text-3xl font-bold mb-8">Generate Barcode External</h1>

        <form
            @submit.prevent="simpan"
            class="max-w-2xl space-y-5 rounded-lg bg-white p-6 shadow"
            novalidate
        >
            <div>
                <label for="partai" class="mb-1 block text-sm font-medium text-gray-700">
                    Partai
                </label>
                <Select2Field
                    id="partai"
                    v-model="form.goods_partai_id"
                    :options="partaiOptions"
                    placeholder="Pilih partai"
                    :class="{ 'field-error': errors.goods_partai_id }"
                />
                <p v-if="errors.goods_partai_id" class="mt-1 text-sm text-red-600">
                    {{ errors.goods_partai_id }}
                </p>
            </div>

            <div class="grid gap-5 md:grid-cols-2">
                <div>
                    <label for="header" class="mb-1 block text-sm font-medium text-gray-700">
                        Header
                    </label>
                    <input
                        id="header"
                        v-model="form.header"
                        type="text"
                        placeholder="Contoh: BFTP"
                        :class="inputClass('header')"
                    />
                    <p v-if="errors.header" class="mt-1 text-sm text-red-600">{{ errors.header }}</p>
                    <p v-else class="mt-1 text-sm text-gray-500">Huruf, angka, dan simbol tanpa spasi.</p>
                </div>

                <div>
                    <label for="digit" class="mb-1 block text-sm font-medium text-gray-700">
                        Item kode digit
                    </label>
                    <input
                        id="digit"
                        v-model="form.item_code_digit"
                        type="number"
                        min="1"
                        max="9"
                        :class="inputClass('item_code_digit')"
                    />
                    <p v-if="errors.item_code_digit" class="mt-1 text-sm text-red-600">
                        {{ errors.item_code_digit }}
                    </p>
                    <p v-else class="mt-1 text-sm text-gray-500">Jumlah digit nomor urut, 1 sampai 9.</p>
                </div>
            </div>

            <div>
                <label for="tahun" class="mb-1 block text-sm font-medium text-gray-700">Tahun</label>
                <input id="tahun" v-model="form.tahun" type="number" :class="inputClass('tahun')" />
                <p v-if="errors.tahun" class="mt-1 text-sm text-red-600">{{ errors.tahun }}</p>
            </div>

            <div class="grid gap-5 md:grid-cols-2">
                <div>
                    <label for="from" class="mb-1 block text-sm font-medium text-gray-700">
                        Item kode dari
                    </label>
                    <input
                        id="from"
                        v-model="form.item_code_from"
                        type="number"
                        min="1"
                        :class="inputClass('item_code_from')"
                    />
                    <p v-if="errors.item_code_from" class="mt-1 text-sm text-red-600">
                        {{ errors.item_code_from }}
                    </p>
                </div>

                <div>
                    <label for="to" class="mb-1 block text-sm font-medium text-gray-700">
                        Item kode sampai
                    </label>
                    <input
                        id="to"
                        v-model="form.item_code_to"
                        type="number"
                        min="1"
                        :class="inputClass('item_code_to')"
                    />
                    <p v-if="errors.item_code_to" class="mt-1 text-sm text-red-600">
                        {{ errors.item_code_to }}
                    </p>
                    <p v-else-if="totalKode" class="mt-1 text-sm text-gray-500">
                        Akan dibuat {{ totalKode }} kode
                    </p>
                </div>
            </div>

            <div v-if="contohKode" class="rounded-md bg-gray-50 px-4 py-3">
                <p class="text-sm text-gray-500">Contoh kode</p>
                <p class="font-mono text-sm font-semibold text-gray-800">{{ contohKode }}</p>
            </div>

            <div class="flex justify-end gap-2 pt-2">
                <RouterLink
                    to="/admin/barcode"
                    class="px-4 py-2 rounded-xl border border-gray-300 text-gray-700 hover:bg-gray-100"
                >
                    Batal
                </RouterLink>
                <button
                    type="submit"
                    :disabled="saving"
                    class="px-4 py-2 rounded-xl bg-blue-600 text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                    {{ saving ? 'Menyimpan...' : 'Generate barcode' }}
                </button>
            </div>
        </form>
    </section>
</template>

<style>
.app-select2-field.field-error .select2-container--default .select2-selection--single {
    border-color: #f87171;
}
</style>