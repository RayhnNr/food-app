<script setup>
import { computed } from 'vue';
import Select2 from 'vue3-select2-component';

defineOptions({ inheritAttrs: false });

const props = defineProps({
    modelValue: {
        type: [String, Number, Array],
        default: '',
    },
    options: {
        type: Array,
        default: () => [],
    },
    placeholder: {
        type: String,
        default: 'Pilih...',
    },
    settings: {
        type: Object,
        default: () => ({}),
    },
    id: {
        type: String,
        default: '',
    },
    disabled: {
        type: Boolean,
        default: false,
    },
    required: {
        type: Boolean,
        default: false,
    },
});

defineEmits(['update:modelValue', 'select']);

const select2Settings = computed(() => ({
    width: '100%',
    dropdownCssClass: 'app-select2-dropdown',
    ...props.settings,
}));

const select2ModelValue = computed(() => {
    if (Array.isArray(props.modelValue)) return props.modelValue.map(String);
    return props.modelValue == null || props.modelValue === '' ? '' : String(props.modelValue);
});
</script>

<template>
    <div class="app-select2-field" v-bind="$attrs">
        <Select2
            :id="id"
            :model-value="select2ModelValue"
            :options="options"
            :placeholder="placeholder"
            :settings="select2Settings"
            :disabled="disabled"
            :required="required"
            @update:model-value="$emit('update:modelValue', $event)"
            @select="$emit('select', $event)"
        />
    </div>
</template>

<style>
.app-select2-field {
    width: 100%;
}

.app-select2-field .select2-container {
    width: 100% !important;
}

.app-select2-field .select2-container--default .select2-selection--single {
    position: relative;
    height: 42px;
    border: 1px solid #cbd5e1;
    border-radius: 6px;
    background: #fff;
    transition:
        border-color 150ms ease,
        box-shadow 150ms ease;
}

.app-select2-field .select2-container--default .select2-selection--single:hover {
    border-color: #94a3b8;
}

.app-select2-field .select2-container--focus .select2-selection--single,
.app-select2-field .select2-container--open .select2-selection--single {
    border-color: #059669;
    box-shadow: 0 0 0 3px rgb(5 150 105 / 12%);
    outline: none;
}

.app-select2-field
    .select2-container--default
    .select2-selection--single
    .select2-selection__rendered {
    padding-left: 12px;
    padding-right: 34px;
    color: #334155;
    font-size: 14px;
    line-height: 40px;
}

.app-select2-field
    .select2-container--default
    .select2-selection--single
    .select2-selection__placeholder {
    color: #94a3b8;
}

.app-select2-field
    .select2-container--default
    .select2-selection--single
    .select2-selection__arrow {
    top: 1px;
    right: 2px;
    height: 40px;
}

.app-select2-field
    .select2-container--default
    .select2-selection--single
    .select2-selection__clear {
    position: absolute;
    top: 50%;
    right: 28px;
    z-index: 1;
    margin: 0;
    color: #64748b;
    transform: translateY(-50%);
}

.app-select2-field .select2-container--disabled .select2-selection--single {
    background: #f1f5f9;
    cursor: not-allowed;
}

.app-select2-dropdown {
    overflow: hidden;
    border: 1px solid #e2e8f0 !important;
    border-radius: 6px;
    box-shadow: 0 8px 20px rgb(15 23 42 / 10%);
}

.app-select2-dropdown .select2-search__field {
    border-color: #cbd5e1 !important;
    border-radius: 4px;
    outline: none;
}

.app-select2-dropdown .select2-search__field:focus {
    border-color: #059669 !important;
    box-shadow: 0 0 0 2px rgb(5 150 105 / 12%);
}

.app-select2-dropdown .select2-results__option[aria-selected='true'] {
    background: #ecfdf5;
    color: #047857;
}

.app-select2-dropdown .select2-results__option--highlighted[aria-selected] {
    background: #059669;
    color: #fff;
}
</style>
