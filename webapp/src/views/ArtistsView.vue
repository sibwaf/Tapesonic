<script setup lang="ts">
import artistApi, { type ArtistListRs } from '@/api/artists';
import { Button, Column, DataTable, IconField, InputIcon, InputText } from 'primevue';
import { ref, watch } from 'vue';

// todo: paging
const limit = 9999;;

const isBusy = ref(false);
const query = ref("");

const artists = ref<ArtistListRs[]>([]);

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

</script>

<template>
    <div>
        <div class="toolbar">
            <IconField class="toolbar-end">
                <InputIcon class="pi pi-search" />
                <InputText placeholder="Search..." v-model="query" />
            </IconField>
        </div>

        <DataTable :value="artists" :loading="isBusy">
            <Column header="Artist">
                <template #body="{ data }">
                    <RouterLink :to="'/artists/' + data.Id">{{ data.Name }}</RouterLink>
                </template>
            </Column>
        </DataTable>
    </div>
</template>

<style scoped>
.toolbar {
    display: flex;
    flex-direction: column;
}

.toolbar-end {
    align-self: flex-end;
}
</style>
