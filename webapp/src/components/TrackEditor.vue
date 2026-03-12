<script setup lang="ts">
import type { SourceTrackRs } from "@/api/sources";
import { type Editable } from "@/model/Editable";
import util from "@/util";
import { computed } from "vue";
import ArtistSelector, { Artist } from "./ArtistSelector.vue";

const props = defineProps<{ modelValue: Editable<SourceTrackRs> }>();

const editedArtist = computed({
    get(): Artist | null {
        const trackValue = props.modelValue.editedValue;
        if (trackValue.Artist == null) {
            return null;
        }
        return new Artist(trackValue.Artist.Id, trackValue.Artist.Name);
    },
    set(value: Artist | null) {
        if (value == null) {
            props.modelValue.editedValue.Artist = null;
        } else {
            props.modelValue.editedValue.Artist = {
                Id: value.id,
                Name: value.name,
            };
        }
    }
});

const startText = computed({
    get(): string {
        return util.msToTimestamp(props.modelValue.editedValue.StartOffsetMs)
    },
    set(value: string) {
        props.modelValue.editedValue.StartOffsetMs = util.timestampToMs(value);
    }
});

const endText = computed({
    get(): string {
        return util.msToTimestamp(props.modelValue.editedValue.EndOffsetMs)
    },
    set(value: string) {
        props.modelValue.editedValue.EndOffsetMs = util.timestampToMs(value);
    }
});
</script>

<template>
    <td>
        <ArtistSelector v-model="editedArtist" />
    </td>
    <td> <input type="text" v-model="modelValue.editedValue.Title"> </td>
    <td> <input type="time" step="0.001" v-model="startText"> </td>
    <td> <input type="time" step="0.001" v-model="endText"> </td>
    <td> <button @click="modelValue.reset" :disabled="!modelValue.isEdited">Reset</button> </td>
</template>
