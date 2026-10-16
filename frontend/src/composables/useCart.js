import { ref } from 'vue'
import api from '@/api'

const cartCount = ref(0)


export function useCart(){
    const getCartCount = async () => {
        try {
            const res = await api.get('/keranjangs')
            cartCount.value = res.data.length
            console.log('cartCount:', cartCount.value)
        } catch (e) {
            console.error(e)
        }
    }

    return {cartCount, getCartCount}
}
