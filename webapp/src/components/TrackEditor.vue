<script setup lang="ts">
import type { TrackRs } from "@/api";
import { type Editable } from "@/model/Editable";
import util from "@/util";
import { computed } from "vue";

const props = defineProps<{ modelValue: Editable<TrackRs> }>();

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
    <td> <input type="text" v-model="modelValue.editedValue.Artist"> </td>
    <td> <input type="text" v-model="modelValue.editedValue.Title"> </td>
    <td> <input type="time" step="0.001" v-model="startText"> </td>
    <td> <input type="time" step="0.001" v-model="endText"> </td>
    <td> <button @click="modelValue.reset" :disabled="!modelValue.isEdited">Reset</button> </td>
</template>
