<script setup>
import { ref, watch, onMounted } from 'vue';
import api from '@/api';
import FoodCard from '@/components/FoodCard.vue';

const foods = ref([]);
const loading = ref(true);
const error = ref(null);
const search = ref('');

const getFoods = async () => {
    loading.value = true;
    error.value = null;
    try {
        const res = await api.get('/products', {
            params: { q: search.value },
        });
        foods.value = res.data.data;
    } catch (e) {
        error.value = 'Gagal mengambil data. Pastikan npm run api sudah jalan.';
    } finally {
        loading.value = false;
    }
};

onMounted(getFoods);

let timer;
watch(search, () => {
    clearTimeout(timer);
    timer = setTimeout(getFoods, 400);
});
</script>

<template>
    <section class="max-w-7xl mx-auto px-4 py-12 md:py-16">
        <div class="flex items-center justify-center gap-4 py-4">
            <div class="text-center">
                <h2 class="text-3xl md:text-5xl font-bold text-slate-800">
                    Daftar <span class="text-emerald-600">Makanan dan Minuman</span>
                </h2>
                <p class="mt-1 text-md text-slate-600">
                    Semua menu yang tersedia, pilih favoritmu.
                </p>
            </div>
        </div>

        <input
            v-model="search"
            type="text"
            placeholder="Cari makanan..."
            class="w-full px-4 py-2 rounded-md border border-slate-300 focus:outline-none focus:ring-4 focus:ring-emerald-300"
        />

        <p v-if="loading" class="mt-6 text-slate-500">Memuat data...</p>
        <p v-else-if="error" class="mt-6 text-red-600">{{ error }}</p>
        
        
        <div
            v-else-if="foods.length === 0"
            class="flex flex-col items-center justify-center px-6 py-16 text-center"
        >
            <div
                class="mb-5 flex h-20 w-20 items-center justify-center rounded-full bg-emerald-50"
            >
                <i class="fas fa-search text-3xl text-emerald-600"></i>
            </div>

            <h3 class="mb-2 text-lg font-semibold text-gray-800">
                Makanan Tidak Ditemukan
            </h3>

            <p class="max-w-sm text-sm leading-6 text-gray-500">
                Maaf, makanan yang kamu cari tidak tersedia.
                Coba gunakan kata kunci lain.
            </p>
        </div>



        <div v-else class="mt-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            <FoodCard v-for="food in foods" :key="food.id" :food="food" />
        </div>
    </section>
</template>
