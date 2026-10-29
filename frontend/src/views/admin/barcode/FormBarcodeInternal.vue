<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import api from '@/api';
import Swal from 'sweetalert2';
import Select2Field from '@/components/Select2Field.vue';

const router = useRouter();

const partais = ref([]);
const errors = ref({});
const saving = ref(false);
const loadingNext = ref(false);
const nextInfo = ref('');

const form = ref({
    goods_partai_id: '',
    tahun: new Date().getFullYear(),
    item_code_from: '',
    item_code_to: '',
});

const inputClass = (field, locked = false) => [
    'w-full px-4 py-2 rounded-md border focus:outline-none focus:ring-2',
    locked ? 'bg-gray-100 text-gray-500 cursor-not-allowed' : 'bg-white',
    errors.value[field]
        ? 'border-red-400 focus:ring-red-400'
        : 'border-gray-300 focus:ring-blue-500',
];

const toInt = (v) => (v === '' || v === null || v === undefined ? null : Number(v));

// Sama dengan tahunKeHuruf di backend: 2026 menjadi BF.
const kodeTahun = computed(() => {
    const y = toInt(form.value.tahun);
    if (y === null || Number.isNaN(y)) return '-';
    const huruf = 'JABCDEFGHI';
    const yy = y % 100;
    return huruf[Math.floor(yy / 10)] + huruf[yy % 10];
});

const totalKode = computed(() => {
    const from = toInt(form.value.item_code_from);
    const to = toInt(form.value.item_code_to);
    if (from === null || to === null || to < from) return 0;
    return to - from + 1;
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

// Isi otomatis "Item kode dari" dengan nomor urut berikutnya untuk partai dan tahun terpilih.
const cekNomorBerikutnya = async () => {
    nextInfo.value = '';
    form.value.item_code_from = '';

    const partaiID = toInt(form.value.goods_partai_id);
    const tahun = toInt(form.value.tahun);
    if (!partaiID || !tahun || tahun < 1990 || tahun > 2100) return;

    loadingNext.value = true;
    try {
        const res = await api.get('/barcode/next-serial', {
            params: { goods_partai_id: partaiID, tahun },
        });
        form.value.item_code_from = res.data.next;
        nextInfo.value = `Nomor urut berikutnya: ${res.data.next}`;
    } catch (e) {
        nextInfo.value = e.response?.data?.message || 'Gagal mengambil nomor urut';
    } finally {
        loadingNext.value = false;
    }
};

const simpan = async () => {
    if (saving.value) return;
    saving.value = true;

    const konfirmasi = await Swal.fire({
        icon: 'question',
        title: `Generate ${totalKode.value} barcode?`,
        text: 'Pastikan rentang nomor barcode sudah benar.',
        showCancelButton: true,
        confirmButtonText: 'Ya, generate',
        cancelButtonText: 'Batal',
    });

    if (!konfirmasi.isConfirmed) {
        saving.value = false;
        return;
    }

    errors.value = {};

    try {
        const res = await api.post('/barcode/internal', {
            goods_partai_id: toInt(form.value.goods_partai_id),
            tahun: toInt(form.value.tahun),
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
            // Nomor urut bisa sudah terpakai oleh generate lain, ambil ulang yang terbaru.
            if (errors.value.item_code_from) cekNomorBerikutnya();
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

watch(() => [form.value.goods_partai_id, form.value.tahun], cekNomorBerikutnya);

onMounted(getPartai);
</script>

<template>
    <section class="p-6">
        <h1 class="text-3xl font-bold mb-8">Generate Barcode Internal</h1>

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

            <div>
                <label for="tahun" class="mb-1 block text-sm font-medium text-gray-700"
                    >Tahun</label
                >
                <input id="tahun" v-model="form.tahun" type="number" :class="inputClass('tahun')" />
                <p v-if="errors.tahun" class="mt-1 text-sm text-red-600">{{ errors.tahun }}</p>
                <p v-else class="mt-1 text-sm text-gray-500">
                    Kode tahun pada barcode: {{ kodeTahun }}
                </p>
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
                        disabled
                        :class="inputClass('item_code_from', true)"
                    />
                    <p v-if="errors.item_code_from" class="mt-1 text-sm text-red-600">
                        {{ errors.item_code_from }}
                    </p>
                    <p v-else-if="loadingNext" class="mt-1 text-sm text-gray-500">
                        Mengecek nomor urut...
                    </p>
                    <p v-else-if="nextInfo" class="mt-1 text-sm text-gray-500">{{ nextInfo }}</p>
                    <p v-else class="mt-1 text-sm text-gray-500">
                        Terisi otomatis setelah partai dan tahun dipilih.
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
                    class="px-4 py-2 rounded-xl bg-emerald-600 text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
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
