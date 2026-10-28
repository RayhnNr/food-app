import { ref } from 'vue';
import api from '@/api';
import Swal from 'sweetalert2';

const Toast = Swal.mixin({
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
});

const toNumber = (value) => {
    if (value === '' || value === null) return null;

    const number = Number(value);
    return Number.isFinite(number) ? number : null;
};

const emptyForm = () => ({ min: '', max: '', nominal: '' });

export function useRentang({ partaiId, resource, fields, idKey, label }) {
    const rows = ref([]);
    const loading = ref(false);
    const saving = ref(false);
    const form = ref(emptyForm());
    const errors = ref({});
    const formError = ref('');
    const loadError = ref('');

    const url = () => `/goods-partai/${partaiId()}/${resource}`;

    const load = async () => {
        loading.value = true;
        loadError.value = '';

        try {
            const res = await api.get(url());
            rows.value = res.data.data;
        } catch (e) {
            loadError.value = e.response?.data?.message || `Gagal mengambil data ${label}.`;
            console.error(e);
        } finally {
            loading.value = false;
        }
    };

    const save = async () => {
        if (saving.value) return;

        saving.value = true;
        errors.value = {};
        formError.value = '';

        const payload = {
            [fields.min]: toNumber(form.value.min),
            [fields.max]: toNumber(form.value.max),
            [fields.nominal]: toNumber(form.value.nominal),
        };

        try {
            await api.post(url(), payload);
            form.value = emptyForm();
            await load();
            Toast.fire({ icon: 'success', title: `Berhasil menambahkan data ${label}` });
        } catch (e) {
            const data = e.response?.data;
            const raw = data?.errors || {};

            errors.value = {
                min: raw[fields.min],
                max: raw[fields.max],
                nominal: raw[fields.nominal],
            };
            formError.value = data?.errors ? '' : data?.message || `Gagal menyimpan data ${label}.`;
            console.error(e);
        } finally {
            saving.value = false;
        }
    };

    const remove = async (row) => {
        const result = await Swal.fire({
            title: 'Yakin hapus data ini?',
            text: `${row[fields.min]} - ${row[fields.max]}`,
            icon: 'warning',
            showCancelButton: true,
            confirmButtonText: 'Ya, hapus',
            cancelButtonText: 'Batal',
        });

        if (!result.isConfirmed) return;

        try {
            await api.delete(`${url()}/${row[idKey]}`);
            await load();
            Toast.fire({ icon: 'success', title: `Berhasil menghapus data ${label}` });
        } catch (e) {
            Swal.fire({
                icon: 'error',
                title: 'Gagal menghapus',
                text: e.response?.data?.message || `Gagal menghapus data ${label}`,
            });
            console.error(e);
        }
    };

    return { rows, loading, saving, form, errors, formError, loadError, load, save, remove };
}