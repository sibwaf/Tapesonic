<script setup lang="ts">
import { type TapeTrack } from "@/api";
import TapeTrackEditor from "@/components/TapeTrackEditor.vue";

defineProps<{ modelValue: TapeTrack[] }>();
</script>

<template>
    <table>
        <thead>
            <th>Artist</th>
            <th>Title</th>
            <th>Start offset</th>
            <th>End offset</th>
        </thead>
        <tbody>
            <TapeTrackEditor v-for="track in modelValue" :key="track.TapeId + '/' + track.TapeTrackIndex"
                :model-value="track" />
        </tbody>
    </table>
</template>
