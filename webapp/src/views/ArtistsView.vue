<script setup lang="ts">
import artistApi, { type ArtistListRs } from '@/api/artists';
import { Button, Column, DataTable, IconField, InputIcon, InputText } from 'primevue';
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

// todo: paging
const limit = 9999;;

const isBusy = ref(false);
const query = ref("");

const artists = ref<ArtistListRs[]>([]);
const selectedArtists = ref<ArtistListRs[]>([]);

const visibleArtists = computed(() => {
    const result: ArtistListRs[] = [];

    const ids = new Set();
    for (const artist of artists.value) {
        ids.add(artist.Id);
    }

    for (const selectedArtist of selectedArtists.value) {
        if (!ids.has(selectedArtist.Id)) {
            result.push(selectedArtist);
        }
    }

    result.push(...artists.value);
    return result;
});

// todo: debounce
watch(query, async (query) => {
    try {
        isBusy.value = true;
        const rs = await artistApi.listArtists(query, limit, 0);
        artists.value = rs;
    } catch (e) {
        console.error("Failed to load artists", e);
    } finally {
        isBusy.value = false;
    }
}, { immediate: true });

async function merge() {
    try {
        isBusy.value = true;
        const rs = await artistApi.mergeArtists({
            Ids: selectedArtists.value.map(it => it.Id)
        });

        await router.push(`/artists/${rs.Id}`);
    } catch (e) {
        console.error("Failed to merge artists", e);
    } finally {
        isBusy.value = false;
    }
}
</script>

<template>
    <div>
        <div class="artist-list-toolbar">
            <IconField class="artist-list-toolbar-end">
                <InputIcon class="pi pi-search" />
                <InputText placeholder="Search..." v-model="query" />
            </IconField>
        </div>

        <DataTable class="artist-list-list" dataKey="Id" v-model:selection="selectedArtists" :value="visibleArtists"
            :loading="isBusy">
            <Column selection-mode="multiple">
                <template #header>
                    <Button :disabled="isBusy || selectedArtists.length < 2" label="Merge" @click="merge" />
                </template>
            </Column>
            <Column header="Artist">
                <template #body="{ data }">
                    <RouterLink :to="'/artists/' + data.Id">{{ data.Name }}</RouterLink>
                </template>
            </Column>
        </DataTable>
    </div>
</template>

<style>
.artist-list-toolbar {
    display: flex;
    flex-direction: column;
}

.artist-list-toolbar-end {
    align-self: flex-end;
}

.artist-list-list .p-datatable-header-cell .p-checkbox {
    display: none;
}
</style>
