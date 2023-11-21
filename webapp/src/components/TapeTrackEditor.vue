<script setup lang="ts">
import { type TapeTrack } from "@/api";
import util from "@/util";
import { computed } from "vue";

const props = defineProps<{ modelValue: TapeTrack }>();

function offsetDiffToText(diff: number): string {
    if (diff == 0) {
        return "";
    } else if (diff > 0) {
        return `+${diff}`;
    } else {
        return `${diff}`;
    }
}

const startOffsetDiff = computed(() => props.modelValue.StartOffsetMs - props.modelValue.RawStartOffsetMs);
const startOffsetDiffText = computed(() => offsetDiffToText(startOffsetDiff.value));
const startText = computed(() => util.msToTimestamp(props.modelValue.StartOffsetMs));

const endOffsetDiff = computed(() => props.modelValue.EndOffsetMs - props.modelValue.RawEndOffsetMs);
const endOffsetDiffText = computed(() => offsetDiffToText(endOffsetDiff.value));
const endText = computed(() => util.msToTimestamp(props.modelValue.EndOffsetMs))

</script>

<template>
    <tr>
        <td>
            <input type="text" v-model="modelValue.Artist">
        </td>
        <td>
            <input type="text" v-model="modelValue.Title">
        </td>
        <td>
            <div>
                <input type="text" v-model.number="modelValue.StartOffsetMs">
            </div>
            <div>
                <span v-if="startOffsetDiffText">{{ startOffsetDiffText }} ms </span>
                <span>{{ startText }}</span>
            </div>
        </td>
        <td>
            <div>
                <input type="text" v-model.number="modelValue.EndOffsetMs">
            </div>
            <div>
                <span v-if="endOffsetDiffText">{{ endOffsetDiffText }} ms </span>
                <span>{{ endText }}</span>
            </div>
        </td>
    </tr>
</template>
