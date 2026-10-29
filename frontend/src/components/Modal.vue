<script setup>
defineProps({
    show: { type: Boolean, default: false },
    title: { type: String, default: '' },
    // Lebar modal. Default max-w-lg, jadi modal lama tidak berubah.
    maxWidth: { type: String, default: 'max-w-lg' },
});

const emit = defineEmits(['close']);
</script>

<template>
    <Teleport to="body">
        <div
            v-if="show"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
            @click.self="emit('close')"
            @keyup.esc="emit('close')"
        >
            <div :class="['w-full bg-white rounded-2xl shadow-xl p-6', maxWidth]">
                <div class="flex items-center justify-between mb-4">
                    <h3 class="text-lg font-bold text-slate-800">{{ title }}</h3>
                    <button
                        @click="emit('close')"
                        aria-label="Tutup"
                        class="text-slate-400 hover:text-slate-600"
                    >
                        <i class="fa-solid fa-xmark"></i>
                    </button>
                </div>

                <!-- isi modal dikirim dari luar -->
                <slot />
            </div>
        </div>
    </Teleport>
</template>
