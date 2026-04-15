<script setup lang="ts">
const props = defineProps<{
    id: string | null,
    size: string,
}>();
</script>

<template>
    <span class="artwork-container" :style="`width: ${size}; height: ${size};`">
        <img v-if="id" class="artwork" :src="`/media/artworks/${props.id}`" />
        <span v-else class="artwork">No artwork</span>
    </span>
</template>

<style>
.artwork-container {
    display: flex;
    justify-content: center;
    align-items: center;

    box-sizing: border-box;

    flex-shrink: 0;
    flex-grow: 0;
}

.artwork {
    max-height: 100%;
    max-width: 100%;
}
</style>
