<script setup>
import { ref, onMounted } from 'vue';
import api from '@/api';
import FoodCard from '@/components/FoodCard.vue';
import ilustrasi from '@/assets/gambar/ilustrasi.png';

const foods = ref([]);
const loading = ref(true);
const error = ref(null);


const getFoods = async () => {
    try {
        const res = await api.get('/products', { params: { _limit: 6 } });
        console.log('Data API: ', res.data);
        foods.value = res.data.data;
    } catch (e) {
        error.value = 'Gagal mengambil data. Pastikan npm run api sudah jalan.';
    } finally {
        loading.value = false;
    }
};

onMounted(getFoods);
</script>

<template>
    <!-- Hero -->
    <section class="max-w-7xl mx-auto px-4 py-12 md:py-16 grid md:grid-cols-2 gap-10 items-center">
        <div>
            <h1 class="text-4xl md:text-5xl font-extrabold text-slate-800 leading-tight">
                Lapar? Pilih Makananmu,
                <span class="text-emerald-600">Kami Siapkan.</span>
            </h1>
            <p class="my-4 text-lg text-slate-600">
                Temukan berbagai menu pilihan, masukkan ke keranjang, dan pesan dalam hitungan
                detik.
            </p>
            <RouterLink
                to="/foods"
                class="inline-flex items-center gap-2 px-5 py-3 rounded-xl bg-emerald-600 text-white text-sm font-medium shadow-sm transition hover:bg-emerald-700 focus:outline-none focus:ring-4 focus:ring-emerald-300"
            >
                Pilihan Menu <i class="fa-solid fa-arrow-right"></i>
            </RouterLink>
        </div>
        <div>
            <img :src="ilustrasi" alt="Aneka makanan" class="w-full h-72 md:h-96 object-contain" />
        </div>
    </section>

    <section class="max-w-7xl mx-auto px-4 py-12 md:py-16">
        <div class="flex items-end justify-between gap-4">
            <div>
                <h2 class="text-2xl font-bold text-slate-800">Favorit Pelanggan</h2>
                <p class="mt-1 text-sm text-slate-600">
                    Menu yang paling sering dipesan minggu ini.
                </p>
            </div>
            <RouterLink
                to="/foods"
                class="text-emerald-600 font-semibold hover:underline whitespace-nowrap"
            >
                Lihat semua <i class="fa-solid fa-arrow-right ml-1"></i>
            </RouterLink>
        </div>

        <p v-if="loading" class="mt-6 text-slate-500">Memuat data...</p>
        <p v-else-if="error" class="mt-6 text-red-600">{{ error }}</p>

        <div v-else class="mt-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            <FoodCard
                v-for="food in foods"
                :key="food.id"
                :food="food"
                badge="Terlaris"
            />
        </div>
    </section>
</template>
