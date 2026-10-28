import { createRouter, createWebHistory } from 'vue-router';
import { useAuth } from '@/auth';
import HomeView from '@/views/HomeView.vue';
import FoodsView from '@/views/FoodsView.vue';
import KeranjangView from '@/views/KeranjangView.vue';
import FoodDetail from '@/views/FoodDetail.vue';
import LatihanViews from '@/views/LatihanViews.vue';
// import LoginView from '@/views/auth/Login.vue'        // BARU
// import RegisterView from '@/views/auth/RegisterView.vue'    // BARU
import AdminLayout from '@/layouts/AdminLayout.vue';
import DashboardView from '@/views/admin/DashboardView.vue';
import AdminFoodsView from '@/views/admin/foods/AdminFoodsView.vue';
import FoodFormView from '@/views/admin/foods/FoodFormView.vue';
import PesananFoodsView from '@/views/admin/pesanans/PesananFoodsView.vue';
import PesananEditView from '@/views/admin/pesanans/PesananEdit.View.vue';
import LoginView from '@/views/auth/Login.vue';
import RegisterView from '@/views/auth/Register.vue';
import CategoryView from '@/views/admin/category/CategoryView.vue';
import CoaView from '@/views/admin/coa/CoaView.vue';
import CoaJournalView from '@/views/admin/coa_journal/CoaJournalView.vue';
import CoaType from '@/views/admin/coa_type/CoaType.vue';
import GoodsPartai from '@/views/admin/goods_partai/GoodsPartai.vue';
import FormGoodsPartai from '@/views/admin/goods_partai/FormGoodsPartai.vue';
import PartaiTarraDecimal from '@/views/admin/goods_partai/PartaiTarraDecimal.vue';

const router = createRouter({
    history: createWebHistory(),
    routes: [
        { path: '/', name: 'Home', component: HomeView },
        { path: '/foods', name: 'Foods', component: FoodsView },
        { path: '/foods/:id', name: 'FoodDetail', component: FoodDetail },
        { path: '/latihan', name: 'Latihan', component: LatihanViews },

        // Halaman login dapat dibuka meski sesi sebelumnya masih aktif
        { path: '/login', name: 'Login', component: LoginView },
        { path: '/register', name: 'Register', component: RegisterView, meta: { guestOnly: true } },

        // Wajib login
        {
            path: '/keranjangs',
            name: 'Keranjang',
            component: KeranjangView,
            meta: { requiresAuth: true },
        },

        // Wajib login + admin
        {
            path: '/admin',
            component: AdminLayout,
            redirect: '/admin/dashboard',
            meta: { requiresAuth: true, requiresAdmin: true },
            children: [
                { path: 'dashboard', component: DashboardView },
                { path: 'foods', component: AdminFoodsView },
                { path: 'foods/tambah', component: FoodFormView },
                { path: 'foods/:id/edit', component: FoodFormView },
                { path: 'category', component: CategoryView },
                { path: 'coa', component: CoaView },
                { path: 'coa-types', component: CoaType },
                { path: 'coa-journal', component: CoaJournalView },
                { path: 'goods-partai', component: GoodsPartai },
                { path: 'goods-partai/tambah', component: FormGoodsPartai },
                { path: 'goods-partai/:id/edit', component: FormGoodsPartai },
                { path: 'goods-partai/:id/tarra-decimal', component: PartaiTarraDecimal, meta: { requiresAuth: true } },
                { path: 'pesanans', component: PesananFoodsView },
                { path: 'pesanans/:id/edit', component: PesananEditView },
            ],
        },
    ],
});

// BARU: guard navigasi
router.beforeEach((to) => {
    const { isLoggedIn, user } = useAuth();

    // to.matched ikut memeriksa meta milik route induk (/admin -> semua child)
    const perluLogin = to.matched.some((r) => r.meta.requiresAuth);
    const perluAdmin = to.matched.some((r) => r.meta.requiresAdmin);
    const khususTamu = to.matched.some((r) => r.meta.guestOnly);

    if (perluLogin && !isLoggedIn.value) {
        // simpan tujuan awal supaya setelah login bisa kembali ke sana
        return { path: '/login', query: { redirect: to.fullPath } };
    }

    if (perluAdmin && user.value?.role !== 'admin') {
        return '/';
    }

    if (khususTamu && isLoggedIn.value) {
        return user.value?.role === 'admin' ? '/admin/dashboard' : '/';
    }
});

export default router;
