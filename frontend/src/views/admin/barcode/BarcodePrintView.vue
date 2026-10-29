<script setup>
import { ref, computed, onMounted, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import QRCode from 'qrcode';
import JsBarcode from 'jsbarcode';
import api from '@/api';

const route = useRoute();

// Tinggi baris pita mengikuti template PHP:
// print_bat_no_urutz (60px) = pita 5x2, print_bat_no_urut (76px) = pita 5x2.5.
const JENIS = {
    'pita-5x2': { judul: 'Pita 5x2', layout: 'pita', tinggi: 60 },
    'pita-5x2.5': { judul: 'Pita 5x2.5', layout: 'pita', tinggi: 76 },
    bal: { judul: 'Bal', layout: 'bal' },
};

const jenisQuery = String(route.query.jenis || '');
const jenis = Object.hasOwn(JENIS, jenisQuery) ? JENIS[jenisQuery] : null;
const itemID = route.params.id;
const itemCode = route.query.item_code ? String(route.query.item_code) : '';

const labels = ref([]);
const loading = ref(true);
const error = ref('');

const svgKeUri = (svg) => `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;

// QR level H dengan margin 1, sama seperti pengaturan Endroid di PHP. Dibuat vektor supaya tajam saat dicetak.
const buatQR = async (kode) =>
    svgKeUri(await QRCode.toString(kode, { type: 'svg', errorCorrectionLevel: 'H', margin: 1 }));

// Code128 lebar 1 dan tinggi 30, sama seperti Picqer di PHP.
const buatBarcode = (kode) => {
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    JsBarcode(svg, kode, { format: 'CODE128', width: 1, height: 30, displayValue: false, margin: 0 });
    return svgKeUri(new XMLSerializer().serializeToString(svg));
};

const siapkan = (rows) =>
    Promise.all(
        rows.map(async (r) => {
            const external = r.item_type === 'external';
            // Label external tidak menampilkan nama partai.
            const [utama = '', kedua = ''] = external ? [] : (r.goods_partai_name || '').split(' ');

            return {
                ...r,
                partaiUtama: utama,
                partaiKedua: kedua,
                // Tahun hanya tampil di label external, sama seperti PHP.
                tahun: external ? r.year_code : '',
                nomor: Number.parseInt(r.serial, 10) || 0,
                qr: await buatQR(r.item_code),
                barcode: external && jenis.layout === 'bal' ? buatBarcode(r.item_code) : '',
            };
        }),
    );

// Pita dicetak dua label per baris.
const pasangan = computed(() => {
    const hasil = [];
    for (let i = 0; i < labels.value.length; i += 2) {
        hasil.push([labels.value[i], labels.value[i + 1]]);
    }
    return hasil;
});

const cetak = async () => {
    await nextTick();
    await Promise.all([...document.images].map((img) => img.decode().catch(() => {})));
    window.print();
};

const tutup = () => window.close();

onMounted(async () => {
    if (!jenis) {
        error.value = 'Jenis cetak tidak dikenal.';
        loading.value = false;
        return;
    }

    try {
        const res = await api.get(`/barcode/${itemID}/print`, {
            params: itemCode ? { item_code: itemCode } : {},
        });
        labels.value = await siapkan(res.data.data);
    } catch (e) {
        error.value = e.response?.data?.message || 'Gagal menyiapkan label cetak';
        console.error(e);
        return;
    } finally {
        loading.value = false;
    }

    await cetak();
});
</script>

<template>
    <div class="cetak">
        <header class="no-print mb-4 flex items-center justify-between rounded-md bg-gray-100 px-4 py-3">
            <p class="text-sm text-gray-700">
                <strong>Cetak {{ jenis?.judul }}</strong>
                <span v-if="!loading && !error"> ({{ labels.length }} label)</span>
            </p>
            <div class="flex gap-2">
                <button
                    type="button"
                    :disabled="loading || !!error || labels.length === 0"
                    class="rounded-md bg-emerald-600 px-3 py-1 text-sm text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50"
                    @click="cetak"
                >
                    Cetak ulang
                </button>
                <button
                    type="button"
                    class="rounded-md border border-gray-300 bg-white px-3 py-1 text-sm text-gray-700 hover:bg-gray-50"
                    @click="tutup"
                >
                    Tutup
                </button>
            </div>
        </header>

        <p v-if="loading" class="no-print text-sm text-gray-500">Menyiapkan label...</p>
        <p v-else-if="error" class="no-print text-sm text-red-600">{{ error }}</p>

        <!-- Pita: dua label per baris -->
        <template v-else-if="jenis.layout === 'pita'">
            <template v-for="(pasang, idx) in pasangan" :key="idx">
                <table class="tabel-pita garis-putih" cellspacing="0" cellpadding="0">
                    <tbody>
                        <tr :style="{ height: jenis.tinggi + 'px' }">
                            <template v-for="(label, k) in pasang" :key="k">
                                <td v-if="k === 1" class="spasi">&nbsp;</td>
                                <td class="sel">
                                    <table
                                        v-if="label"
                                        class="tabel-dalam garis-putih"
                                        cellspacing="0"
                                        cellpadding="0"
                                    >
                                        <tbody>
                                            <tr>
                                                <td align="center">
                                                    <img class="qr" :src="label.qr" alt="" />
                                                </td>
                                                <td align="right" class="teks">
                                                    <b class="baris">
                                                        {{ label.tahun }}
                                                        <span class="utama">{{ label.partaiUtama }}</span>
                                                        {{ label.partaiKedua }}
                                                        {{ label.company_code }}</b
                                                    ><br />
                                                    <b class="urut">{{ label.serial }}</b>
                                                </td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </td>
                            </template>
                        </tr>
                    </tbody>
                </table>
                <div class="jeda">&nbsp;</div>
            </template>
        </template>

        <!-- Bal: satu label per baris. Internal dua sisi QR, external satu sisi QR dan satu sisi Code128 -->
        <template v-else>
            <div v-for="label in labels" :key="label.item_code" class="blok">
                <table class="tabel-bal garis-hitam" cellspacing="0" cellpadding="0">
                    <tbody>
                        <tr class="baris-bal">
                            <template
                                v-for="(sisi, k) in ['miring', label.item_type === 'external' ? 'kode' : 'mirings']"
                                :key="k"
                            >
                                <td v-if="k === 1" class="tengah">&nbsp;</td>

                                <td v-if="sisi !== 'kode'" class="sisi">
                                    <div :class="['isi-sisi', sisi]">
                                        <b class="teks15">
                                            {{ label.tahun }}
                                            {{ label.company_code }}
                                            <span class="utama">{{ label.partaiUtama }}</span>
                                            {{ label.partaiKedua }}</b
                                        >
                                        <img class="qr miring" :src="label.qr" alt="" />
                                        <b class="teks15">{{ label.nomor }}</b>
                                    </div>
                                </td>

                                <td v-else class="sisi">
                                    <div class="kode-bar">
                                        <img class="barcode" :src="label.barcode" alt="" />
                                    </div>
                                    <b class="teks13">{{ label.tahun }} {{ label.item_code }}</b>
                                </td>
                            </template>
                        </tr>
                    </tbody>
                </table>
                <div class="jeda">&nbsp;</div>
            </div>
        </template>
    </div>
</template>

<style scoped>
.cetak {
    font-family: Tahoma, sans-serif;
    /* Margin bawaan body browser (8px) yang ada di halaman PHP, tapi dihapus oleh reset Tailwind. */
    margin: 8px;
}

.garis-putih,
.garis-hitam {
    border-radius: 6px;
    padding: 0;
}

.garis-putih {
    border: 1px solid white;
}

.garis-hitam {
    border: 1px solid black;
}

.tabel-pita,
.tabel-bal {
    border-collapse: separate;
    border-spacing: 0;
    width: 363px;
    padding: 0;
}

.tabel-dalam {
    border-collapse: separate;
    border-spacing: 0;
    width: 160px;
    padding: 0;
}

.sel {
    text-align: center;
    width: 160px;
}

.spasi {
    text-align: center;
    width: 28px;
}

.teks {
    padding-right: 5px;
}

.baris {
    font-size: 15px;
}

.utama {
    font-size: 20px;
    font-weight: bold;
}

.urut {
    font-size: 22px;
}

.qr {
    flex: none;
    width: 60px;
    height: 60px;
}

.jeda {
    height: 8px;
}

/* Satu label bal beserta jarak 8px di bawahnya. Label tidak boleh terbelah di batas halaman/stiker:
   kalau tidak muat, seluruh label pindah ke halaman berikutnya.
   Dua aturan pertama meniadakan reset Tailwind (line-height 1.5, border-box) di area label bal,
   supaya hitungan tinggi sama dengan halaman PHP yang tidak memakai reset itu. */
.blok,
.blok * {
    box-sizing: content-box;
    line-height: normal;
}

.blok {
    break-inside: avoid;
    page-break-inside: avoid;
}

/* Sama dengan PHP: tinggi baris 95px hanya minimum, isinya mengalir sehingga baris bisa lebih tinggi. */
.baris-bal {
    height: 95px;
}

/* Ukuran tabel bal disamakan dengan hasil ukur halaman PHP di DevTools (363 x 106):
   lebar sudah termasuk garis tepi (PHP berjalan di quirks mode, tabel memakai border-box). */
.blok .tabel-bal {
    box-sizing: border-box;
    height: 106px;
}

.sisi {
    text-align: center;
    width: 130px;
}

/* Isi mengalir seperti di PHP: teks, QR, nomor. QR dibuat block supaya tidak ada celah di bawah
   gambar, sama seperti halaman PHP yang berjalan di quirks mode (tanpa doctype). */
.isi-sisi {
    width: 130px;
}

.isi-sisi .qr {
    display: block;
    margin: 0 auto;
}

.tengah {
    text-align: center;
    width: 100px;
}

.miring {
    transform: rotate(90deg);
}

.mirings {
    transform: rotate(270deg);
}

.teks15 {
    font-size: 15px;
}

.teks13 {
    font-size: 13px;
}

.kode-bar {
    text-align: center;
}

.barcode {
    display: block;
    margin: 0 auto;
    max-width: 100%;
    height: 30px;
}

@media print {
    .no-print {
        display: none;
    }
}
</style>