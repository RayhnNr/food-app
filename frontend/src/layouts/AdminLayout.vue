<script setup>
import { ref } from 'vue'

const sidebarOpen = ref(false)

const menus = [
  { to: '/admin/dashboard', label: 'Dashboard', icon: 'fa-gauge' },
  { to: '/admin/foods', label: 'Kelola Menu', icon: 'fa-utensils' },
  { to: '/admin/pesanans', label: 'Pesanans', icon: 'fa-utensils' },
]
</script>

<template>
  <div class="min-h-screen bg-slate-100 lg:flex">
    <!-- Overlay (HP) -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-30 bg-black/40 lg:hidden"
      @click="sidebarOpen = false"
    ></div>

    <!-- Sidebar -->
    <aside
      class="fixed inset-y-0 left-0 z-40 w-64 bg-slate-900 text-slate-200 transform transition-transform lg:static lg:translate-x-0"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <div class="h-16 flex items-center gap-2 px-6 font-bold text-lg text-white border-b border-slate-700">
        <i class="fa-solid fa-utensils text-emerald-400"></i> Admin Panel
      </div>

      <nav class="p-4 space-y-1">
        <RouterLink
          v-for="m in menus"
          :key="m.to"
          :to="m.to"
          class="flex items-center gap-3 px-3 py-2 rounded-lg transition hover:bg-slate-800"
          active-class="bg-emerald-600 text-white"
          @click="sidebarOpen = false"
        >
          <i :class="['fa-solid w-5 text-center', m.icon]"></i>
          {{ m.label }}
        </RouterLink>
      </nav>

      <div class="absolute bottom-0 w-full p-4 border-t border-slate-700">
        <RouterLink to="/" class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-800">
          <i class="fa-solid fa-arrow-left w-5 text-center"></i> Kembali ke Situs
        </RouterLink>
      </div>
    </aside>

    <!-- Konten -->
    <div class="flex-1 min-w-0">
      <!-- Navbar atas -->
      <header class="sticky top-0 z-20 h-16 bg-white shadow-sm flex items-center justify-between px-4">
        <button
          class="lg:hidden p-2 rounded-lg hover:bg-slate-100"
          aria-label="Buka sidebar"
          @click="sidebarOpen = true"
        >
          <i class="fa-solid fa-bars text-xl"></i>
        </button>

        <h1 class="font-semibold text-slate-700 hidden lg:block">Dashboard Admin</h1>

        <div class="flex items-center gap-3">
          <span class="text-sm text-slate-600">Admin</span>
          <div class="w-9 h-9 rounded-full bg-emerald-600 text-white flex items-center justify-center">
            <i class="fa-solid fa-user"></i>
          </div>
        </div>
      </header>

      <main>
        <RouterView />
      </main>
    </div>
  </div>
</template>