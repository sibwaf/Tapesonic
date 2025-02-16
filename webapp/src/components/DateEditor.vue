<script setup lang="ts">
import util from '@/util';
import { computed } from 'vue';

const model = defineModel<string | null>();

const date = computed({
    get(): string | null {
        const val = model.value;
        if (val == null) {
            return null;
        }

        return util.timestampToDate(val);
    },
    set(val: string | null) {
        if (val == null) {
            model.value = null;
            return;
        }

        const match = val.match(/^(\d{4}-\d{2}-\d{2})$/);
        if (match != null) {
            model.value = `${match[1]}T00:00:00Z`;
        }
    },
});

</script>

<template>
    <input type="date" v-model="date">
</template>
