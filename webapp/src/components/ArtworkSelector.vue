<script setup lang="ts">
import Artwork from '@/components/Artwork.vue';

const props = defineProps<{
    artworkIds: string[],
    size: string,
}>();

const selectedArtworkId = defineModel<string | null>({ required: true });

function onSelectArtwork(id: string | null) {
    selectedArtworkId.value = id;
}
</script>

<template>
    <div class="artwork-selector">
        <Artwork v-for="artworkId in [null, ...props.artworkIds]" :key="artworkId || 'null'" :size="size"
            :id="artworkId" :class="selectedArtworkId == artworkId ? 'selected' : 'unselected'"
            @click="onSelectArtwork(artworkId)" />
    </div>
</template>

<style lang="css">
.artwork-selector {
    display: flex;
    flex-direction: row;
    overflow-x: auto;
}

.artwork-selector > .unselected {
    border: 4px solid lightgray;
    padding: 4px;
}

.artwork-selector > .selected {
    border: 4px solid black;
    padding: 4px;
}
</style>
