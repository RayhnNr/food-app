<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue';
import api from '@/api';
import Swal from 'sweetalert2';
import { useCart } from '@/composables/useCart';
import Modal from '@/components/Modal.vue';

const coatypes = ref([]);
const loading = ref(true);
const error = ref(null);
const search = ref('');
const actionButtonClass =
    'inline-flex h-9 w-9 items-center justify-center rounded-md border text-gray-600 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500';

const getCoaType = async () => {
    loading.value = true;
    error.value = null;

    try {
        const res = await api.get('/coa-types', {
            params: {
                q: search.value,
                _sort: 'coa_type_id',
                _order: 'DESC',
            },
        });
        coatypes.value = res.data.data;
    } catch (e) {
        error.value = 'Gagal mengambil data. Pastikan npm run api sudah jalan.';
        console.error(e);
    } finally {
        loading.value = false;
    }
};

onMounted(getCoaType);

let timer;
watch(search, () => {
    clearTimeout(timer);
    timer = setTimeout(getCoaType, 400);
});

onBeforeUnmount(() => {
    clearTimeout(timer);
});
</script>

<template>
    <section class="p-6">
        <h1 class="text-3xl font-bold mb-8">Data CoA Type</h1>

        <div class="flex flex-col md:flex-row justify-between items-center mb-6">
            <div class="w-full md:w-1/3 mb-4 md:mb-0">
                <input
                    v-model="search"
                    type="text"
                    placeholder="Cari kategori..."
                    class="w-full px-4 py-2 rounded-md border border-gray-300 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
            </div>
        </div>

        <!-- User Table -->
        <div class="overflow-x-auto bg-white rounded-lg shadow">
            <table class="w-full table-auto">
                <thead>
                    <tr class="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                        <th class="py-3 px-6 text-left">ID</th>
                        <th class="py-3 px-6 text-left">CoA Type Name</th>
                        <th class="py-3 px-6 text-left">Menambah</th>
                        <th class="py-3 px-6 text-left">Mengurangi</th>
                    </tr>
                </thead>
                <tbody class="text-gray-600 text-sm">
                    <tr v-if="coatypes.length === 0">
                        <td colspan="7" class="text-center py-10 text-gray-500">
                            coatype masih kosong.
                        </td>
                    </tr>
                    <tr
                        v-else
                        v-for="(coatype, index) in coatypes"
                        :key="coatype.id"
                        class="border-b border-gray-200 hover:bg-gray-100"
                    >
                        <td class="py-3 px-6 text-left">
                            {{ index + 1 }}
                        </td>
                        <td class="py-3 px-6 text-left">{{ coatype.coa_type_name }}</td>
                        <td class="py-3 px-6 text-left">{{ coatype.coa_type_increase }}</td>
                        <td class="py-3 px-6 text-left">{{ coatype.coa_type_decrease }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </section>
</template>
