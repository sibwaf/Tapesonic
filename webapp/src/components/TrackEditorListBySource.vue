<script setup lang="ts">
import type { ListSourceHierarchyRs, TrackRs } from "@/api";
import TrackEditor from "@/components/TrackEditor.vue";
import { type Editable } from "@/model/Editable";
import { computed } from "vue";

const props = defineProps<{
    modelValue: Editable<TrackRs>[],
    source: ListSourceHierarchyRs,
}>();

const tracksAreChanged = computed(() => props.modelValue.some(it => it.isEdited));

function resetTracks() {
    for (const track of props.modelValue) {
        track.reset();
    }
}
</script>

<template>
    <tr v-for="track, i in modelValue" :key="track.editedValue.Id">
        <td v-if="i == 0" :rowspan="modelValue.length">
            <div>
                <RouterLink :to="`/sources/${source.Id}`">{{ source.Title }}</RouterLink>
            </div>
            <div>
                <a :href="source.Url" target="_blank">Original URL</a>
                <button @click="resetTracks" :disabled="!tracksAreChanged">Reset</button>
            </div>
        </td>
        <TrackEditor :modelValue="track" />
    </tr>
</template>
