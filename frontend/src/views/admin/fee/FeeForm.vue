<script setup>
import { ref, reactive, onMounted } from 'vue';
import api from '@/api';
import Swal from 'sweetalert2';

const Toast = Swal.mixin({
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
});

const fees = ref([]);
const form = reactive({});
const loading = ref(false);
const saving = ref(false);
const loadError = ref('');
const errors = ref({});
const formError = ref('');

const load = async () => {
    loading.value = true;
    loadError.value = '';

    try {
        const res = await api.get('/fee');
        fees.value = res.data.data;
        fees.value.forEach((fee) => {
            form[fee.fee_slug] = fee.fee_cost;
        });
    } catch (e) {
        loadError.value = e.response?.data?.message || 'Gagal mengambil data biaya.';
        console.error(e);
    } finally {
        loading.value = false;
    }
};

const save = async () => {
    if (saving.value) return;

    saving.value = true;
    errors.value = {};
    formError.value = '';

    const payload = {
        fees: fees.value.map((fee) => ({
            fee_slug: fee.fee_slug,
            fee_cost: form[fee.fee_slug] === '' ? null : Number(form[fee.fee_slug]),
        })),
    };

    try {
        await api.put('/fee', payload);
        await load();
        Toast.fire({ icon: 'success', title: 'Berhasil menyimpan biaya' });
    } catch (e) {
        const data = e.response?.data;
        errors.value = data?.errors || {};
        formError.value = data?.errors?.fees || (data?.errors ? '' : data?.message || 'Gagal menyimpan biaya.');
        console.error(e);
    } finally {
        saving.value = false;
    }
};

onMounted(load);
</script>

<template>
    <section class="p-6">
        <h1 class="mb-8 text-3xl font-bold">Data Biaya</h1>

        <form class="max-w-2xl overflow-hidden rounded-lg bg-white shadow" @submit.prevent="save">
            <div class="bg-blue-600 px-6 py-3 text-white">Biaya</div>

            <div class="space-y-4 p-6">
                <p v-if="loadError" role="alert" class="text-sm text-red-600">{{ loadError }}</p>
                <p v-if="loading" class="text-gray-500">Memuat data...</p>

                <div
                    v-for="fee in fees"
                    :key="fee.fee_slug"
                    class="grid grid-cols-3 gap-4"
                >
                    <label :for="`fee-${fee.fee_slug}`" class="pt-2 font-semibold">
                        {{ fee.fee_name }}
                    </label>
                    <div class="col-span-2">
                        <input
                            :id="`fee-${fee.fee_slug}`"
                            v-model="form[fee.fee_slug]"
                            type="number"
                            step="any"
                            min="0"
                            placeholder="0"
                            class="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                        <p v-if="errors[`fees.${fee.fee_slug}`]" class="mt-1 text-sm text-red-600">
                            {{ errors[`fees.${fee.fee_slug}`] }}
                        </p>
                    </div>
                </div>

                <p v-if="formError" role="alert" class="text-sm text-red-600">{{ formError }}</p>
            </div>

            <div class="flex justify-end bg-gray-50 px-6 py-4">
                <button
                    type="submit"
                    :disabled="saving || loading"
                    class="rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
                >
                    <i class="fa-solid fa-floppy-disk mr-1"></i>
                    {{ saving ? 'Menyimpan...' : 'Simpan' }}
                </button>
            </div>
        </form>
    </section>
</template>