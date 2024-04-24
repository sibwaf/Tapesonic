<script setup lang="ts">
import api, { type ImportQueueItem } from "@/api";
import { ref } from "vue";

const props = defineProps<{ item: ImportQueueItem }>();
const emit = defineEmits(["deleted"]);

enum State {
    STANDBY,
    DELETING,
    DELETING_ERROR,
}

const state = ref(State.STANDBY);

async function deleteItem() {
    try {
        state.value = State.DELETING;

        await api.deleteFromImportQueue(props.item.Id);
        emit("deleted");
    } catch (e) {
        state.value = State.DELETING_ERROR;
        console.error(e);
    }
}
</script>

<template>
    <div>
        <a :href="item.Url">{{ item.Url }}</a>
        <button :disabled="state == State.DELETING" @click="deleteItem">Delete</button>
        <span v-if="state == State.DELETING_ERROR">Failed to remove from queue</span>
    </div>
</template>
