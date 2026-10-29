<script setup>
import { ref, watch, computed } from 'vue';
import { useRoute } from 'vue-router';
import Navbar from '@/components/Navbar.vue';
import { useCart } from './composables/useCart';
import { useAuth } from './auth';

const route = useRoute();
const isAdmin = computed(() => route.path.startsWith('/admin'));
const { cartCount, getCartCount } = useCart();
const { isLoggedIn } = useAuth();

watch(
    isLoggedIn,
    (loggedIn) => {
        if (loggedIn) {
            getCartCount();
        } else {
            cartCount.value = 0;
        }
    },
    { immediate: true },
);
</script>

<template>
    <div class="min-h-screen bg-slate-100">
        <Navbar v-if="!isAdmin && !route.meta.tanpaNavbar" :cart-count="cartCount" />
        
        <RouterView />
    </div>
</template>
