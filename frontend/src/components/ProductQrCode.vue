<script setup>
import QrcodeVue from 'qrcode.vue'

const props = defineProps({
    kodeProduk: {
        type: String,
        required: true,
    },
    size: {
        type: Number,
        default: 80,
    },
})
</script>

<template>
    <QrcodeVue
        :value="kodeProduk"
        :size="size"
        level="H"
    />
</template>