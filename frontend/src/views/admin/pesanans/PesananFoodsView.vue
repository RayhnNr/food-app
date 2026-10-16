<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '@/api'
import { useToast } from 'vue-toast-notification'
import { useCart } from '@/composables/useCart'
import Modal from '@/components/Modal.vue'

const { getCartCount } = useCart()
const pesanans = ref([])
const loading = ref(true)
const error = ref(null)
const search = ref('')
const toast = useToast()
const showModal = ref(false)
const pesananDipilih = ref(null)

const bukaDetail = (pesanan) => {
  pesananDipilih.value = pesanan   // kirim data ke modal
  showModal.value = true
}

const tutup = () => {
  showModal.value = false
  pesananDipilih.value = null
}

const getPesanan = async () => {
    loading.value = true
    error.value = null

    try {
        const res = await api.get('/pesanans', {
            params: { 
                q: search.value ,
                _sort: 'id',
                _order: 'DESC'
            }
        })
        pesanans.value = res.data
    } catch (e) {
        error.value = 'Gagal mengambil data. Pastikan npm run api sudah jalan.';
    } finally {
        loading.value = false;
    }
}

const hapusMakanan = async (id) => {
    if(!confirm('Apakah anda yakin ingin mengahapus Pesanan ini?')){
        return;
    }

    try {
        await api.delete(`/pesanans/${id}`);
        await getCartCount()
        pesanans.value = pesanans.value.filter(
            (item) => item.id !== id
        );
        toast.success('Berhasil Mengahapus Makanan')
    } catch (e) {
        toast.error('Gagal menghapus Makanan')
    }
}

const ubahStatus = async (pesanan) => {
    try {
        await api.patch(`/pesanans/${pesanan.id}`, { status: 'selesai' })
        toast.success('Berhasil Mengahapus Makanan')
        await getPesanans()
    } catch (error) {
        toast.error('Gagal memperbarui status')
        console.error(e)
    }
    
}

function formatRupiah(angka) {
    return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
    }).format(angka);
}

const formatTanggal = (iso) => {
    if (!iso) return '-'
    return new Date(iso).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
        timeZone: 'Asia/Jakarta',
    })
}
const warnaStatus = (status) =>
    status === 'selesai' ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'

onMounted(getPesanan)

let timer;
watch(search, () => {
    clearTimeout(timer);
    timer = setTimeout(getpesanans, 400);
});

</script>

<template>
  <section class="p-6">
    <h1 class="text-3xl font-bold mb-8">Daftar Pesanan</h1>

        <!-- Search and Add User (Static) -->
        <div class="flex flex-col md:flex-row justify-between items-center mb-6">
            <div class="w-full md:w-1/3 mb-4 md:mb-0">
                <input v-model="search" type="text" placeholder="Search users..." class="w-full px-4 py-2 rounded-md border border-gray-300 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500">
            </div>
            <RouterLink to="/admin/pesanans/tambah" class="px-4 py-2 rounded-xl bg-emerald-600 text-white hover:bg-emerald-700">
                <i class="fa-solid fa-plus mr-1"></i> Tambah
            </RouterLink>
           
        </div>

        <!-- User Table -->
        <div class="overflow-x-auto bg-white rounded-lg shadow">
            <table class="w-full table-auto">
                <thead>
                    <tr class="bg-gray-200 text-gray-600 uppercase text-sm leading-normal">
                        <th class="py-3 px-6 text-left">ID</th>
                        <th class="py-3 px-2 text-left">Customer</th>
                        <th class="py-3 px-6 text-left">No. Meja</th>
                        <th class="py-3 px-6 text-left">Tanggal</th>
                        <th class="py-3 px-6 text-left">Total</th>
                        <th class="py-3 px-6 text-left">Status</th>
                        <th class="py-3 px-6 text-center">Actions</th>
                    </tr>
                </thead>
                <tbody class="text-gray-600 text-sm">
                    <tr v-if="pesanans.length === 0">
                        <td colspan="7" class="text-center py-10 text-gray-500">
                            Pesanan masih kosong.
                        </td>
                    </tr>
                    <tr v-else v-for="(pesanan, index) in pesanans" :key="pesanan.id" class="border-b border-gray-200 hover:bg-gray-100">
                        <td class="py-3 px-6 text-left">
                            {{ index + 1 }}
                        </td>
                        <td class="py-3 px-2 text-left">{{ pesanan.nama }}</td>
                        <td class="py-3 px-2 text-left">{{ pesanan.noMeja }}</td>
                        <td class="py-3 px-2 text-left">{{ formatTanggal(pesanan.tanggal) }}</td>
                        <td class="py-3 px-6 text-left">{{ formatRupiah(pesanan.total) }}</td>
                        <td class="py-3 px-6 text-left">
                            <span
                                class="px-3 py-1 rounded-full text-xs font-semibold capitalize"
                                :class="warnaStatus(pesanan.status)"
                            >
                                {{ pesanan.status }}
                            </span>
                        </td>
                        <td class="py-3 px-6 text-center">
                            <div class="flex item-center justify-center">
                                <RouterLink :to="`/admin/pesanans/${pesanan.id}/edit`" class="w-4 mr-2 transform hover:text-blue-500 hover:scale-110">
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                                    </svg>
                                </RouterLink>
                                <button
                                    v-if="pesanan.status !== 'selesai'"
                                    @click="ubahStatus(pesanan)"
                                    title="Tandai selesai"
                                    class="mr-2 text-emerald-600 hover:text-emerald-800 transition"
                                >
                                    <i class="fa-solid fa-check"></i>
                                </button>
                                <button @click="bukaDetail(pesanan)" title="Detail" class="text-sky-600 hover:underline">
                                    <i class="fa-solid fa-eye"></i>
                                </button>

                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
  </section>
  <Modal :show="showModal" title="Detail Pesanan" @close="tutup">
    <div v-if="pesananDipilih">
      <p class="font-semibold">{{ pesananDipilih.nama }} - Meja {{ pesananDipilih.noMeja }}</p>

      <ul class="mt-3 space-y-1 text-sm">
        <li v-for="item in pesananDipilih.keranjangs" :key="item.id">
          {{ item.products.nama }} x{{ item.jumlah }}
        </li>
      </ul>

      <p class="mt-4 font-bold text-emerald-600">
        Total: Rp {{ pesananDipilih.total?.toLocaleString('id-ID') }}
      </p>
    </div>
  </Modal>
</template>