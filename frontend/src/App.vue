<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRoute } from 'vue-router'
import Navbar from '@/components/Navbar.vue';
import api from '@/api';
import { useCart } from './composables/useCart';

const route = useRoute()
const isAdmin = computed(() => route.path.startsWith('/admin'))
const { cartCount, getCartCount } = useCart();

onMounted(getCartCount);
</script>

<template>
  <div class="min-h-screen bg-slate-100">
    <Navbar v-if="!isAdmin" :cart-count="cartCount" />
    <RouterView />
  </div>
</template>
