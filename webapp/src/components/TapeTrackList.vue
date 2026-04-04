<script setup lang="ts">
import type { TapeRsTrack } from '@/api/tapes';
import { Button, DataTable, Column } from 'primevue';

const model = defineModel<TapeRsTrack[]>()

function onUpdateTrackOrder(event: any) {
    model.value = event.value;
}

function onRemoveTrack(index: number) {
    model.value?.splice(index, 1);
}
</script>

<template>
    <DataTable :value="model" data-key="Id" @row-reorder="onUpdateTrackOrder" :show-headers="false">
        <Column row-reorder style="width: 3em" />

        <Column style="width: 3em">
            <template #body="{ index }">
                <Button icon="pi pi-trash" variant="text" severity="contrast" aria-label="Remove track"
                    @click="onRemoveTrack(index)" />
            </template>
        </Column>

        <Column>
            <template #body="{ data: { Artist, Title } }">
                <span v-if="Artist?.Name">{{ Artist?.Name ?? "" }}</span>
                <span v-if="Artist?.Name && Title">&ensp;-&ensp;</span>
                <span v-if="Title">{{ Title }}</span>
            </template>
        </Column>
    </DataTable>
</template>
