<script setup>
import { ref, reactive, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import api from '@/api';
import { useRentang } from '@/composables/useRentang';

const route = useRoute();
const router = useRouter();

const partaiId = () => route.params.id;
const partaiName = ref('');
const partaiError = ref('');

const fieldKeys = ['min', 'max', 'nominal'];

const configs = [
    {
        key: 'tarra',
        resource: 'tarra',
        label: 'tarra',
        formTitle: 'Data Tarra',
        listTitle: 'Data Tarra',
        idKey: 'goods_partai_tarra_id',
        digits: 0,
        labels: { min: 'Bruto Bal Min', max: 'Bruto Bal Max', nominal: 'Tarra Nominal' },
        fields: {
            min: 'goods_partai_tarra_berat_min',
            max: 'goods_partai_tarra_berat_max',
            nominal: 'goods_partai_tarra_nominal',
        },
    },
    {
        key: 'decimal',
        resource: 'decimal',
        label: 'desimal',
        formTitle: 'Data Partai Desimal',
        listTitle: 'Data Desimal',
        idKey: 'goods_partai_decimal_id',
        digits: 2,
        labels: { min: 'Desimal Min', max: 'Desimal Max', nominal: 'Nominal Potongan' },
        fields: {
            min: 'goods_partai_decimal_min',
            max: 'goods_partai_decimal_max',
            nominal: 'goods_partai_bruto_nominal',
        },
        allowNegativeFields: ['nominal'],
    },
];

const panels = configs.map((config) => ({
    ...config,
    state: reactive(useRentang({ partaiId, ...config })),
}));

const formatAngka = (value, minDigits = 0) =>
    Number(value).toLocaleString('en-US', {
        minimumFractionDigits: minDigits,
        maximumFractionDigits: 4,
    });

const sanitizeNumericInput = (value, allowNegative) => {
    const negative = allowNegative && value.startsWith('-');
    const unsigned = value.replace(/[^\d.]/g, '');
    const decimalIndex = unsigned.indexOf('.');
    const number =
        decimalIndex === -1
            ? unsigned
            : `${unsigned.slice(0, decimalIndex + 1)}${unsigned.slice(decimalIndex + 1).replace(/\./g, '')}`;

    return `${negative ? '-' : ''}${number}`;
};

const updateNumericInput = (event, panel, key) => {
    const allowNegative = panel.allowNegativeFields?.includes(key) ?? false;
    const value = sanitizeNumericInput(event.target.value, allowNegative);
    event.target.value = value;
    panel.state.form[key] = value;
};

const loadPartai = async () => {
    partaiError.value = '';

    try {
        const res = await api.get(`/goods-partai/${partaiId()}`);
        partaiName.value = res.data.data.goods_partai_name;
    } catch (e) {
        partaiError.value = e.response?.data?.message || 'Gagal mengambil data partai.';
        console.error(e);
    }
};

const loadAll = () => {
    loadPartai();
    panels.forEach((panel) => panel.state.load());
};

const kembali = () => router.push('/admin/goods-partai');

onMounted(loadAll);
watch(() => route.params.id, loadAll);
</script>

<template>
    <section class="min-h-screen bg-slate-50 p-4 md:p-6">
        <div class="mx-auto max-w-7xl space-y-6">
            <header>
                <h1 class="text-2xl font-bold tracking-tight text-slate-800">Tarra &amp; Desimal</h1>
                <p class="mt-1 text-sm text-slate-500">
                    Kelola rentang bruto dan nominal untuk partai
                    <span class="font-semibold text-slate-700">
                        {{ partaiName || '...' }}
                    </span>
                </p>
            </header>

            <p
                v-if="partaiError"
                role="alert"
                class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
            >
                {{ partaiError }}
            </p>

            <div class="grid grid-cols-1 gap-6 xl:grid-cols-2">
                <form
                    v-for="panel in panels"
                    :key="panel.key"
                    class="overflow-hidden rounded-md border border-slate-200 bg-white shadow-sm"
                    @submit.prevent="panel.state.save()"
                >
                    <div class="border-b border-slate-200 px-5 py-4 md:px-6">
                        <div class="flex items-center gap-3">
                            <div>
                                <h2 class="font-semibold text-slate-800">{{ panel.formTitle }}</h2>
                                <p class="mt-0.5 text-sm text-slate-500">
                                    Isi nilai rentang dan nominal
                                </p>
                            </div>
                        </div>
                    </div>

                    <div class="space-y-5 p-5 md:p-6">
                        <div
                            class="grid gap-1.5 sm:grid-cols-[minmax(0,1fr)_minmax(0,1.35fr)] sm:items-start sm:gap-4"
                        >
                            <label
                                :for="`${panel.key}-partai`"
                                class="pt-2 text-sm font-medium text-slate-700"
                            >
                                Nama Partai
                            </label>
                            <div>
                                <input
                                    :id="`${panel.key}-partai`"
                                    :value="partaiName"
                                    type="text"
                                    readonly
                                    class="w-full rounded-md border border-slate-200 bg-slate-100 px-3.5 py-2.5 text-sm text-slate-600"
                                />
                            </div>
                        </div>

                        <div
                            v-for="key in fieldKeys"
                            :key="key"
                            class="grid gap-1.5 sm:grid-cols-[minmax(0,1fr)_minmax(0,1.35fr)] sm:items-start sm:gap-4"
                        >
                            <label
                                :for="`${panel.key}-${key}`"
                                class="pt-2 text-sm font-medium text-slate-700"
                            >
                                {{ panel.labels[key] }}
                            </label>
                            <div>
                                <input
                                    :id="`${panel.key}-${key}`"
                                    :value="panel.state.form[key]"
                                    type="text"
                                    inputmode="decimal"
                                    placeholder="0"
                                    @input="updateNumericInput($event, panel, key)"
                                    :aria-invalid="Boolean(panel.state.errors[key])"
                                    class="w-full rounded-md border border-slate-300 px-3.5 py-2.5 text-sm text-slate-800 transition placeholder:text-slate-400 focus:border-emerald-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
                                />
                                <p
                                    v-if="panel.state.errors[key]"
                                    class="mt-1.5 text-sm text-red-600"
                                >
                                    {{ panel.state.errors[key] }}
                                </p>
                            </div>
                        </div>

                        <p
                            v-if="panel.state.formError"
                            role="alert"
                            class="rounded-md border border-red-200 bg-red-50 px-3.5 py-2.5 text-sm text-red-700"
                        >
                            {{ panel.state.formError }}
                        </p>
                    </div>

                    <div
                        class="flex flex-col-reverse gap-3 border-t border-slate-200 bg-slate-50 px-5 py-4 sm:flex-row sm:justify-end md:px-6"
                    >
                        <button
                            type="button"
                            @click="kembali"
                            class="inline-flex items-center justify-center gap-2 rounded-md border border-slate-300 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 transition hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2"
                        >
                            <i class="fa-solid fa-chevron-left" aria-hidden="true"></i>
                            Kembali
                        </button>
                        <button
                            type="submit"
                            :disabled="panel.state.saving"
                            class="inline-flex items-center justify-center gap-2 rounded-md bg-emerald-600 px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                            <i class="fa-solid fa-floppy-disk" aria-hidden="true"></i>
                            {{ panel.state.saving ? 'Menyimpan...' : 'Simpan' }}
                        </button>
                    </div>
                </form>
            </div>

            <section
                v-for="panel in panels"
                :key="panel.key"
                class="overflow-hidden rounded-md border border-slate-200 bg-white shadow-sm"
            >
                <div
                    class="flex flex-col gap-1 border-b border-slate-200 px-5 py-4 sm:flex-row sm:items-center sm:justify-between md:px-6"
                >
                    <div>
                        <h2 class="font-semibold text-slate-800">{{ panel.listTitle }}</h2>
                        <p class="mt-0.5 text-sm text-slate-500">
                            Daftar rentang yang sudah tersimpan
                        </p>
                    </div>
                    <span
                        class="inline-flex w-fit items-center rounded-full bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700"
                    >
                        {{ panel.state.rows.length }} data
                    </span>
                </div>

                <div class="overflow-x-auto p-4 md:p-6">
                    <p
                        v-if="panel.state.loadError"
                        role="alert"
                        class="mb-4 rounded-md border border-red-200 bg-red-50 px-3.5 py-2.5 text-sm text-red-700"
                    >
                        {{ panel.state.loadError }}
                    </p>

                    <table class="w-full min-w-[640px] table-auto text-sm">
                        <thead>
                            <tr
                                class="bg-gray-200 text-xs font-semibold uppercase tracking-wide text-gray-600"
                            >
                                <th class="px-4 py-3 text-center">#</th>
                                <th
                                    v-for="key in fieldKeys"
                                    :key="key"
                                    class="px-4 py-3 text-center"
                                >
                                    {{ panel.labels[key] }}
                                </th>
                                <th class="px-4 py-3 text-center">Aksi</th>
                            </tr>
                        </thead>
                        <tbody class="text-slate-600">
                            <tr v-if="panel.state.loading">
                                <td colspan="5" class="px-4 py-10 text-center text-slate-500">
                                    Memuat data...
                                </td>
                            </tr>
                            <tr v-else-if="panel.state.rows.length === 0">
                                <td colspan="5" class="px-4 py-10 text-center text-slate-500">
                                    Belum ada data.
                                </td>
                            </tr>
                            <template v-else>
                                <tr
                                    v-for="(row, index) in panel.state.rows"
                                    :key="row[panel.idKey]"
                                    class="border-b border-slate-100 transition-colors last:border-b-0 hover:bg-slate-50"
                                >
                                    <td class="px-4 py-3 text-center text-slate-400">
                                        {{ index + 1 }}
                                    </td>
                                    <td
                                        v-for="key in fieldKeys"
                                        :key="key"
                                        class="px-4 py-3 text-center font-medium text-slate-700"
                                    >
                                        {{ formatAngka(row[panel.fields[key]], panel.digits) }}
                                    </td>
                                    <td class="px-4 py-3 text-center">
                                        <button
                                            type="button"
                                            @click="panel.state.remove(row)"
                                            class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-red-200 bg-red-50 text-red-600 transition hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                                            title="Hapus"
                                            aria-label="Hapus"
                                        >
                                            <i class="fa-solid fa-trash-can"></i>
                                        </button>
                                    </td>
                                </tr>
                            </template>
                        </tbody>
                    </table>
                </div>
            </section>
        </div>
    </section>
</template>