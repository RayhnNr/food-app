<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Swal from 'sweetalert2'
import api from '@/api'

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => !!route.params.id)

const form = ref({
    // kode: '',
    nama: '',
    harga: 0,
    is_ready: true,
    gambar: '',
})

const errors = ref({})

const formatHarga = (value) => {
    const angka = value.replace(/\D/g, '')

    form.value.harga = angka === '' ? '' : Number(angka)

    return angka.replace(/\B(?=(\d{3})+(?!\d))/g, '.')
}

const handleHarga = (event) => {
    event.target.value = formatHarga(event.target.value)
}

onMounted(async () => {
    if (isEdit.value) {
        try {
            const res = await api.get(`/products/${route.params.id}`)
            form.value = res.data
        } catch (e) {
            Swal.fire({
                icon: 'error',
                title: 'Gagal',
                text: 'Gagal mengambil data menu',
            })
            console.error(e)
        }
    }
})

const validasi = () => {
    errors.value = {}

    // if (!form.value.kode) errors.value.kode = 'Kode wajib diisi'

    if (!form.value.nama) {
        errors.value.nama = 'Nama wajib diisi'
    } else if (form.value.nama.length < 3) {
        errors.value.nama = 'Nama minimal 3 karakter'
    }

    if (!form.value.harga || form.value.harga <= 0) {
        errors.value.harga = 'Harga harus lebih dari 0'
    }

    // true kalau tidak ada error
    return Object.keys(errors.value).length === 0
}

const simpan = async () => {
    if (!validasi()) return

    try {
        if (isEdit.value) {
            await api.put(`/products/${route.params.id}`, form.value)
        } else {
            await api.post('/products', form.value)
        }

        await Swal.fire({
            icon: 'success',
            title: 'Berhasil',
            text: isEdit.value ? 'Berhasil diperbarui' : 'Berhasil ditambahkan',
            timer: 1500,
            showConfirmButton: false,
        })

        router.push('/admin/foods')
    } catch (e) {
        Swal.fire({
            icon: 'error',
            title: 'Gagal',
            text: e.response?.data?.message || 'Gagal menyimpan',
        })
        console.error(e)
    }
}

// Hapus pesan error ketika pengguna mulai memperbaiki isian
watch(
    form,
    () => {
        errors.value = {}
    },
    { deep: true },
)
</script>

<template>
    <section class="max-w-xl mx-auto px-4 py-12">
        <h1 class="text-2xl font-bold text-slate-800 mb-6">
            {{ isEdit ? 'Edit Menu' : 'Tambah Menu' }}
        </h1>

        <form @submit.prevent="simpan" novalidate class="space-y-4">
            <!-- Kode -->
            
            <div  v-if="isEdit">
                <label class="block text-sm font-medium mb-1">Kode</label>
                <input
                    v-model.trim="form.kode"
                    class="w-full px-4 py-2 rounded-xl border focus:outline-none focus:ring-4"
                    :class="
                        errors.kode
                            ? 'border-red-500 focus:ring-red-200'
                            : 'border-slate-300 focus:ring-emerald-300'
                    "
                    readonly
                />
                <!-- <p v-if="errors.kode" class="mt-1 text-sm text-red-600">{{ errors.kode }}</p> -->
            </div>

            <!-- Nama -->
            <div>
                <label class="block text-sm font-medium mb-1">Nama</label>
                <input
                    v-model.trim="form.nama"
                    class="w-full px-4 py-2 rounded-xl border focus:outline-none focus:ring-4"
                    :class="
                        errors.nama
                            ? 'border-red-500 focus:ring-red-200'
                            : 'border-slate-300 focus:ring-emerald-300'
                    "
                />
                <p v-if="errors.nama" class="mt-1 text-sm text-red-600">{{ errors.nama }}</p>
            </div>

            <!-- Harga -->
            <div>
                <label class="block text-sm font-medium mb-1">Harga</label>       
                <input
                    :value="form.harga === '' ? '' : Number(form.harga).toLocaleString('id-ID')"
                    @input="handleHarga"
                    type="text"
                    inputmode="numeric"
                    placeholder="Contoh: 1.500.000"
                    class="w-full px-4 py-2 rounded-xl border focus:outline-none focus:ring-4"
                    :class="
                        errors.harga
                            ? 'border-red-500 focus:ring-red-200'
                            : 'border-slate-300 focus:ring-emerald-300'
                    "
                />
                <p v-if="errors.harga" class="mt-1 text-sm text-red-600">{{ errors.harga }}</p>
            </div>

            <!-- Gambar -->
            <div>
                <label class="block text-sm font-medium mb-1">Nama file gambar</label>
                <input
                    v-model.trim="form.gambar"
                    placeholder="contoh: sate-ayam.jpg"
                    class="w-full px-4 py-2 rounded-xl border border-slate-300 focus:outline-none focus:ring-4 focus:ring-emerald-300"
                />
            </div>

            <!-- Tersedia -->
            <label class="relative inline-flex items-center cursor-pointer text-gray-900 gap-3">
                <input type="checkbox" v-model="form.is_ready" class="sr-only peer" checked />
                <div class="w-16 h-8 bg-slate-300 rounded-full peer peer-checked:bg-indigo-600 transition-colors duration-200"></div>
                <span class="dot absolute left-1 top-1 w-6 h-6 bg-white rounded-full transition-transform duration-200 ease-in-out peer-checked:translate-x-8"></span>
                Tersedia
            </label>
            <!-- <label class="flex items-center gap-2">
                <input v-model="form.is_ready" type="checkbox" /> Tersedia
            </label> -->

            <!-- Tombol -->
            <div class="flex gap-3">
                <button
                    type="submit"
                    class="px-5 py-2.5 rounded-xl bg-emerald-600 text-white hover:bg-emerald-700"
                >
                    Simpan
                </button>
                <RouterLink
                    to="/admin/foods"
                    class="px-5 py-2.5 rounded-xl border border-slate-300"
                >
                    Batal
                </RouterLink>
            </div>
        </form>
    </section>
</template>