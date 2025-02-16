<script setup lang="ts">
import { type EditableTrack } from "@/model/EditableTrack";
import { computed } from "vue";
import TrackEditorListBySource from "@/components/TrackEditorListBySource.vue";
import type { ListSourceHierarchyRs } from "@/api";

const props = defineProps<{
    modelValue: EditableTrack[],
    orderedSourceHierarchy: ListSourceHierarchyRs[],
}>();

interface SourceAndTracks {
    source: ListSourceHierarchyRs;
    tracks: EditableTrack[];
}

const tracksBySource = computed(() => {
    const tracksBySourceId = new Map<string, EditableTrack[]>();

    for (const track of props.modelValue) {
        const sourceId = track.editedValue.SourceId;

        let tracks = tracksBySourceId.get(sourceId);
        if (tracks == null) {
            tracks = [];
            tracksBySourceId.set(sourceId, tracks);
        }

        tracks.push(track);
    }

    for (const [_, tracks] of tracksBySourceId.entries()) {
        tracks.sort((x, y) => x.editedValue.StartOffsetMs - y.editedValue.StartOffsetMs);
    }

    const result: SourceAndTracks[] = [];
    for (const source of props.orderedSourceHierarchy) {
        const tracks = tracksBySourceId.get(source.Id);
        if (tracks != null) {
            result.push({ source: source, tracks: tracks });
        }
    }
    return result;
});
</script>

<template>
    <table border="1" frame="void" rules="rows">
        <thead>
            <tr>
                <th> Source </th>
                <th> Artist </th>
                <th> Title </th>
                <th> Start offset </th>
                <th> End offset </th>
                <th></th>
            </tr>
        </thead>
        <tbody>
            <TrackEditorListBySource v-for="item in tracksBySource" :key="item.source.Id" :source="item.source"
                :modelValue="item.tracks" />
        </tbody>
    </table>
</template>
