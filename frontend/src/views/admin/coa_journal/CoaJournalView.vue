```vue
<script setup>
import { ref, onMounted } from 'vue';
import Select2 from 'vue3-select2-component';
import Swal from 'sweetalert2';
import api from '@/api';

const items = ref([]);
const coaOptions = ref([]);
const loading = ref(false);
const submitting = ref(false);

function fillMappings(rows) {
    items.value = rows.map((row) => ({
        id_type_coa: Number(row.id_type_coa),
        name_coa_journal: row.name_coa_journal,
        coa_id: row.coa_id == null ? '' : String(row.coa_id),
    }));
}

async function fetchJournalMappings() {
    const response = await api.get('/coa-journal');
    return response.data.data;
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
        const [journalRows, options] = await Promise.all([
            fetchJournalMappings(),
            fetchCoaOptions(),
        ]);

        fillMappings(journalRows);
        coaOptions.value = options;
    } catch (error) {
        console.error('Gagal mengambil data mapping COA:', error);
        await Swal.fire({
            icon: 'error',
            title: 'Gagal memuat data',
            text: error.response?.data?.message || 'Gagal memuat data mapping COA.',
            confirmButtonColor: '#059669',
        });
    } finally {
        loading.value = false;
    }
}

async function submitForm() {
    if (submitting.value || loading.value || items.value.length === 0) return;

    submitting.value = true;

    try {
        const payload = {
            items: items.value.map((item) => ({
                id_type_coa: item.id_type_coa,
                coa_id: item.coa_id == null || item.coa_id === '' ? null : Number(item.coa_id),
            })),
        };

        const response = await api.put('/coa-journal', payload);
        fillMappings(response.data.data);
        await Swal.fire({
            icon: 'success',
            title: 'Berhasil',
            text: 'Mapping COA berhasil disimpan.',
            confirmButtonColor: '#059669',
        });
    } catch (error) {
        console.error('Gagal menyimpan mapping COA:', error);
        const data = error.response?.data;
        if (error.response?.status === 422 && data?.errors) {
            const messages = Object.entries(data.errors).map(([typeId, message]) => {
                const row = items.value.find((item) => String(item.id_type_coa) === typeId);
                return `${row?.name_coa_journal || `Tipe COA ${typeId}`}: ${message}`;
            });
            await Swal.fire({
                icon: 'warning',
                title: 'Periksa mapping COA',
                text: messages.join('\n'),
                confirmButtonColor: '#059669',
            });
        } else {
            await Swal.fire({
                icon: 'error',
                title: 'Gagal menyimpan',
                text: data?.message || 'Mapping COA gagal disimpan.',
                confirmButtonColor: '#059669',
            });
        }
    } finally {
        submitting.value = false;
    }
}

onMounted(loadPageData);
</script>

<template>
    <div class="min-h-screen bg-slate-50 p-4 md:p-6">
        <div class="mx-auto max-w-5xl space-y-6">
            <!-- Header -->
            <div>
                <h1 class="text-2xl font-bold tracking-tight text-slate-800">CoA Journal</h1>

                <p class="mt-1 text-sm text-slate-500">Atur akun CoA untuk setiap jenis CoA.</p>
            </div>

            <!-- Form -->
            <form
                @submit.prevent="submitForm"
                class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm"
            >
                <div class="border-b border-slate-200 px-5 py-4 md:px-6">
                    <h2 class="font-semibold text-slate-800">Konfigurasi Akun</h2>

                    <p class="mt-1 text-sm text-slate-500">
                        Pilih akun COA melalui dropdown yang tersedia.
                    </p>
                </div>

                <div v-if="loading" class="p-8 text-center text-slate-500">Memuat data COA...</div>

                <div v-else-if="items.length === 0" class="p-8 text-center text-sm text-slate-500">
                    Data jenis COA belum tersedia.
                </div>

                <div v-else class="divide-y divide-slate-100">
                    <div
                        v-for="(item, index) in items"
                        :key="item.id_type_coa"
                        class="grid gap-3 px-5 py-4 transition-colors hover:bg-slate-50/70 md:grid-cols-[180px_minmax(0,1fr)] md:items-center md:gap-6 md:px-6"
                    >
                        <!-- Type COA -->
                        <div>
                            <p class="text-sm font-medium text-slate-700">
                                #{{ item.id_type_coa }} {{ item.name_coa_journal }}
                            </p>

                            <!-- <p class="mt-1 text-xs text-slate-400">ID: {{ item.id_type_coa }}</p> -->
                        </div>

                        <!-- Dropdown -->
                        <div class="min-w-0">
                            <Select2
                                class="coa-journal-select2"
                                v-model="item.coa_id"
                                :options="coaOptions"
                                placeholder="Cari atau pilih akun COA..."
                                :settings="{
                                    width: '100%',
                                    allowClear: true,
                                    dropdownCssClass: 'coa-journal-select2-dropdown',
                                }"
                                :disabled="submitting"
                            />
                        </div>
                    </div>
                </div>

                <!-- Footer -->
                <div
                    class="flex flex-col-reverse gap-3 border-t border-slate-200 bg-slate-50 px-5 py-4 sm:flex-row sm:items-center sm:justify-between md:px-6"
                >
                    <p class="text-xs text-slate-500">Total {{ items.length }} jenis COA</p>

                    <button
                        type="submit"
                        :disabled="loading || submitting || items.length === 0"
                        class="inline-flex items-center justify-center gap-2 rounded-lg bg-emerald-600 px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                        <span>
                            {{ submitting ? 'Menyimpan...' : 'Simpan' }}
                        </span>
                    </button>
                </div>
            </form>
        </div>
    </div>
</template>

<style>
.coa-journal-select2 {
    width: 100%;
}

.coa-journal-select2 .select2-container {
    width: 100% !important;
}

.coa-journal-select2 .select2-container--default .select2-selection--single {
    position: relative;
    height: 42px;
    border: 1px solid #cbd5e1;
    border-radius: 6px;
    background: #fff;
    transition:
        border-color 150ms ease,
        box-shadow 150ms ease;
}

.coa-journal-select2 .select2-container--default .select2-selection--single:hover {
    border-color: #94a3b8;
}

.coa-journal-select2 .select2-container--focus .select2-selection--single,
.coa-journal-select2 .select2-container--open .select2-selection--single {
    border-color: #059669;
    box-shadow: 0 0 0 3px rgb(5 150 105 / 12%);
    outline: none;
}

.coa-journal-select2
    .select2-container--default
    .select2-selection--single
    .select2-selection__rendered {
    padding-left: 12px;
    padding-right: 34px;
    color: #334155;
    font-size: 14px;
    line-height: 40px;
}

.coa-journal-select2
    .select2-container--default
    .select2-selection--single
    .select2-selection__placeholder {
    color: #94a3b8;
}

.coa-journal-select2
    .select2-container--default
    .select2-selection--single
    .select2-selection__arrow {
    top: 1px;
    right: 6px;
    height: 40px;
}

.coa-journal-select2
    .select2-container--default
    .select2-selection--single
    .select2-selection__clear {
    position: absolute;
    top: 50%;
    right: 28px;
    z-index: 1;
    margin: 0;
    color: #64748b;
    transform: translateY(-50%);
}

.coa-journal-select2
    .select2-container--default
    .select2-selection--single
    .select2-selection__arrow {
    right: 2px;
}

.coa-journal-select2 .select2-container--disabled .select2-selection--single {
    background: #f1f5f9;
    cursor: not-allowed;
}

.coa-journal-select2-dropdown {
    overflow: hidden;
    border: 1px solid #e2e8f0 !important;
    border-radius: 6px;
    box-shadow: 0 8px 20px rgb(15 23 42 / 10%);
}

.coa-journal-select2-dropdown .select2-search__field {
    border-color: #cbd5e1 !important;
    border-radius: 4px;
    outline: none;
}

.coa-journal-select2-dropdown .select2-search__field:focus {
    border-color: #059669 !important;
    box-shadow: 0 0 0 2px rgb(5 150 105 / 12%);
}

.coa-journal-select2-dropdown .select2-results__option[aria-selected='true'] {
    background: #ecfdf5;
    color: #047857;
}

.coa-journal-select2-dropdown .select2-results__option--highlighted[aria-selected] {
    background: #059669;
    color: #fff;
}
</style>
