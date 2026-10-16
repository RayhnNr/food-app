import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import FoodsView from '@/views/FoodsView.vue'
import KeranjangView from '@/views/KeranjangView.vue'
import FoodDetail from '@/views/FoodDetail.vue'
import LatihanViews from '@/views/LatihanViews.vue'
import AdminLayout from '@/layouts/AdminLayout.vue'
import DashboardView from '@/views/admin/DashboardView.vue'
import AdminFoodsView from '@/views/admin/foods/AdminFoodsView.vue'
import FoodFormView from '@/views/admin/foods/FoodFormView.vue'
import PesananFoodsView from '@/views/admin/pesanans/PesananFoodsView.vue'
import PesananEditView from '@/views/admin/pesanans/PesananEdit.View.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'Home', component: HomeView },
    { path: '/foods', name: 'Foods', component: FoodsView },
    { path: '/foods/:id', name: 'FoodDetail', component: FoodDetail },
    { path: '/latihan', name: 'Latihan', component: LatihanViews },
    { path: '/keranjangs', name: 'Keranjang', component: KeranjangView },
    {
      path: '/admin',
      component: AdminLayout,
      redirect: '/admin/dashboard',
      children: [
        { path: 'dashboard', component: DashboardView },
        { path: 'foods', component: AdminFoodsView },
        { path: 'foods/tambah', component: FoodFormView },
        { path: 'foods/:id/edit', component: FoodFormView },
        { path: 'pesanans', component: PesananFoodsView },
        { path: 'pesanans/:id/edit', component: PesananEditView },
      ],
    },
  ],
})

export default router