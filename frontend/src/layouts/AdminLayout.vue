<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuth } from '@/auth';

const sidebarOpen = ref(false);
const router = useRouter();
const { logout } = useAuth();

function handleLogout() {
    logout();
    router.replace('/login');
}

const menus = [
    { to: '/admin/dashboard', label: 'Dashboard', icon: 'fa-gauge' },
    { to: '/admin/foods', label: 'Kelola Menu', icon: 'fa-utensils' },
    { to: '/admin/pesanans', label: 'Pesanans', icon: 'fa-receipt' },
];
</script>

<template>
    <div class="min-h-screen bg-slate-100 lg:flex">
        <!-- Overlay (HP) -->
        <div
            v-if="sidebarOpen"
            class="fixed inset-0 z-30 bg-black/40 lg:hidden"
            @click="sidebarOpen = false"
        ></div>

        <!-- Sidebar (gaya Flowbite) -->
        <aside
            class="fixed inset-y-0 left-0 z-40 w-64 transform transition-transform lg:static lg:translate-x-0"
            :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
            aria-label="Sidebar"
        >
            <div
                class="flex h-full flex-col px-3 py-4 overflow-y-auto bg-gray-50 dark:bg-gray-800 lg:rounded-r"
            >
                <!-- Brand -->
                <div
                    class="flex items-center p-2 mb-4 text-lg font-bold text-gray-900 dark:text-white"
                >
                    <i class="fa-solid fa-utensils w-6 text-center text-emerald-600"></i>
                    <span class="ml-3">Admin Panel</span>
                </div>

                <!-- Menu -->
                <ul class="space-y-2">
                    <li v-for="m in menus" :key="m.to">
                        <RouterLink
                            :to="m.to"
                            class="group flex items-center p-2 text-base font-normal text-gray-900 rounded-lg dark:text-white hover:bg-gray-100 dark:hover:bg-gray-700"
                            active-class="!bg-gray-200 dark:!bg-gray-700"
                            @click="sidebarOpen = false"
                        >
                            <i
                                :class="[
                                    'fa-solid w-6 text-center text-lg text-gray-500 transition duration-75 dark:text-gray-400',
                                    'group-hover:text-gray-900 dark:group-hover:text-white',
                                    m.icon,
                                ]"
                            ></i>
                            <span class="flex-1 ml-3 whitespace-nowrap">{{ m.label }}</span>
                        </RouterLink>
                    </li>
                </ul>

                <!-- Bawah -->
                <ul class="pt-4 mt-auto space-y-2 border-t border-gray-200 dark:border-gray-700">
                    <li>
                        <RouterLink
                            to="/"
                            class="group flex items-center p-2 text-base font-normal text-gray-900 rounded-lg dark:text-white hover:bg-gray-100 dark:hover:bg-gray-700"
                        >
                            <i
                                class="fa-solid fa-arrow-left w-6 text-center text-lg text-gray-500 transition duration-75 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-white"
                            ></i>
                            <span class="flex-1 ml-3 whitespace-nowrap">Kembali ke Situs</span>
                        </RouterLink>
                    </li>
                </ul>
            </div>
        </aside>

        <!-- Konten -->
        <div class="flex-1 min-w-0">
            <!-- Navbar atas -->
            <header
                class="sticky top-0 z-20 h-16 bg-white shadow-sm flex items-center justify-between px-4"
            >
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
                    <div
                        class="w-9 h-9 rounded-full bg-emerald-600 text-white flex items-center justify-center"
                    >
                        <i class="fa-solid fa-user"></i>
                    </div>
                    <button
                        type="button"
                        class="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm font-medium text-slate-700 transition hover:bg-slate-50"
                        @click="handleLogout"
                    >
                        <i class="fa-solid fa-right-from-bracket"></i>
                        Logout
                    </button>
                </div>
            </header>

            <main>
                <RouterView />
            </main>
        </div>
    </div>
</template>
