<script setup>
import { ref, onMounted } from 'vue';
import api from '@/api';
import { useRoute } from 'vue-router';
import Swal from 'sweetalert2';
import { useCart } from '@/composables/useCart';

const { getCartCount } = useCart();
const route = useRoute();

const food = ref(null);
const pesan = ref({
    jumlah: 1,
    keterangan: '',
});

// Toast kecil ala SweetAlert (pengganti vue-toast-notification)
const Toast = Swal.mixin({
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
});

onMounted(async () => {
    try {
        const res = await api.get(`/products/${route.params.id}`);
        food.value = res.data;
    } catch (e) {
        Swal.fire({
            title: 'Gagal',
            text: 'Gagal mengambil data produk',
            icon: 'error',
        });
        console.error(e);
    }
});

const pemesanan = async () => {
    if (!food.value) {
        Toast.fire({ icon: 'error', title: 'Data produk belum tersedia' });
        return;
    }

    if (Number(pesan.value.jumlah) < 1) {
        Toast.fire({ icon: 'error', title: 'Jumlah minimal 1' });
        return;
    }

    try {
        await api.post('/keranjangs', {
            product_id: food.value.id,
            jumlah: Number(pesan.value.jumlah),
            keterangan: pesan.value.keterangan.trim(),
        });

        Toast.fire({ icon: 'success', title: 'Berhasil ditambahkan ke keranjang' });

        await getCartCount();
    } catch (e) {
        Toast.fire({
            icon: 'error',
            title:
                e.response?.data?.message ||
                e.response?.data?.errors?.product_id?.[0] ||
                e.response?.data?.errors?.jumlah?.[0] ||
                'Gagal menambah ke keranjang',
        });
        console.error(e);
    }
};
</script>

<template>
    <section class="max-w-5xl mx-auto px-4 py-12">
        <p v-if="!food" class="text-emerald-600">Memuat data...</p>
        <div v-else class="grid md:grid-cols-2 gap-10 items-center">
            <!-- Gambar -->
            <img
                :src="`/gambar/${food.gambar}`"
                :alt="food.nama"
                class="w-full h-72 md:h-120 object-cover rounded-2xl shadow-md"
            />
            <div>
                <nav aria-label="breadcrumb" class="w-max">
                    <ol class="flex w-full flex-wrap items-center rounded-md py-1">
                        <li
                            class="flex cursor-pointer items-center text-sm text-emerald-600 transition-colors duration-300 hover:text-emerald-800"
                        >
                            <RouterLink to="/">
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    class="h-4 w-4"
                                    viewBox="0 0 20 20"
                                    fill="currentColor"
                                >
                                    <path
                                        d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z"
                                    ></path>
                                </svg>
                            </RouterLink>
                            <span class="pointer-events-none mx-2 text-emerald-800"> / </span>
                        </li>
                        <li
                            class="flex cursor-pointer items-center text-sm text-emerald-600 transition-colors duration-300 hover:text-emerald-800"
                        >
                            <!-- <a href="#">Components</a> -->
                            <RouterLink to="/foods"> Foods</RouterLink>
                            <span class="pointer-events-none mx-2 text-emerald-800"> / </span>
                        </li>
                        <li
                            class="flex cursor-pointer items-center text-sm text-emerald-600 transition-colors duration-300 hover:text-emerald-800"
                        >
                            <a href="">{{ food.nama }}</a>
                        </li>
                    </ol>
                </nav>

                <h1 class="mt-3 text-3xl md:text-4xl font-bold text-emerald-800">
                    {{ food.nama }}
                </h1>
                <p class="mt-2 text-2xl text-emerald-600 font-bold">
                    Rp {{ food.harga.toLocaleString('id-ID') }}
                </p>
                <p
                    class="mt-3 text-sm"
                    :class="food.is_ready ? 'text-emerald-600' : 'text-red-500'"
                >
                    <i
                        class="fa-solid"
                        :class="food.is_ready ? 'fa-circle-check' : 'fa-circle-xmark'"
                    ></i>
                    {{ food.is_ready ? 'Tersedia' : 'Habis' }}
                </p>

                <form v-on:submit.prevent class="mt-6 space-y-4">
                    <div>
                        <label for="jumlah" class="block text-sm font-medium text-slate-700 mb-1"
                            >Jumlah Pesanan</label
                        >
                        <input
                            id="jumlah"
                            v-model="pesan.jumlah"
                            type="number"
                            class="w-full px-4 py-2 rounded-xl border border-slate-300 focus:outline-none focus:ring-4 focus:ring-emerald-300"
                        />
                    </div>

                    <div>
                        <label
                            for="keterangan"
                            class="block text-sm font-medium text-slate-700 mb-1"
                            >Keterangan</label
                        >
                        <textarea
                            id="keterangan"
                            v-model="pesan.keterangan"
                            placeholder="Contoh: tidak pedas"
                            class="w-full px-4 py-2 rounded-xl border border-slate-300 focus:outline-none focus:ring-4 focus:ring-emerald-300"
                        ></textarea>
                    </div>
                </form>

                <button
                    :disabled="!food.is_ready"
                    @click="pemesanan"
                    class="mt-6 inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-emerald-600 text-white font-medium transition hover:bg-emerald-700 focus:outline-none focus:ring-4 focus:ring-emerald-300 disabled:bg-slate-300 disabled:cursor-not-allowed"
                >
                    <i class="fa-solid fa-cart-plus"></i> Tambah ke Keranjang
                </button>
            </div>
        </div>
    </section>
</template>
