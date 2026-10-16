<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'vue-toast-notification'
import { useCart } from '@/composables/useCart'

const foods = ref([])
const loading = ref(true)
const error = ref(null)
const search = ref('')
const toast = useToast()
const { getCartCount } = useCart()

const getFoods = async () => {
    loading.value = true
    error.value = null

    try {
        const res = await api.get('/products', {
            params: { 
                q: search.value ,
                _sort: 'id',
                _order: 'DESC'
            }
        })
        foods.value = res.data.data
    } catch (e) {
        error.value = 'Gagal mengambil data. Pastikan npm run api sudah jalan.';
    } finally {
        loading.value = false;
    }
}

const hapusMakanan = async (id) => {
    if(!confirm('Apakah anda yakin ingin mengahapus Makanan ini?')){
        return;
    }

    try {
        await api.delete(`/products/${id}`);
        await getCartCount()
        foods.value = foods.value.filter(
            (item) => item.id !== id
        );
        toast.success('Berhasil Mengahapus Makanan')
    } catch (e) {
        toast.error('Gagal menghapus Makanan')
    }
}

function formatRupiah(angka) {
    return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
    }).format(angka);
}

onMounted(getFoods)

let timer;
watch(search, () => {
    clearTimeout(timer);
    timer = setTimeout(getFoods, 400);
});

</script>

<template>
  <section class="p-6">
    <h1 class="text-3xl font-bold mb-8"> Kelola Menu</h1>

        <!-- Search and Add User (Static) -->
        <div class="flex flex-col md:flex-row justify-between items-center mb-6">
            <div class="w-full md:w-1/3 mb-4 md:mb-0">
                <input v-model="search" type="text" placeholder="Search users..." class="w-full px-4 py-2 rounded-md border border-gray-300 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500">
            </div>
            <RouterLink to="/admin/foods/tambah" class="px-4 py-2 rounded-xl object-cover bg-emerald-600 text-white hover:bg-emerald-700">
                <i class="fa-solid fa-plus mr-1"></i> Tambah
            </RouterLink>
           
        </div>

        <!-- User Table -->
        <div class="overflow-x-auto bg-white rounded-lg shadow">
            <table class="w-full table-auto">
                <thead>
                    <tr class="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                        <th class="py-3 px-6 text-left">ID</th>
                        <th class="py-3 px-2 text-left">Kode Product</th>
                        <th class="py-3 px-6 text-left">Makanan</th>
                        <th class="py-3 px-6 text-left">Harga</th>
                        <th class="py-3 px-6 text-center">Actions</th>
                    </tr>
                </thead>
                <tbody class="text-gray-600 text-sm">
                    <tr v-if="foods.length === 0">
                        <td colspan="7" class="text-center py-10 text-gray-500">
                            Makanan masih kosong.
                        </td>
                    </tr>
                    <tr v-else v-for="(food, index) in foods" :key="food.id" class="border-b border-gray-200 hover:bg-gray-100">
                        <td class="py-3 px-6 text-left">
                            {{ index + 1 }}
                        </td>
                        <td class="py-3 px-2 text-left">{{ food.kode }}</td>
                        <td class="py-3 px-6 text-left">
                            <div class="flex items-center">
                                <img
                                    class="h-16 w-16 mr-4"
                                    :src="`/gambar/${food.gambar}`"
                                    :alt="food.nama"
                                />
                                <span class="font-semibold">{{ food.nama }}</span>
                            </div>
                        </td>
                        <td class="py-3 px-6 text-left">{{ formatRupiah(food.harga) }}</td>
                        <td class="py-3 px-6 text-center">
                            <div class="flex item-center justify-center">
                                <RouterLink :to="`/admin/foods/${food.id}/edit`" class="w-4 mr-2 transform hover:text-blue-500 hover:scale-110">
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                                    </svg>
                                </RouterLink>
                                <button @click="hapusMakanan(food.id)" class="w-4 mr-2 transform hover:text-red-500 hover:scale-110">
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                    </svg>
                                </button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
  </section>
</template>