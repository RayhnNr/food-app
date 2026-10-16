<script setup>
import { ref } from 'vue'
import api from '@/api'

defineProps({
    cartCount: { 
        type: Number, 
        default: 0 
    },
})

const menuOpen = ref(false)

const links = [
  { to: '/', label: 'Home', icon: 'fa-house' },
  { to: '/foods', label: 'Foods', icon: 'fa-utensils' },
  { to: '/latihan', label: 'Latihan', icon: 'fa-pencil' },
  { to: '/keranjangs', label: 'Keranjang', icon: 'fa-cart-shopping', cart: true },
  { to: '/admin', label: 'Admin', icon: 'fa-user-shield' },
  // { to: '/about', label: 'About', icon: 'fa-circle-info' },
]
</script>

<template>
  <nav class="sticky top-0 z-50 bg-emerald-600 text-white shadow-md">
    <div class="max-w-7xl mx-auto px-4 py-3 flex items-center justify-between">
      <!-- Logo -->
      <RouterLink to="/" class="flex items-center gap-2 font-bold text-lg">
        <i class="fa-solid fa-utensils"></i>
        BelajarVue
      </RouterLink>

      <!-- Menu desktop -->
      <div class="hidden md:flex items-center gap-2">
        <RouterLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="relative flex items-center gap-2 px-3 py-2 rounded-lg transition hover:bg-white/10"
          active-class="bg-white/20 font-semibold"
        >
          <i :class="['fa-solid', link.icon]"></i>
          {{ link.label }}
          <span
            v-if="link.cart && cartCount > 0"
            class="ml-1 min-w-5 h-5 px-1 rounded-full bg-amber-400 text-emerald-900 text-xs font-bold flex items-center justify-center"
          >
            {{ cartCount }}
          </span>
        </RouterLink>
      </div>

      <!-- Tombol hamburger (HP) -->
      <button
        class="md:hidden p-2 rounded-lg hover:bg-white/10"
        @click="menuOpen = !menuOpen"
        aria-label="Buka menu"
      >
        <i :class="['fa-solid', 'text-xl', menuOpen ? 'fa-xmark' : 'fa-bars']"></i>
      </button>
    </div>

    <!-- Menu mobile -->
    <div v-if="menuOpen" class="md:hidden border-t border-white/20 px-4 py-5 space-y-1">
      <RouterLink
        v-for="link in links"
        :key="link.to"
        :to="link.to"
        class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-white/10"
        active-class="bg-white/20 font-semibold"
        @click="menuOpen = false"
      >
        <i :class="['fa-solid', link.icon, 'w-5']"></i>
        {{ link.label }}
        <span
          v-if="link.cart && cartCount > 0"
          class="ml-auto min-w-5 h-5 px-1 rounded-full bg-amber-400 text-emerald-900 text-xs font-bold flex items-center justify-center"
        >
          {{ cartCount }}
        </span>
      </RouterLink>
    </div>
  </nav>
</template>