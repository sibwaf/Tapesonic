<script setup lang="ts">
const props = defineProps<{
    id: string | null,
    size: string,
}>();
</script>

<template>
    <span class="thumbnail-container" :style="`width: ${size}; height: ${size};`">
        <img v-if="id" class="thumbnail" :src="`/media/thumbnails/${props.id}`" />
        <span v-else class="thumbnail">No thumbnail</span>
    </span>
</template>

<style>
.thumbnail-container {
    display: flex;
    justify-content: center;
    align-items: center;

    box-sizing: border-box;

    flex-shrink: 0;
    flex-grow: 0;
}

.thumbnail {
    max-height: 100%;
    max-width: 100%;
}
</style>
