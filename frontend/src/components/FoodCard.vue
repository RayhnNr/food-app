<script setup>
defineProps({
  food: { type: Object, required: true },
  badge: { type: String, default: '' },
  showButton: { type: Boolean, default: true },
})

const emit = defineEmits(['add'])

const rupiah = (n) => 'Rp ' + Number(n).toLocaleString('id-ID')
</script>

<template>
    <div class="relative bg-white rounded-2xl p-4 shadow-sm transition hover:shadow-md hover:-translate-y-1" >
        <span v-if="badge" class="absolute top-6 left-6 z-10 px-2 py-0.5 rounded-full bg-amber-400 text-emerald-900 text-xs font-bold">
            <i class="fa-solid fa-fire mr-1"></i>{{ badge }}
        </span>

        <div class="h-60 rounded-xl bg-amber-50 overflow-hidden flex items-center justify-center text-5xl">
            <img
            v-if="food.gambar"
            :src="`/gambar/${food.gambar}`"
            :alt="food.nama"
            class="w-full h-full object-cover" 
            />
            <span v-else>{{ food.emoji }}</span>
        </div>

        <h3 class="mt-3 font-semibold text-slate-800 text-sm md:text-base line-clamp-1">
            {{ food.nama }}
        </h3>
        <p class="text-emerald-600 font-bold">{{ rupiah(food.harga) }}</p>

        <RouterLink v-if="showButton" :to="`/foods/${food.id}`" class="mt-3 block w-full px-4 py-2 rounded-xl bg-emerald-600 text-white text-sm font-medium text-center transition hover:bg-emerald-700 focus:outline-none focus:ring-4 focus:ring-emerald-300">
            <i class="fa-solid fa-eye mr-1"></i> Lihat Detail
        </RouterLink>
    </div>
</template>