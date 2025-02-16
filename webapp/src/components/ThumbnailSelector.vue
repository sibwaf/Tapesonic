<script setup lang="ts">
import Thumbnail from '@/components/Thumbnail.vue';

const props = defineProps<{
    thumbnailIds: string[],
    size: string,
}>();

const selectedThumbnailId = defineModel<string | null>({ required: true });

function onSelectThumbnail(id: string | null) {
    selectedThumbnailId.value = id;
}
</script>

<template>
    <div class="thumbnail-selector">
        <Thumbnail v-for="thumbnailId in [null, ...props.thumbnailIds]" :key="thumbnailId || 'null'" :size="size"
            :id="thumbnailId" :class="selectedThumbnailId == thumbnailId ? 'selected' : 'unselected'"
            @click="onSelectThumbnail(thumbnailId)" />
    </div>
</template>

<style lang="css" scoped>
.thumbnail-selector {
    display: flex;
    flex-direction: row;
    overflow-x: auto;
}

.unselected {
    border: 4px solid lightgray;
    padding: 4px;
}

.selected {
    border: 4px solid black;
    padding: 4px;
}
</style>
