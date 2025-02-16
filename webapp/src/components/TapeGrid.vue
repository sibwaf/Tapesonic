<script setup lang="ts">
import { TapeType, type ListTape } from "@/api";
import { RouterLink } from "vue-router";
import Thumbnail from "@/components/Thumbnail.vue";

defineProps<{ modelValue: ListTape[] }>();
</script>

<template>
    <div class="tape-list">
        <RouterLink v-for="tape in modelValue" :key="tape.Id" :to="'/tapes/' + tape.Id" class="tape-list-item">
            <Thumbnail size="12em" :id="tape.ThumbnailId" />
            <div class="tape-name" :title="tape.Name">{{ tape.Name }}</div>
            <div class="tape-artist" :title="tape.Artist" v-if="tape.Type == TapeType.Album && tape.Artist">{{ tape.Artist }}</div>
        </RouterLink>
    </div>
</template>

<style>
.tape-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, 12em);
    justify-content: center;
    gap: 1em;
}

.tape-list-item {
    width: 12em;
}

.tape-name {
    font-size: 1.0em;

    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.tape-artist {
    font-size: 0.8em;

    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>
