<script setup>
// import { ref, onMounted, reactive } from 'vue';
import { ref, onMounted, onUpdated, onUnmounted, reactive, computed } from 'vue'
import api from '@/api';
import FoodCard from '@/components/FoodCard.vue';
import ilustrasi from '@/assets/gambar/ilustrasi.png';

const foods = ref([]);
const loading = ref(true);
const error = ref(null);

// Test Reff templates
const inputEl = ref(null)
const box = ref(null)

const angka = ref(0)

onMounted(() => {
  console.log('2. onMounted: komponen sudah TAMPIL')
  inputEl.value.focus()
  console.log("Lebar box: ", box.value.offsetWidth, "px")
})

onUpdated(() => {
  console.log('3. onUpdated: tampilan DIPERBARUI (angka sekarang ' + angka.value + ')')
})

onUnmounted(() => {
  console.log('4. onUnmounted: komponen DIHAPUS (pindah halaman)')
})

const isVisible = ref(true)
const buahs = ref([
    'mangga',
    'apel',
    'jeruk',
    'alpukat'
])

const user = reactive({
  name: "Budi",
  age: 20,
  city: "Jakarta"
})

const users = ref([
  { id: 1, name: "Budi", age: 20 },
  { id: 2, name: "Siti", age: 22 },
  { id: 3, name: "Andi", age: 25 }
])

const categories = ref([
  { name: "Buah", items: ["Apel", "Mangga", "Jeruk"] },
  { name: "Sayur", items: ["Bayam", "Wortel", "Tomat"] }
])

const name = ref("");
const message = ref("");
const gender = ref("");
const agree = ref(true);
const city = ref("");

// Computed properties
const harga = ref(10000)
const jumlah = ref(2)
const age = ref(20)

const total = computed(() => harga.value * jumlah.value)
const status = computed(() => {
    return age.value >= 18 ? "Dewasa" : "Anak-Anak"
})


// const getFoods = async () => {
//     try {
//         const res = await api.get('/best-products', { params: { _limit: 6 } });
//         console.log('Data API: ', res.data);
//         foods.value = res.data;
//     } catch (e) {
//         error.value = 'Gagal mengambil data. Pastikan npm run api sudah jalan.';
//     } finally {
//         loading.value = false;
//     }
// };

// onMounted(getFoods);
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
    <section class="max-w-7xl mx-auto px-4 py-12 md:py-16 grid md:grid-cols-2 gap-10 items-center">
        <div>
            <h1 class="text-xl font-bold py-2" v-show="isVisible">Teks ini bisa disembunyikan dengan v-show</h1>
            <button @click="isVisible = !isVisible" class="text-white bg-blue-600 px-7 py-3 rounded-xl cursor-pointer hover:bg-blue-800">
                Toogle
            </button>

            <h1 class="text-xl font-bold py-2">Perulangan dengan v-for</h1>
            <ul>
                <li v-for="buah in buahs" :key="buah">{{ buah }}</li>
            </ul>

            <h1 class="text-xl font-bold py-2">Perulangan dengan v-for pake index</h1>
            <ul>
                <li v-for="(buah, index) in buahs" :key="index">{{ index+1 }}. {{ buah }}</li>
            </ul>

            <h1 class="text-xl font-bold py-2">Render Object</h1>
            <ul>
                <li v-for="(value, key) in user" :key="key">
                    {{ key }}: {{ value }}
                </li>
            </ul>

            <h1 class="text-xl font-bold py-2">Render Daftar Object</h1>
            <div v-for="user in users" :key="user.id" class="card">
                <h3>{{ user.name }}</h3>
                <p>Umur: {{ user.age }}</p>
            </div>

            <h1 class="text-xl font-bold py-2">Nested Loop(Loop Didalam Loop)</h1>
            <div v-for="category in categories" :key="category.name">
                <h4>{{ category.name }}</h4>
                <ul>
                    <li v-for="item in category.items" :key="item">{{ item }}</li>
                </ul>
            </div>

            <h1 class="text-xl font-bold py-2">V-model dalam inputan</h1>
            <input v-model="name" placeholder="Ketika nama">
            <p>Real time Haisl ketikan: {{ name }}</p>

            <h1 class="text-xl font-bold py-2">V-model dalam textarea</h1>
            <textarea v-model="message" placeholder="Ketika nama"></textarea>
            <p>Real time Haisl ketikan: {{ message }}</p>

            <h1 class="text-xl font-bold py-2">V-model dalam Checkbox</h1>
            <label for="">
                <input v-model="agree" type="checkbox" name="" id=""> Saya Setuju
            </label>
            <p>Status: {{ agree }}</p>

            <h1 class="text-xl font-bold py-2">V-model dalam Radio</h1>
            <div>
                <label>
                    <input type="radio" value="Laki-laki" v-model="gender" /> Laki-laki
                </label>
                <label>
                    <input type="radio" value="Perempuan" v-model="gender" /> Perempuan
                </label>
                <p>Jenis Kelamin: {{ gender }}</p>
            </div>

            <h1 class="text-xl font-bold py-2">V-model dalam Radio</h1>
            <select v-model="city">
                <option disabled value="">-- Pilih Kota --</option>
                <option>Jakarta</option>
                <option>Bandung</option>
                <option>Surabaya</option>
            </select>

            <p>Kota dipilih: {{ city }}</p>

        </div>
    </section>

    <section class="max-w-7xl mx-auto px-4 py-12">
        <h1 class="text-2xl font-bold">Belajar Lifecycle</h1>
        <p class="mt-4 text-xl">Angka: {{ angka }}</p>

        <button @click="angka++" class="mt-4 px-4 py-2 rounded-xl bg-emerald-600 text-white hover:bg-emerald-700">
            Tambah angka
        </button>
        <p class="mt-6 text-sm text-slate-500">Buka F12, tab Console, lalu perhatikan teksnya.</p>
    </section>

    <section class="max-w-7xl mx-auto px-4 py-12">
        <h1 class="text-2xl font-bold">Templates Reff</h1>
        <input ref="inputEl" placeholder="Auto focus saat mount" />
        <!-- ketika ada duplicar reff auto focus ini maka yang dijalankan hanya 1 saja -->
        <!-- <input ref="inputEl" placeholder="Auto focus saat mount" />  -->
        <br>
        <div class="mt-12" ref="box" style="width: 200px; height: 100px; background: lightblue;">
            Kotak
        </div>
    </section>

    <section class="max-w-7xl mx-auto px-4 py-12">
        <h1 class="text-2xl font-bold">Computed dan Watchers</h1>
        <p class="mt-6 text-lg text-black">Hasil Perhitungan: {{ total }}</p>
        <input v-model.number="jumlah" type="number">
        <p class="mt-6 text-lg text-black">Age: {{ age }}</p>
        <p class="text-lg text-black">Status: {{ status }}</p>
        <input v-model.number="age" type="number">
    </section>

</template>
