<script setup>
import { computed } from 'vue';

const props = defineProps({
    currentPage: { type: Number, required: true },
    totalPages: { type: Number, required: true },
    totalItems: { type: Number, required: true },
    pageSize: { type: Number, required: true },
    loading: { type: Boolean, default: false },
});

const emit = defineEmits(['change']);

const pageNumbers = computed(() => {
    const total = props.totalPages;
    if (total <= 5) return Array.from({ length: total }, (_, index) => index + 1);
    if (props.currentPage <= 3) return [1, 2, 3, '...', total];
    if (props.currentPage >= total - 2) return [1, '...', total - 2, total - 1, total];
    return [
        1,
        '...',
        props.currentPage - 1,
        props.currentPage,
        props.currentPage + 1,
        '...',
        total,
    ];
});

const mulaiData = computed(() =>
    props.totalItems === 0 ? 0 : (props.currentPage - 1) * props.pageSize + 1,
);
const akhirData = computed(() => Math.min(props.currentPage * props.pageSize, props.totalItems));

const gantiHalaman = (page) => {
    if (props.loading || page < 1 || page > props.totalPages || page === props.currentPage) return;
    emit('change', page);
};
</script>

<template>
    <div class="mt-5 flex flex-col gap-4 px-4 pb-4 sm:flex-row sm:items-center sm:justify-between sm:px-6 sm:pb-6">
        <p class="text-sm text-gray-500">
            Menampilkan
            <span class="font-semibold text-gray-700">{{ mulaiData }}–{{ akhirData }}</span>
            dari
            <span class="font-semibold text-gray-700">{{ totalItems }}</span>
            data
        </p>

        <nav
            v-if="totalPages > 1"
            aria-label="Navigasi halaman"
            class="flex flex-wrap items-center justify-center gap-1.5"
        >
            <button
                type="button"
                @click="gantiHalaman(currentPage - 1)"
                :disabled="currentPage === 1 || loading"
                class="flex h-9 items-center gap-1 rounded-lg border border-gray-200 bg-white px-3 text-sm font-medium text-gray-600 transition hover:border-emerald-500 hover:bg-emerald-50 hover:text-emerald-700 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-gray-200 disabled:hover:bg-white disabled:hover:text-gray-600"
                aria-label="Halaman sebelumnya"
            >
                <i class="fas fa-chevron-left text-xs"></i>
                <span class="hidden sm:inline">Sebelumnya</span>
            </button>

            <template v-for="(page, index) in pageNumbers" :key="`${page}-${index}`">
                <span v-if="page === '...'" aria-hidden="true" class="px-1 text-gray-500">...</span>
                <button
                    v-else
                    type="button"
                    @click="gantiHalaman(page)"
                    :disabled="loading"
                    :aria-label="`Halaman ${page}`"
                    :aria-current="currentPage === page ? 'page' : undefined"
                    class="h-9 min-w-9 rounded-lg border px-3 text-sm font-semibold transition disabled:cursor-not-allowed"
                    :class="
                        currentPage === page
                            ? '!border-emerald-600 !bg-emerald-600 !text-white shadow-sm'
                            : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-500 hover:bg-emerald-50 hover:text-emerald-700'
                    "
                >
                    {{ page }}
                </button>
            </template>

            <button
                type="button"
                @click="gantiHalaman(currentPage + 1)"
                :disabled="currentPage >= totalPages || loading"
                class="flex h-9 items-center gap-1 rounded-lg border border-gray-200 bg-white px-3 text-sm font-medium text-gray-600 transition hover:border-emerald-500 hover:bg-emerald-50 hover:text-emerald-700 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-gray-200 disabled:hover:bg-white disabled:hover:text-gray-600"
                aria-label="Halaman berikutnya"
            >
                <span class="hidden sm:inline">Berikutnya</span>
                <i class="fas fa-chevron-right text-xs"></i>
            </button>
        </nav>
    </div>
</template>
